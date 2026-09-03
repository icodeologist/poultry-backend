package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/models"
	"github.com/icodeologist/poultry-backend/service"
)

type ProductHandler struct {
	Svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		Svc: svc,
	}
}
func (p *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		writeProductError(w, service.ErrNotFound)
		return
	}

	product, err := p.Svc.GetProductByID(uint(id))
	if err != nil {
		writeProductError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (p *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	createdProduct, err := p.Svc.AddProduct(product)
	if err != nil {
		writeProductError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)

}

type editPriceReq struct {
	Price float64 `json:"price"`
}

func (p *ProductHandler) EditPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeProductError(w, err)
		return
	}
	var ep editPriceReq
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	product, err := p.Svc.EditPrice(uint(id), ep.Price)
	if err != nil {
		writeProductError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeProductError(w, err)
		return
	}

	err = p.Svc.DeleteProduct(uint(id))
	if err != nil {
		writeProductError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateStock struct {
	Stock int `json:"stock"`
}

func (p *ProductHandler) UpdateProductStck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeProductError(w, err)
		return
	}
	var up updateStock
	if err := json.NewDecoder(r.Body).Decode(&up); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	pdct, err := p.Svc.UpdateProductStock(uint(id), up.Stock)
	if err != nil {
		writeProductError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(pdct)

}

func (p *ProductHandler) GetAllProductsFromDB(w http.ResponseWriter, r *http.Request) {
	products, err := p.Svc.GetAllProducts()
	if err != nil {
		writeProductError(w, err)
		return
	}
	w.WriteHeader(http.StatusContinue)
	json.NewEncoder(w).Encode(products)
}

// handler where admin updates the product
// pointer so that we can difference between nil value and actual 0
type UpdateProductInfo struct {
	Title         *string  `json:"title"`
	Price         *float64 `json:"price"`
	StockQuantity *int     `json:"stock"`
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeProductError(w, err)
		return
	}

	var req UpdateProductInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid json", "err", err)
		writeProductError(w, err)
		return
	}
	// updates gives the only field we need to update in the product model
	updates := map[string]any{}

	if req.Title != nil {
		if *req.Title == "" {
			writeProductError(w, err)
			return
		}
		updates["title"] = *req.Title
	}
	if req.Price != nil {
		if *req.Price < 0 {
			writeProductError(w, err)
			return
		}
		updates["price"] = *req.Price
	}
	if req.StockQuantity != nil {
		if *req.StockQuantity < 0 {
			writeProductError(w, err)
			return
		}
		updates["stock_quantity"] = *req.StockQuantity
	}

	if len(updates) == 0 {
		writeProductError(w, err)
		return
	}

	product, err := h.Svc.UpdateProduct(id, updates)
	if errors.Is(err, service.ErrNotFound) {
		writeProductError(w, err)
		return
	} else if err != nil {
		slog.Error("update product failed", "err", err)
		writeProductError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func writeProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrEmptyTitle):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrDuplicate):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		slog.Error("Unhandled service error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
