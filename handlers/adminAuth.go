package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

const secretkey = "hello"

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

func (a *AdminHandler) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	var adminLoginInfo models.AdminLoginInfo
	if err := json.NewDecoder(r.Body).Decode(&adminLoginInfo); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// cross check password
	// TODO: hanlde email verification with @ and all symbol
	if adminLoginInfo.Email == "" {
		slog.Warn("Empty or invalid email")
		http.Error(w, "email cannot be emtpy", http.StatusBadRequest)
		return
	}

	var checkPass models.Admin
	res := a.DB.Where("admin_email=?", adminLoginInfo.Email).First(&checkPass)
	if res.Error != nil {
		slog.Error("No matches found")
		http.Error(w, "No email found.", http.StatusNotFound)
		return
	}
	if adminLoginInfo.Password != checkPass.Password {
		slog.Error("Wrong Password")
		http.Error(w, "password do no match", http.StatusUnauthorized)
		return
	}
	claims := jwt.MapClaims{
		"role": "admin",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretkey))
	if err != nil {
		slog.Error("Signing string with key to a token error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "admin-authorization",
		Value:    signedToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // flip to true in prod via env var
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode("You logged in as a admin")
}
