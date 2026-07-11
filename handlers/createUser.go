package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"log/slog"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	DB *gorm.DB
}

func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{
		DB: db,
	}
}

func (ns *CustomerHandler) CreateNewUser(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if customer.Name == "" {
		slog.Error("Invalid json")
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(customer.PhoneNumber) != 10 {
		slog.Info("phone number length", "len(phonenumber)", len(customer.PhoneNumber))
		slog.Error("Invalid phone number")
		http.Error(w, "invalid phone number", http.StatusBadRequest)
		return
	}

	slog.Info("customer details", "customer", customer)

	// check if the number already exists
	var existing models.Customer
	res := ns.DB.Where("phone_number = ?", customer.PhoneNumber).First(&existing)
	if res.RowsAffected != 0 {
		slog.Error("User already exists")
		http.Error(w, "user already exists with the phone number", http.StatusConflict)
		return
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		slog.Error("DB error", "err", res.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	result := ns.DB.Create(&customer)
	if result.Error != nil {
		slog.Error("Failed to create customer", "err", result.Error)
		http.Error(w, "failed to create customer", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}
