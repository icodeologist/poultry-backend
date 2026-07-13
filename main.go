package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/handlers"
)

func main() {
	db.ConnectTODB()
	r := mux.NewRouter()
	customerHandler := handlers.NewCustomerHandler(db.DB)
	productHandler := handlers.NewProductHandler(db.DB)
	searchHandler := handlers.NewSearchUserCustomHandler(db.DB)

	r.HandleFunc("/create", customerHandler.CreateNewUser)
	r.HandleFunc("/add/product", productHandler.AddProduct)
	r.HandleFunc("/search/{phonenumber}", searchHandler.SearchByPhoneNumber)
	log.Fatal(http.ListenAndServe(":3000", r))
}
