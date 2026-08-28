package service

import (
	"errors"
	"fmt"
	"log"
	"math"
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

func (os *OrderService) CreateOrder(customerID *uint, items []models.CartlineInput) (models.Order, error) {
	// validate products so no dummy or lose products in the items
	// aslo get the total price
	validProducts, totalBill, err := os.ValidateCart(items)
	if err != nil {
		return models.Order{}, err
	}
	// now here if one db transaction fails we have to kind of mark it as a failure and start again
	// so making a transaction here  is wise
	// have to  create order then orderitems which are linked to single order and then  if something goes wrong cancel all of them
	// linking customer id with this order

	var customer models.Customer
	res := os.DB.Where("id=?", customerID).First(&customer)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return models.Order{}, fmt.Errorf("Customer not found : ID : %v\n", customerID)
		}
		// real DB error
		return models.Order{}, res.Error
	}
	var order models.Order
	err = os.DB.Transaction(func(tx *gorm.DB) error {
		order = models.Order{
			CustomerID:     &customer.ID,
			Customer:       customer,
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

// Recordpayment make a db transaction to make all the calls succeed or all calls to db fail
// The reason is since its a payment procedure and requires multiple models or data objects to alter and update slight uncertainity will make wrong writes to models
// So if one failes rest will never happen and the current one will wont write
// In this function we do 3 things. First For a given order with ORDER ID ill record the payment that they are giving whether cash or upi
// 2 is Im handling what if the customer wants to pay previous credit with remaining change (only applicable if the customer chose paymentMethod as - cash)
// 3 if the customer wants to pay entire order with credit- so store it to previous balance

func (os *OrderService) RecordPayment(orderID uint, tendered float64, paymentMethod string, payPreviousCredit bool, payThroughCredit bool) (models.Order, models.Payment, models.Customer, error) {
	var order models.Order
	var payment models.Payment
	var customer models.Customer

	err := os.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Customer").First(&order, orderID).Error; err != nil {
			return fmt.Errorf("fetching order with id %v: %v", orderID, err)
		}
		if order.Status == "Completed" {
			log.Printf("Order %v is full paid.", orderID)
			return nil
		}
		if err := tx.Where("id=?", order.CustomerID).First(&customer).Error; err != nil {
			return fmt.Errorf("fetch customer with id : %v\n", err)
		}
		amountDue := order.PaymentBalance
		var applied, change float64
		var isPaymentByCredit bool
		// if payment is through credit just update previous balance and make payemtn
		if payThroughCredit {
			customer.Balance += amountDue
			applied = amountDue
			isPaymentByCredit = true
		} else {
			if tendered <= 0 {
				return fmt.Errorf("tendered cannot be 0 or negative, order %v", orderID)
			}
			// if not then simply do basic math
			if tendered >= amountDue {
				applied = amountDue
				change = tendered - amountDue
			} else {
				applied = tendered
				change = 0.0
				shortFall := amountDue - tendered
				customer.Balance += shortFall
			}

			// if change is greater and only if custmoer wants to pay then proceed with this step
			if payPreviousCredit {
				if change > 0.0 && customer.Balance > 0.0 {
					creditPaid := math.Min(change, customer.Balance)
					customer.Balance -= creditPaid
					change -= creditPaid
				} else {
					log.Printf("insufficient funds : %v or Your balance is already paid : %v\n", change, customer.Balance)
				}
			}
		}
		// create  a payment
		payment = models.Payment{
			OrderID:              orderID,
			AmountTendered:       tendered,
			ChangeGiven:          change,
			Method:               paymentMethod,
			TimeStamp:            time.Now(),
			PaymentThroughCredit: isPaymentByCredit,
		}
		if err := tx.Create(&payment).Error; err != nil {
			return fmt.Errorf("creating payment: %w", err)
		}
		// update the balance of the order in order to determine that if this order is done or not
		order.PaymentBalance -= applied
		if order.PaymentBalance <= 0.01 {
			order.PaymentBalance = 0
			order.Status = "Completed"
		}
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("updating order: %w", err)
		}
		if err := tx.Save(&customer).Error; err != nil {
			return fmt.Errorf("updating customer: %w", err)
		}
		return nil
	})
	return order, payment, customer, err
}
