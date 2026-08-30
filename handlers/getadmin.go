package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/models"
)

// TODO: tehre is some bug here so lets fix after wards
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
	a.Password = "haha you wont get it"
	return a
}

func GetCurrentRole(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok {
		http.Error(w, fmt.Sprint("role cannot be empty"), http.StatusBadRequest)
		return
	}
	// var adminID, staffID float64
	if role == "admin" {
		adminIDVal, ok := r.Context().Value("adminID").(float64)
		if !ok {
			http.Error(w, fmt.Sprint("no id admin"), http.StatusBadRequest)
			return
		}
		log.Printf("adminValID : %v\n", adminIDVal)
	} else if role == "staff" {
		staffIDVal, ok := r.Context().Value("staffID").(float64)
		if !ok {
			http.Error(w, fmt.Sprint("no id staff"), http.StatusBadRequest)
			return
		}
		log.Printf("staffIDval : %v\n", staffIDVal)
	}
}
