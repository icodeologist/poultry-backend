package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/models"
)

func GetAdmin(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value("adminID")
	log.Println("admin id inside get/me : ", adminID)
	if adminID == 0 {
		http.Error(w, "unauthorized heere", http.StatusUnauthorized)
		return
	}

	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		http.Error(w, "admin not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAdminResponse(admin)) // strips password
}

func toAdminResponse(a models.Admin) models.Admin {
	log.Println("admin real : ", a)
	a.Password = "haha you wont get it"
	log.Println("admin fake : ", a)
	return a
}
