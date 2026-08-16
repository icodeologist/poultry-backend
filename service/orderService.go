package service

import (
	"errors"
	"fmt"
	"log"
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
			TimeStamp:      time.Now(),
			TotalAmount:    totalBill,
			PaymentBalance: totalBill,
			Status:         "Pending",
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
	if err != nil {
		log.Printf("Create Order transaction failed : %v\n", err)
		return models.Order{}, err
	}
	log.Printf("Order : %v\n", order)
	return order, nil
}

// separate way to record payment and update stocks
func (os *OrderService) RecordPayment(orderID uint, tendered float64, paymentMethod string) (models.Order, models.Payment, error) {
	var order models.Order
	var payment models.Payment

	err := os.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, orderID).Error; err != nil {
			return fmt.Errorf("fetching order with id %v: %v", orderID, err)
		}
		if order.Status == "Completed" {
			return fmt.Errorf("order is fully paid %v", orderID)
		}
		if tendered <= 0 {
			return fmt.Errorf("tendered cannot be 0 or negative, order %v", orderID)
		}

		amountDue := order.PaymentBalance
		var applied, change float64
		if tendered >= amountDue {
			applied = amountDue
			change = tendered - amountDue
		} else {
			applied = tendered
			change = 0
		}

		payment = models.Payment{
			OrderID:        orderID,
			AmountTendered: tendered,
			ChangeGiven:    change,
			Method:         paymentMethod,
			TimeStamp:      time.Now(),
		}
		if err := tx.Create(&payment).Error; err != nil {
			return fmt.Errorf("creating payment: %w", err)
		}

		order.PaymentBalance -= applied
		if order.PaymentBalance <= 0.01 {
			order.PaymentBalance = 0
			order.Status = "Completed"
		}
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("updating order: %w", err)
		}
		return nil
	})

	return order, payment, err
}
