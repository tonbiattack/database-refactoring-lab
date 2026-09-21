package order

import "time"

type Status struct {
	LegacyCode int
	Code       string
}

type Order struct {
	ID           int64
	StatusCode   *int
	Status       *string
	CustomerNote *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OrderView struct {
	ID           int64   `json:"id"`
	Status       string  `json:"status"`
	CustomerNote *string `json:"customerNote"`
}
