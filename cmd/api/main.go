package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/tonbiattack/database-refactoring-lab/internal/db"
	"github.com/tonbiattack/database-refactoring-lab/internal/order"
)

func main() {
	readMode, err := order.ParseReadMode(os.Getenv("READ_MODE"))
	if err != nil {
		log.Fatal(err)
	}
	writeMode, err := order.ParseWriteMode(os.Getenv("WRITE_MODE"))
	if err != nil {
		log.Fatal(err)
	}
	config, err := db.ConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	connection, err := db.Open(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	service := order.NewService(order.NewSQLRepository(connection), readMode, writeMode)
	log.Fatal(http.ListenAndServe(":"+port, order.NewHandler(service)))
}
