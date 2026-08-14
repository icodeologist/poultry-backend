package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/handlers"
	"github.com/icodeologist/poultry-backend/service"
	"github.com/rs/cors"
)

func main() {
	db.ConnectTODB()
	r := mux.NewRouter()
	customerHandler := handlers.NewCustomerHandler(db.DB)
	searchHandler := handlers.NewSearchUserCustomHandler(db.DB)
	adminHandler := handlers.NewAdminHandler(db.DB)
	productServiceHandler := service.NewProductService(db.DB)
	productHandler := handlers.NewProductHandler(productServiceHandler)
	orderService := service.NewOrderService(db.DB)
	orderHandler := handlers.NewOrderHandler(orderService)
	// input := models.CheckOutrequest{
	// 	Items: []models.CartlineInput{
	// 		{
	// 			ProductID: 1,
	// 			Unit:      "kg",
	// 			Quantity:  2,
	// 		},
	// 		{
	// 			ProductID: 2,
	// 			Unit:      "piece",
	// 			Quantity:  3,
	// 		},
	// 		{
	// 			ProductID: 3,
	// 			Unit:      "litre",
	// 			Quantity:  1,
	// 		},
	// 	},
	// 	PaymentMethod: "COD",
	// }
	//
	// validProduts, total, err := orderService.ValidateCart(input.Items)
	// if err != nil {
	// 	log.Fatalf("Err : %v\n", err)
	// }
	// for _, p := range validProduts {
	// 	log.Printf("product : %v\n --- Quantity : %v\n", p.Product, p.Quantity)
	// }
	// log.Println("Total Bill : ", total)
	//
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	r.HandleFunc("/create", customerHandler.CreateNewUser)
	r.HandleFunc("/search/{phonenumber}", searchHandler.SearchByPhoneNumber)
	r.HandleFunc("/admin/register", adminHandler.RegisterAdmin)
	r.HandleFunc("/admin/login", adminHandler.LoginAdmin)
	r.HandleFunc("/product/create", productHandler.AddProduct)
	r.HandleFunc("/product/editprice/{id}", productHandler.EditPrice)
	r.HandleFunc("/product/delete/{id}", productHandler.DeleteProduct)
	r.HandleFunc("/products", productHandler.GetAllProductsFromDB)
	r.HandleFunc("/new-order", orderHandler.CheckOutFlow)
	log.Fatal(http.ListenAndServe(":3000", c.Handler(r)))
}
