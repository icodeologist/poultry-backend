package handlers

//TODO: recorde payment should record credit to customer.
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	log.Println("JUST GOT HIT HMM")
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
		"order":        order,
		"payment":      payment,
		"message":      "payment succesfull",
		"change_given": payment.ChangeGiven,
	})
	// json.NewEncoder(w).Encode(payment)
}

func (o *OrderHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	orderID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || orderID == 0 {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	invoice, err := o.OrdSvc.GetInvoice(uint(orderID))
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		slog.Error("Failed to create invoice", "err", err)
		http.Error(w, "failed to create invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoice)
}

func (o *OrderHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			http.Error(w, "limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	invoices, err := o.OrdSvc.GetInvoices(limit)
	if err != nil {
		slog.Error("Failed to get invoices", "err", err)
		http.Error(w, "failed to get invoices", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

// no old pay no through credit just exact amounmt
// no old pay no thorugh credit more money - exact change
// old pay no through credit more money - change 0 customer.Balance -  980
