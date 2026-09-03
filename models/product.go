package models

type Product struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	Title          string  `json:"title"`
	Price          float64 `json:"price"`
	Unit           string  `json:"unit"` // kg piece or litre pack
	TaxRate        float64 `json:"taxRate"`
	Stock_Quantity int     `gorm:"stock_quantity" json:"stockQuantity"`
}
