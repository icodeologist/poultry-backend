package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		DB: db,
	}
}

func (a *AdminHandler) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
	var admin models.Admin
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if admin.AdminName == "" {
		slog.Warn("Empty or invalid name")
		http.Error(w, "product title cannot be emtpy", http.StatusBadRequest)
		return
	}
	// TODO : add secure password
	if len(admin.Password) < 8 {
		slog.Warn("Passwords must be atleast of 8 characters")
		http.Error(w, "password must be atleast of 8 characters", http.StatusBadRequest)
		return
	}
	var existingProduct models.Admin
	exRes := a.DB.Where("admin_email=?", admin.AdminEmail).First(&existingProduct)
	if exRes.RowsAffected != 0 {
		slog.Error("You have already created admin account. Please log in")
		http.Error(w, "email already exists", http.StatusConflict)
		return
	} else if !errors.Is(exRes.Error, gorm.ErrRecordNotFound) {
		slog.Error("DB error", "err", exRes.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	res := a.DB.Create(&admin)
	if res.Error != nil {
		slog.Error("Failed to create the admin", "err", res.Error)
		http.Error(w, "failed to create the admin product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(admin)
}

// func (a *AdminHandler) LoginAdmin(w http.ResponseWriter, r *http.Request) {
// 	var adminLoginInfo models.AdminLoginInfo
//
// }
