package models

import "time"

type Payment struct {
	ID                   uint `gorm:"primaryKey"`
	OrderID              uint
	AmountTendered       float64
	ChangeGiven          float64
	Method               string
	TimeStamp            time.Time
	PaymentThroughCredit bool
}
