package handlers

import (
	"encoding/json"
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

func (a *AdminHandler) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	var userLoginInfo models.UserLoginInfo
	if err := json.NewDecoder(r.Body).Decode(&userLoginInfo); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// cross check password
	// TODO: hanlde email verification with @ and all symbol
	if userLoginInfo.Email == "" {
		slog.Warn("Empty or invalid email")
		http.Error(w, "email cannot be emtpy", http.StatusBadRequest)
		return
	}

	var checkPass models.Admin
	res := a.DB.Where("admin_email=?", userLoginInfo.Email).First(&checkPass)
	if res.Error != nil {
		slog.Error("No matches found")
		http.Error(w, "No email found.", http.StatusNotFound)
		return
	}
	if userLoginInfo.Password != checkPass.Password {
		slog.Error("Wrong Password")
		http.Error(w, "password do no match", http.StatusUnauthorized)
		return
	}
	claims := jwt.MapClaims{
		"role":    checkPass.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"user_id": checkPass.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretkey))
	if err != nil {
		slog.Error("Signing string with key to a token error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "user-authorization",
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
