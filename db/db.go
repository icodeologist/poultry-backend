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
	// allowing TranslateError true so it catches when duplicated key is trying to create new row
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         nil,
		TranslateError: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to DB : %v\n", err)
	}

	err = DB.AutoMigrate(&models.Customer{}, &models.Product{}, &models.Admin{}, &models.Order{}, &models.OrderItem{}, &models.Payment{}, models.OrderAndPaymentIdempotency{})
	if err != nil {
		log.Fatalf("Migrations failed %v\n", err)
	}
	log.Println("Migrations run Successfully")

	log.Println("Successfully connected to DB")
}
