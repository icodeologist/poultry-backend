package service

import (
	"errors"
	"log/slog"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

// Business-meaning errors. The handler maps these to HTTP statuses —
// it never needs to know *why* the price was invalid, just that it was.
var (
	ErrEmptyTitle   = errors.New("product title cannot be empty")
	ErrInvalidPrice = errors.New("product price must be greater than 0")
	ErrDuplicate    = errors.New("product already exists")
	ErrNotFound     = errors.New("product not found")
)

type ProductService struct {
	DB *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{DB: db}
}

func (s *ProductService) AddProduct(product models.Product) (models.Product, error) {
	//:TODO add product title
	// if product.Title == "" {
	// 	slog.Warn("Empty product name")
	// 	return models.Product{}, ErrEmptyTitle
	// }
	// TODO: Remove the comment here
	// if product.Price == 0.0 {
	// 	slog.Warn("Product's price cannot be 0")
	// 	return models.Product{}, ErrInvalidPrice
	// }

	var existingProduct models.Product
	exRes := s.DB.Where("title=?", product.Title).First(&existingProduct)
	if exRes.RowsAffected != 0 {
		slog.Error("Product already exists", "Title", product.Title)
		return models.Product{}, ErrDuplicate
	} else if !errors.Is(exRes.Error, gorm.ErrRecordNotFound) {
		slog.Error("DB error", "err", exRes.Error)
		return models.Product{}, exRes.Error // unexpected DB error, handler treats as 500
	}

	res := s.DB.Create(&product)
	if res.Error != nil {
		slog.Error("Failed to add the product", "err", res.Error)
		return models.Product{}, res.Error
	}
	return product, nil
}

// EditPrice is its own method because it's a distinct business action,
// not just a generic "update any field" — e.g. this is the one place
// you'd add a rule like "reject price changes over 500% without confirmation."
func (s *ProductService) EditPrice(productID uint, newPrice float64) (models.Product, error) {
	if newPrice <= 0.0 {
		slog.Warn("Invalid price on edit", "price", newPrice)
		return models.Product{}, ErrInvalidPrice
	}

	var product models.Product
	res := s.DB.First(&product, productID)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		slog.Warn("Product not found for price edit", "id", productID)
		return models.Product{}, ErrNotFound
	} else if res.Error != nil {
		slog.Error("DB error", "err", res.Error)
		return models.Product{}, res.Error
	}

	product.Price = newPrice
	if err := s.DB.Save(&product).Error; err != nil {
		slog.Error("Failed to update price", "err", err)
		return models.Product{}, err
	}
	return product, nil
}

func (s *ProductService) DeleteProduct(productID uint) error {
	var product models.Product
	res := s.DB.First(&product, productID)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		slog.Warn("Product not found for delete", "id", productID)
		return ErrNotFound
	} else if res.Error != nil {
		slog.Error("DB error", "err", res.Error)
		return res.Error
	}

	// This is exactly the spot where a future rule would live, e.g.:
	// "can't delete a product referenced in an unpaid order" —
	// you'd check that here, before calling Delete below.

	if err := s.DB.Delete(&product).Error; err != nil {
		slog.Error("Failed to delete product", "err", err)
		return err
	}
	return nil
}

func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	resError := s.DB.Find(&products).Error
	if resError != nil {
		slog.Error("DB error", "err", resError)
		return products, resError
	}
	return products, nil
}

func (s *ProductService) UpdateProductStock(productID uint, updateUnits int) (models.Product, error) {
	var product models.Product
	res := s.DB.First(&product, productID)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		slog.Warn("Product not found for delete", "id", productID)
		return product, ErrNotFound
	} else if res.Error != nil {
		slog.Error("DB error", "err", res.Error)
		return product, res.Error
	}
	product.Stock_Quantity += updateUnits

	if err := s.DB.Save(&product).Error; err != nil {
		slog.Error("Failed to save product", "err", err)
		return product, err
	}
	return product, nil
}
