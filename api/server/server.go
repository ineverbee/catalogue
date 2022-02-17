package server

import (
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Server interface {
	GetAllCats(*Filters) ([]Cat, error)
	Set(*Cat) error
}

type ApiServer struct {
	Router *http.ServeMux
	DB     *pgxpool.Pool
}

var s Server

func NewServer(r *http.ServeMux, p *pgxpool.Pool) {
	s = &ApiServer{Router: r, DB: p}
}
