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

func (os *OrderService) ValidateCart(items []models.CartlineInput) ([]models.Product, error) {
	var products []models.Product
	for i, lineProduct := range items {
		fmt.Printf("I : %v\n", i)
		fmt.Printf("lineProduct : %v\n", lineProduct)
		var product models.Product
		res := os.DB.First(&product, lineProduct.ProductID)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// something was missing and mid checkout admin updated something
			// so if one invalid product is in the cart line whole cehck out should cancel and restart
			return nil, fmt.Errorf("product %v is no longer available", lineProduct.ProductID)
		}
		products = append(products, product)
	}
	return products, nil
}

func (os *OrderService) CalculateTotal(items []models.CartlineInput) (float64, error) {
	var total float64
	for _, lItem := range items {
		var product models.Product
		res := os.DB.First(&product, lItem.ProductID)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0.0, fmt.Errorf("product %v is no longer available", lItem.ProductID)
		}
		// current product per unit price
		oneUnitPrice := product.Price
		fmt.Println("Current per unit price : ", oneUnitPrice)
		fmt.Println("Last total : ", total)
		total += oneUnitPrice * float64(lItem.Quantity)
	}
	return total, nil

}
