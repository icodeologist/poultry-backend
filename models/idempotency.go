package models

import (
	"time"
)

type OrderIdempotency struct {
	ID             uint   `gorm:"primaryKey"`
	Key            string `gorm:"unique"`
	RequestHash    string
	ResponseBody   string
	ResponseStatus int
	CreatedAt      time.Time
}

type PaymentIdempotency struct {
	Key            string `gorm:"primaryKey"`
	RequestHash    string
	ResponseBody   string
	ResponseStatus int
	CreatedAt      time.Time
}
