package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

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

	var color string
	var count int64
	rows, err := conn.Query(ctx, "select color, count(*) from cats group by color")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&color, &count)
		_, err := conn.Exec(ctx, "insert into cat_colors_info(color, count) values ($1, $2)", color, count)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Insertion failed: %v\n", err)
			os.Exit(1)
		}
	}

	new_rows, err := conn.Query(ctx, "select * from cat_colors_info")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	defer new_rows.Close()

	for new_rows.Next() {
		new_rows.Scan(&color, &count)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Insertion failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(color, count)
	}
}
