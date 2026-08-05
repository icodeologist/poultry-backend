package models

type Customer struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phone_number"`
	Balance     float64 `json:"balance"`
	Orders      []Order
}
