package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		DB: db,
	}
}
func (p *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if product.Title == "" {
		slog.Warn("Empty product name")
		http.Error(w, "product title cannot be emtpy", http.StatusBadRequest)
		return
	}
	if product.Price == 0.0 {
		slog.Warn("Product's price cannot be 0")
		http.Error(w, "product's price cannot be 0", http.StatusBadRequest)
		return
	}
	var existingProduct models.Product
	exRes := p.DB.Where("title=?", product.Title).First(&existingProduct)
	if exRes.RowsAffected != 0 {
		slog.Error("Product already exists", "Title", product.Title)
		http.Error(w, "product already exists", http.StatusConflict)
		return
	} else if !errors.Is(exRes.Error, gorm.ErrRecordNotFound) {
		slog.Error("DB error", "err", exRes.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return

	}
	res := p.DB.Create(&product)
	if res.Error != nil {
		slog.Error("Failed to add the product", "err", res.Error)
		http.Error(w, "failed to add the product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}
