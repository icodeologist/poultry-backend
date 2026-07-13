package models

type Product struct {
	ID      uint    `gorm:"primaryKey"`
	Title   string  `gorm:"title"`
	Price   float32 `gorm:"price"`
	InStock uint    `gorm:"instock"`
}
