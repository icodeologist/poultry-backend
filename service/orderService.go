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
func (os *OrderService) RecordPayment(orderID uint, tendered float64, paymentMethod string, payPreviousCredit bool, payThroughCredit bool) (models.Order, models.Payment, error) {
	var order models.Order
	var payment models.Payment
	var customer models.Customer

	err := os.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, orderID).Error; err != nil {
			return fmt.Errorf("fetching order with id %v: %v", orderID, err)
		}
		if order.Status == "Completed" {
			log.Printf("Order %v is full paid.", orderID)
			return nil
		}
		// fetch customer this if customer wants to pay the credit
		// fetching first anyways because need to update balance
		if err := tx.Where("id=?", order.CustomerID).First(&customer).Error; err != nil {
			return fmt.Errorf("fetch customer with id : %v\n", err)
		}
		amountDue := order.PaymentBalance
		var applied, change float64
		// just update customer balance
		log.Printf("Value of paythorughCredit : %v\n", payThroughCredit)
		if payThroughCredit == true {
			log.Println("Didnt go in here - if paythorugh credit was true")
			preBalance := customer.Balance
			customer.Balance += amountDue
			log.Printf("order : %v paid with full credit", order.ID)
			log.Printf("Previous balance : %v Current balance : %v Diffrence : %v\n", preBalance, customer.Balance, customer.Balance-preBalance)
			log.Printf("Current order : %v total Bill %v\n", order.ID, order.PaymentBalance)

		} else {
			log.Println("Went here directly DORA")
			if tendered <= 0 {
				return fmt.Errorf("tendered cannot be 0 or negative, order %v", orderID)
			}

			// if customer wants to just pay the current bill and given tenndered amount >= amount due apply
			if tendered >= amountDue {
				applied = amountDue
				change = tendered - amountDue
			} else {
				applied = tendered
				change = 0.0
			}
			var previousBalance float64
			var subFromBalance float64
			//if he wants to pay for old credit
			if payPreviousCredit {
				// check if he has enough
				if change > 0.0 {
					subFromBalance = change
					previousBalance = customer.Balance
					customer.Balance = subFromBalance - customer.Balance
					if err := tx.Save(&customer).Error; err != nil {
						return fmt.Errorf("save to db customer : %v\n", err)
					}
					log.Printf("updated customer %v with balance %v and paid %v\n ", customer.ID, previousBalance, subFromBalance)
				}
			}
		}

		log.Printf("Payment was done through : %v\n", paymentMethod)

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
