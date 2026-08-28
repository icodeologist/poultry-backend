package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/handlers"
	"github.com/icodeologist/poultry-backend/middleware"

	// "github.com/icodeologist/poultry-backend/middleware"
	"github.com/icodeologist/poultry-backend/service"
	"github.com/rs/cors"
)

func main() {
	db.ConnectTODB()
	r := mux.NewRouter()
	customerHandler := handlers.NewCustomerHandler(db.DB)
	// searchHandler := handlers.NewSearchUserCustomHandler(db.DB)
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
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	r.HandleFunc("/customer/register", customerHandler.CreateNewUser)
	r.HandleFunc("/customer/{phone}", customerHandler.GetCustomerByPhone)
	// r.HandleFunc("/search/{phonenumber}", searchHandler.SearchByPhoneNumber)
	r.HandleFunc("/admin/register", adminHandler.RegisterAdmin)
	r.HandleFunc("/admin/login", adminHandler.LoginAdmin)
	r.HandleFunc("/product/create", productHandler.AddProduct)
	r.HandleFunc("/product/editprice/{id}", productHandler.EditPrice)
	// r.HandleFunc("/product/updateStock/{id}", productHandler.UpdateProductStck)
	r.HandleFunc("/products", productHandler.GetAllProductsFromDB)
	// r.HandleFunc("/new-order", orderHandler.CreateNewOrder)
	// r.HandleFunc("/orders/{id}/payment", orderHandler.RecordPayment)
	r.HandleFunc("/customers/{id}", customerHandler.AddBalance)
	r.HandleFunc("/updateStock/{id}", productHandler.UpdateProduct)

	adminRoute := r.NewRoute().Subrouter()
	adminRoute.Use(middleware.AdminMiddleware)
	adminRoute.HandleFunc("/dummy", dummy)
	adminRoute.HandleFunc("/admin/product/{id}", productHandler.UpdateProduct)
	adminRoute.HandleFunc("/admin/products/{id}", productHandler.DeleteProduct)
	adminRoute.HandleFunc("/admin/me", handlers.GetAdmin)

	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.IdempotencyMiddleware(db.DB))
	protected.HandleFunc("/new/order", orderHandler.CreateNewOrder).Methods("POST")
	protected.HandleFunc("/orders/{id}/payment", orderHandler.RecordPayment).Methods("POST")

	log.Fatal(http.ListenAndServe(":5000", c.Handler(r)))
}

func dummy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("the dummy endpoint was hit")
	fmt.Fprint(w, "Hello people")
}
