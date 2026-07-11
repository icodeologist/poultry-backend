package models

type Product struct {
	ID      uint
	Title   string
	Price   float32 // price per kg or unit
	InStock uint
}
