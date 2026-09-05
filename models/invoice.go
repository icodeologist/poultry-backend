package models

import "time"

// Invoice is a report built from an order and its payments. It is not stored in
// a separate table because the order and payment records are the source of truth.
type Invoice struct {
	InvoiceID     string    `json:"invoice_id"`
	OrderID       uint      `json:"order_id"`
	CustomerID    uint      `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerType  string    `json:"customer_type"`
	Time          time.Time `json:"time"`
	Amount        float64   `json:"amount"`
	AmountStatus  string    `json:"amount_status"`
	ItemsSold     float64   `json:"items_sold"`
	PaymentMethod string    `json:"payment_method,omitempty"`
}
