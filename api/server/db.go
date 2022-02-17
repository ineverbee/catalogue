package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	attributes = map[string]bool{
		"name":            true,
		"color":           true,
		"tail_length":     true,
		"whiskers_length": true,
	}
	orders = map[string]bool{
		"asc":  true,
		"desc": true,
	}
)

// {"name": "Tihon", "color": "red & white", "tail_length": 15, "whiskers_length": 12}
type Cat struct {
	Name           string `json:"name"`
	Color          string `json:"color"`
	TailLength     int    `json:"tail_length"`
	WhiskersLength int    `json:"whiskers_length"`
}

type Filters struct {
	Attribute string
	Order     string
	Offset    int
	Limit     int
}

func NewDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(ctx, connString)

	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (s *ApiServer) GetAllCats(ftrs *Filters) ([]Cat, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var l int
	err := s.DB.QueryRow(ctx, "select count(*) from cats").Scan(&l)
	if err != nil {
		return nil, err
	}
	if ftrs.Offset >= l {
		return nil, &StatusError{http.StatusBadRequest, fmt.Errorf("error: offset (%d) equal or more than rows (%d) in the table", ftrs.Offset, l)}
	}

	q := "select * from cats"
	if ftrs.Attribute != "" {
		q += fmt.Sprintf(" order by %s", ftrs.Attribute)
	}
	if ftrs.Order != "" {
		q += fmt.Sprintf(" %s", ftrs.Order)
	}
	if ftrs.Offset != 0 {
		q += fmt.Sprintf(" offset %d", ftrs.Offset)
	}
	if ftrs.Limit != 0 {
		q += fmt.Sprintf(" limit %d", ftrs.Limit)
	}

	rows, err := s.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if ftrs.Offset != 0 {
		l -= ftrs.Offset
		if ftrs.Limit != 0 && ftrs.Limit < l {
			l = ftrs.Limit
		}
	} else if ftrs.Limit != 0 && ftrs.Limit < l {
		l = ftrs.Limit
	}
	cats := make([]Cat, l)
	for i := 0; rows.Next(); i++ {
		rows.Scan(&cats[i].Name, &cats[i].Color, &cats[i].TailLength, &cats[i].WhiskersLength)
	}
	return cats, nil
}

func (s *ApiServer) Set(c *Cat) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var temp string
	s.DB.QueryRow(ctx, "select name from cats where name=$1", c.Name).Scan(&temp)
	if temp != "" {
		return &StatusError{http.StatusBadRequest, fmt.Errorf("error: cat with name %s already exists", c.Name)}
	}
	_, err := s.DB.Exec(ctx, "insert into cats(name, color, tail_length, whiskers_length) values ($1,$2,$3,$4)",
		c.Name,
		c.Color,
		c.TailLength,
		c.WhiskersLength,
	)
	if err != nil {
		return &StatusError{http.StatusBadRequest, err}
	}
	return nil
}
