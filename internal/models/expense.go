package models

import "time"

// Expense — одна трата.
// CategoryID == 0 означает «без категории».
type Expense struct {
	ID           int64
	Amount       float64
	Description  string
	Date         time.Time
	CategoryID   int64
	CategoryName string
}
