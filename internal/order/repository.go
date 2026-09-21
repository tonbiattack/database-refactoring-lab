package order

import (
	"context"
	"database/sql"
)

type Repository interface {
	FindByID(context.Context, int64) (Order, error)
	UpdateStatus(context.Context, int64, Status, WriteMode) error
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindByID(ctx context.Context, id int64) (Order, error) {
	var order Order
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status_code, status, customer_note, created_at, updated_at
		FROM orders
		WHERE id = ?
	`, id).Scan(
		&order.ID,
		&order.StatusCode,
		&order.Status,
		&order.CustomerNote,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	return order, err
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id int64, status Status, mode WriteMode) error {
	var result sql.Result
	var err error
	if mode == WriteModeLegacy {
		result, err = r.db.ExecContext(ctx, `
			UPDATE orders
			SET status_code = ?, updated_at = NOW()
			WHERE id = ?
		`, status.LegacyCode, id)
	} else {
		result, err = r.db.ExecContext(ctx, `
			UPDATE orders
			SET status_code = ?, status = ?, updated_at = NOW()
			WHERE id = ?
		`, status.LegacyCode, status.Code, id)
	}
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
