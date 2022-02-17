package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/jackc/pgx/v4/pgxpool"
)

func query(ctx context.Context, conn *pgxpool.Pool, q string, arr []int) {
	rows, err := conn.Query(ctx, q)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		rows.Scan(&arr[i])
	}
}

func mean(numbers []int) float64 {
	total := 0.0

	for _, v := range numbers {
		total += float64(v)
	}

	return total / float64(len(numbers))
}

func median(numbers []int) float64 {
	sort.Ints(numbers) // sort the numbers
	n := len(numbers)
	mNumber := n / 2

	if n%2 != 0 {
		return float64(numbers[mNumber])
	}

	return float64((numbers[mNumber-1] + numbers[mNumber]) / 2)
}

func mode(numbers []int) []int {
	m := make(map[int]int)

	for _, v := range numbers {
		m[v]++
	}
	res := []int{}
	var max int
	for k, v := range m {
		if v == max {
			res = append(res, k)
		} else if v > max {
			res = []int{k}
			max = v
		}
	}
	return res
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urlExample := "postgres://wg_forge:42a@localhost:5432/wg_forge_db"
	conn, err := pgxpool.Connect(ctx, urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	var l int
	err = conn.QueryRow(ctx, "select count(*) from cats").Scan(&l)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	tail_arr := make([]int, l)
	query(ctx, conn, "select tail_length from cats", tail_arr)

	tail_mean := mean(tail_arr)
	tail_median := median(tail_arr)
	tail_mode := mode(tail_arr)

	whiskers_arr := make([]int, l)
	query(ctx, conn, "select whiskers_length from cats", whiskers_arr)

	whiskers_mean := mean(whiskers_arr)
	whiskers_median := median(whiskers_arr)
	whiskers_mode := mode(whiskers_arr)

	fmt.Printf("%.2f %.2f %v\n%.2f %.2f %v\n",
		tail_mean,
		tail_median,
		tail_mode,
		whiskers_mean,
		whiskers_median,
		whiskers_mode,
	)

	_, err = conn.Exec(ctx, "insert into cats_stat(tail_length_mean,tail_length_median,tail_length_mode,whiskers_length_mean,whiskers_length_median,whiskers_length_mode) values ($1,$2,$3,$4,$5,$6)",
		tail_mean,
		tail_median,
		tail_mode,
		whiskers_mean,
		whiskers_median,
		whiskers_mode,
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Insertion failed: %v\n", err)
		os.Exit(1)
	}
}
