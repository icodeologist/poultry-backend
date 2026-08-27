package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

func (ns *CustomerHandler) GetCustomerByPhone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phone := vars["phone"]

	if len(phone) != 10 {
		http.Error(w, "invalid phone number", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	res := ns.DB.Where("phone_number = ?", phone).First(&customer)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	} else if res.Error != nil {
		slog.Error("DB error", "err", res.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customer)
}
