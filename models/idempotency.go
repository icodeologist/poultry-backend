package models

import (
	"time"
)

type OrderAndPaymentIdempotency struct {
	ID             uint   `gorm:"primaryKey"`
	Key            string `gorm:"unique"`
	RequestHash    string
	ResponseBody   string
	ResponseStatus int
	CreatedAt      time.Time
}
