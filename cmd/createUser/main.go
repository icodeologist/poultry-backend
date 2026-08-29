package main

import (
	"fmt"
	"log"
	"os"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	dsn := "postgres://poultry_admin:idontknow@localhost:5432/poultry_db?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	user := models.Admin{
		AdminEmail: email,
		AdminName:  name,
		Password:   password,
		Role:       role,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatal("could not create user:", err)
	}

	fmt.Printf("Created %s account for %s (ID: %d)\n", role, email, user.ID)
}
