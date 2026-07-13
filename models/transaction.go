package models

import "time"

// keeps every transaction

type Transaction struct {
	ID           int
	CustomerID   int
	Product      int
	QuantitySold int
	CashPaid     float64
	CreditGiven  float64
	CreatedAt    time.Time
}
