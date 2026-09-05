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
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	r.HandleFunc("/customer/register", customerHandler.CreateNewUser)
	r.HandleFunc("/customer/{phone}", customerHandler.GetCustomerByPhone)
	r.HandleFunc("/admin/login", adminHandler.LoginAdmin)
	r.HandleFunc("/customers/{id}", customerHandler.AddBalance)
	r.HandleFunc("/credit/summary", customerHandler.GetCreditSummary).Methods("GET")
	r.HandleFunc("/invoice/{id}", orderHandler.GetInvoice).Methods("GET")
	r.HandleFunc("/invoices", orderHandler.GetInvoices).Methods("GET")

	productRoutes := r.NewRoute().Subrouter()
	productRoutes.Use(middleware.Authenticate)
	productRoutes.Use(middleware.RequireRoles("admin", "staff"))
	productRoutes.HandleFunc("/api/products", productHandler.GetAllProductsFromDB).Methods("GET")
	productRoutes.HandleFunc("/api/products/{id}", productHandler.GetProductByID).Methods("GET")
	productRoutes.HandleFunc("/api/products/{id}/stock", productHandler.UpdateProductStck).Methods("PATCH")
	// Keep the legacy read paths authenticated for existing clients.
	productRoutes.HandleFunc("/products", productHandler.GetAllProductsFromDB).Methods("GET")
	productRoutes.HandleFunc("/products/{id}", productHandler.GetProductByID).Methods("GET")

	adminProducts := r.NewRoute().Subrouter()
	adminProducts.Use(middleware.UserMiddleware)
	adminProducts.HandleFunc("/api/products", productHandler.AddProduct).Methods("POST")
	adminProducts.HandleFunc("/api/products/{id}", productHandler.UpdateProduct).Methods("PATCH")
	adminProducts.HandleFunc("/api/products/{id}", productHandler.DeleteProduct).Methods("DELETE")
	// Keep legacy management paths, now with explicit admin authorization.
	adminProducts.HandleFunc("/products/new", productHandler.AddProduct).Methods("POST")
	adminProducts.HandleFunc("/product/editprice/{id}", productHandler.EditPrice).Methods("PATCH", "PUT")
	adminProducts.HandleFunc("/products/edit/{id}", productHandler.UpdateProduct).Methods("PATCH")

	adminRoute := r.NewRoute().Subrouter()
	adminRoute.Use(middleware.UserMiddleware)
	adminRoute.HandleFunc("/dummy", dummy)
	adminRoute.HandleFunc("/admin/product/{id}", productHandler.UpdateProduct)
	adminRoute.HandleFunc("/admin/products/{id}", productHandler.DeleteProduct)
	// adminRoute.HandleFunc("/admin/me", handlers.GetAdmin)
	adminRoute.HandleFunc("/admin/me", handlers.GetAdmin)
	adminRoute.HandleFunc("/role", handlers.GetCurrentRole)

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
