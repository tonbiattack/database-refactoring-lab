package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/tonbiattack/database-refactoring-lab/internal/db"
	"github.com/tonbiattack/database-refactoring-lab/internal/order"
)

func main() {
	batchSize := flag.Int("batch-size", 100, "1回に更新する最大件数")
	fromID := flag.Int64("from-id", 0, "対象IDの下限")
	toID := flag.Int64("to-id", 0, "対象IDの上限。0は上限なし")
	flag.Parse()
	if *batchSize < 1 {
		log.Fatal("batch-size must be positive")
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

	updated, err := backfill(context.Background(), connection, *batchSize, *fromID, *toID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated: %d\n", updated)
}

func backfill(ctx context.Context, db *sql.DB, batchSize int, fromID, toID int64) (int, error) {
	total := 0
	for {
		query := "SELECT id, status_code FROM orders WHERE status IS NULL AND status_code IN (1, 2, 3) AND id >= ?"
		args := []any{fromID}
		if toID > 0 {
			query += " AND id <= ?"
			args = append(args, toID)
		}
		query += " ORDER BY id LIMIT ?"
		args = append(args, batchSize)

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		type row struct {
			id   int64
			code int
		}
		var batch []row
		for rows.Next() {
			var item row
			if err := rows.Scan(&item.id, &item.code); err != nil {
				rows.Close()
				return 0, err
			}
			batch = append(batch, item)
		}
		if err := rows.Close(); err != nil {
			return 0, err
		}
		if len(batch) == 0 {
			return total, nil
		}

		transaction, err := db.BeginTx(ctx, nil)
		if err != nil {
			return 0, err
		}
		for _, item := range batch {
			status, _ := order.StatusFromLegacy(&item.code)
			result, err := transaction.ExecContext(ctx, "UPDATE orders SET status = ?, updated_at = updated_at WHERE id = ? AND status IS NULL", status.Code, item.id)
			if err != nil {
				transaction.Rollback()
				return 0, err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				transaction.Rollback()
				return 0, err
			}
			total += int(affected)
		}
		if err := transaction.Commit(); err != nil {
			return 0, err
		}
	}
}
