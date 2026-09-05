package main

import (
	"fmt"
	"log"
	"os"

	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/models"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println("usage: createuser <email> <name> <password> <role>")
		os.Exit(1)
	}
	email, name, password, role := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	if role != "admin" && role != "staff" {
		log.Fatal("role must be 'admin' or 'staff'")
	}

	// Use the application's connection and migration path so provisioned
	// accounts are written to the same database used by login.
	db.ConnectTODB()

	user := models.Admin{
		AdminEmail: email,
		AdminName:  name,
		Password:   password,
		Role:       role,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		log.Fatal("could not create user:", err)
	}

	fmt.Printf("Created %s account for %s (ID: %d)\n", role, email, user.ID)
}
