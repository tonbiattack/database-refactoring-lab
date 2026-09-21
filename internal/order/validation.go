package order

import (
	"context"
	"database/sql"
)

type ValidationClass string

const (
	ValidationMigrated     ValidationClass = "Migrated"
	ValidationInconsistent ValidationClass = "Inconsistent"
	ValidationUnknown      ValidationClass = "Unknown"
)

type ValidationSummary struct {
	Total        int
	Migrated     int
	Inconsistent int
	Unknown      int
}

func Classify(statusCode *int, status *string) ValidationClass {
	expected, known := StatusFromLegacy(statusCode)
	if !known {
		if status == nil {
			return ValidationUnknown
		}
		return ValidationInconsistent
	}
	if status != nil && *status == expected.Code {
		return ValidationMigrated
	}
	return ValidationInconsistent
}

func Validate(ctx context.Context, db *sql.DB) (ValidationSummary, error) {
	rows, err := db.QueryContext(ctx, "SELECT status_code, status FROM orders")
	if err != nil {
		return ValidationSummary{}, err
	}
	defer rows.Close()

	var summary ValidationSummary
	for rows.Next() {
		var statusCode *int
		var status *string
		if err := rows.Scan(&statusCode, &status); err != nil {
			return ValidationSummary{}, err
		}
		summary.Total++
		switch Classify(statusCode, status) {
		case ValidationMigrated:
			summary.Migrated++
		case ValidationInconsistent:
			summary.Inconsistent++
		case ValidationUnknown:
			summary.Unknown++
		}
	}
	return summary, rows.Err()
}
