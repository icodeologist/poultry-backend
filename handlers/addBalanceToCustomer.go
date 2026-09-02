package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

type AddBalanceRequest struct {
	Balance float64 `json:"balance"`
}

func (ch *CustomerHandler) AddBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	customerID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		slog.Error("invalid customer id", "err", err)
		http.Error(w, "invalid customer id", http.StatusBadRequest)
		return
	}

	var req AddBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	res := ch.DB.First(&customer, uint(customerID))
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	} else if res.Error != nil {
		slog.Error("db error", "err", res.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	customer.Balance += req.Balance

	if err := ch.DB.Save(&customer).Error; err != nil {
		slog.Error("failed to update customer balance", "err", err)
		http.Error(w, "failed to update balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customer)
}
