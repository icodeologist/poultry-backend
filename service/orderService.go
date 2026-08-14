package service

import (
	"errors"
	"fmt"
	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{DB: db}
}

// FIX : normal unit - int others - float64
type ValidProduct struct {
	Product  models.Product
	Quantity float64
}

func (os *OrderService) ValidateCart(items []models.CartlineInput) ([]ValidProduct, error) {
	var validProducts []ValidProduct
	var totalBill float64
	for _, lineProduct := range items {
		var product models.Product
		res := os.DB.First(&product, lineProduct.ProductID)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// something was missing and mid checkout admin updated something
			// so if one invalid product is in the cart line whole cehck out should cancel and restart
			return nil, fmt.Errorf("product %v is no longer available", lineProduct.ProductID)
		}
		vp := ValidProduct{
			Product:  product,
			Quantity: float64(lineProduct.Quantity),
		}
		validProducts = append(validProducts, vp)

		// calculate total per unit
		// unit depends on the product
		oneUnitPrice := product.Price
		totalBill += oneUnitPrice * float64(lineProduct.Quantity)

	}
	return validProducts, nil
}
