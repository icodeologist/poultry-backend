package handlers

//TODO: recorde payment should record credit to customer.
import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/icodeologist/poultry-backend/models"
	"github.com/icodeologist/poultry-backend/service"
)

type OrderHandler struct {
	OrdSvc *service.OrderService
}

func NewOrderHandler(ordsvc *service.OrderService) *OrderHandler {
	return &OrderHandler{OrdSvc: ordsvc}
}

func (o *OrderHandler) CreateNewOrder(w http.ResponseWriter, r *http.Request) {
	var checkOutReq models.CheckOutrequest
	if err := json.NewDecoder(r.Body).Decode(&checkOutReq); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order, err := o.OrdSvc.CreateOrder(checkOutReq.CustomerID, checkOutReq.Items)
	if err != nil {
		slog.Error("create_order_error", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)

}

func (o *OrderHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]
	orderid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		slog.Error("Converting string id to int", "err", err)
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	var req models.CheckOutPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	order, payment, _, err := o.OrdSvc.RecordPayment(uint(orderid), req.TenderedAmount, req.PaymentMethod, req.PayPreviousCredit, req.PayThroughCredit)
	if err != nil {
		slog.Error("Failed to record payment", "err", err)
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"order":   order,
		"payment": payment,
		"message": "payment succesfull",
	})
}

// no old pay no through credit just exact amounmt
// no old pay no thorugh credit more money - exact change
// old pay no through credit more money - change 0 customer.Balance -  980
