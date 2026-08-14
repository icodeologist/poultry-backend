package service

import (
	"errors"
	"fmt"
	"time"

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

func (os *OrderService) ValidateCart(items []models.CartlineInput) ([]ValidProduct, float64, error) {
	var validProducts []ValidProduct
	var totalBill float64
	for _, lineProduct := range items {
		var product models.Product
		res := os.DB.First(&product, lineProduct.ProductID)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// something was missing and mid checkout admin updated something
			// so if one invalid product is in the cart line whole cehck out should cancel and restart
			return nil, 0.0, fmt.Errorf("product %v is no longer available", lineProduct.ProductID)
		}
		vp := ValidProduct{
			Product:  product,
			Quantity: float64(lineProduct.Quantity),
		}
		validProducts = append(validProducts, vp)

		// calculate total per unit
		// unit depends on the product
		totalBill += product.Price * float64(lineProduct.Quantity)
	}
	return validProducts, totalBill, nil
}

func (os *OrderService) CreateOrder(items []models.CartlineInput) (models.Order, error) {
	// validate products so no dummy or lose products in the items
	// aslo get the total price
	validProducts, totalBill, err := os.ValidateCart(items)
	if err != nil {
		return models.Order{}, err
	}
	// now here if one db transaction fails we have to kind of mark it as a failure and start again
	// so making a transaction here  is wise
	// have to  create order then orderitems which are linked to single order and then  if something goes wrong cancel all of them
	// TODO: find a way to link customer id to order
	var order models.Order
	err = os.DB.Transaction(func(tx *gorm.DB) error {
		order = models.Order{
			TimeStamp:   time.Now(),
			TotalAmount: totalBill,
			// PaymentMethod:  paymentMethod,
			// PaymentBalance: totalBill - amountPaid,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// then loop through products and create separate orderedItem and add them to order
		for _, v := range validProducts {
			if v.Product.Stock_Quantity < int(v.Quantity) {
				return fmt.Errorf("insufficient stock for %v", v.Product.Title)
			}
			if err := tx.Model(&v.Product).Update("Stock_Quantity", v.Product.Stock_Quantity-int(v.Quantity)).Error; err != nil {
				return err
			}
			item := models.OrderItem{
				ProductId:   v.Product.ID,
				OrderId:     order.ID,
				Quantity:    int(v.Quantity),
				PriceAtSale: v.Product.Price,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return order, nil
}
