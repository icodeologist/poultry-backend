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
