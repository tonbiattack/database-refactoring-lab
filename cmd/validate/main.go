package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tonbiattack/database-refactoring-lab/internal/db"
	"github.com/tonbiattack/database-refactoring-lab/internal/order"
)

func main() {
	config, err := db.ConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	connection, err := db.Open(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	summary, err := order.Validate(context.Background(), connection)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total:          %d\nMigrated:       %d\nInconsistent:   %d\nUnknown:        %d\n", summary.Total, summary.Migrated, summary.Inconsistent, summary.Unknown)
}
