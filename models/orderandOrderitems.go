package models

import "time"

type Order struct {
	ID              uint
	UserID          uint // fk -> User
	TimeStamp       time.Time
	TotalAmount     float64
	TransactionType TransactionType
	PaymentMethod   PaymentMethod
	PaymentBalance  float64
	Status          string
	OrderItems      []OrderItem // one  to many with orderItem
}

type OrderItem struct {
	ID          uint
	ProductId   uint
	OrderId     uint
	Quantity    int
	PriceAtSale float64
}

type PaymentMethod struct {
	UPITransaction string
	CashInHand     string
	CardPayment    string
}

type TransactionType struct {
	Debit  string
	Credit string
}
