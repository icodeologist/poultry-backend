package db

import (
	"log"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectTODB() {
	dsn := "postgres://poultry_admin:idontknow@localhost:5432/poultry_db?sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: nil,
	})

	if err != nil {
		log.Fatalf("Failed to connect to DB : %v\n", err)
	}

	err = DB.AutoMigrate(&models.Customer{}, &models.Product{})
	if err != nil {
		log.Fatalf("Migrations failed %v\n", err)
	}

	log.Println("Successfully connected to DB")
}
