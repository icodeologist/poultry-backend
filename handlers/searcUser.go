package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gorm.io/gorm"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/models"
)

// CustomerHandler holds dependencies (like the DB) so they can be injected
// rather than relying on a global.
type SearchUserCustomHandler struct {
	DB *gorm.DB
}

func NewSearchUserCustomHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{DB: db}
}

func (h *CustomerHandler) SearchByPhoneNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phoneNumber := vars["phonenumber"]

	if len(phoneNumber) != 10 {
		http.Error(w, "invalid phone number", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	res := h.DB.Where("phone_number = ?", phoneNumber).First(&customer)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "customer not found", http.StatusNotFound)
			return
		}
		slog.Error("DB error while searching customer", "err", res.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	returnResponse := models.UserAccountInfo{
		UserName:        customer.Name,
		UserPhoneNumber: customer.PhoneNumber,
		UserBalance:     customer.Balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(returnResponse)
}
