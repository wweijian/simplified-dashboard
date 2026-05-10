package finance

import (
	"database/sql"
	"time"
)

type Category struct {
	ID   int64
	Name string
	Type string
}

type Transaction struct {
	ID           int64
	Date         string
	Amount       float64
	CategoryID   sql.NullInt64
	CategoryName sql.NullString
	CategoryType sql.NullString
	Description  sql.NullString
	CreatedAt    string
}

type CreateTransactionInput struct {
	Date        string
	Amount      float64
	CategoryID  sql.NullInt64
	Description sql.NullString
}

type UpdateTransactionInput = CreateTransactionInput

type MonthlySummary struct {
	Month    time.Time
	Income   float64
	Expenses float64
	Net      float64
}

type CategorySpend struct {
	Name   string
	Amount float64
}

type MonthSpend struct {
	Month  time.Time
	Amount float64
}

type CategoryFilter struct {
	ID   sql.NullInt64
	Name string
}
