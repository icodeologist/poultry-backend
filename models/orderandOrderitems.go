package models

import "time"

type Order struct {
	ID             uint
	CustomerID     *uint // fk -> User
	Customer       Customer
	TimeStamp      time.Time
	TotalAmount    float64
	PaymentMethod  PaymentMethod
	PaymentBalance float64
	Status         string
	OrderItems     []OrderItem // one  to many with orderItem
}

type OrderItem struct {
	ID          uint
	ProductId   uint
	OrderId     uint
	Quantity    float64
	PriceAtSale float64
}

type PaymentMethod string

const (
	UPI  PaymentMethod = "UPI"
	CASH PaymentMethod = "CASH"
	CARD PaymentMethod = "CARD"
)
