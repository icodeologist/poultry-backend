package models

type Product struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	Title          string  `gorm:"not null" json:"title"`
	Price          float64 `gorm:"not null" json:"price"`
	Unit           string  `gorm:"size:20" json:"unit"` // kg piece or litre pack
	TaxRate        float64 `json:"taxRate"`
	Stock_Quantity int     `gorm:"stock_quantity" json:"stockQuantity"`
}
