package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ineverbee/catalogue/api/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("api/.env")
	if err != nil {
		log.Fatal(err)
	}

	err = Start()
	if err != nil {
		log.Fatal(err)
	}
}

func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?connect_timeout=10",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := server.NewDB(ctx, connString)

	if err != nil {
		return err
	}

	r := http.NewServeMux()
	server.NewServer(r, db)
	server.ConfigureHandlers(r)

	log.Print("Listening and serving on http://localhost:8080/")
	err = http.ListenAndServe(":8080", r)
	return err
}
