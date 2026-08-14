package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/icodeologist/poultry-backend/models"
	"github.com/icodeologist/poultry-backend/service"
)

type OrderHandler struct {
	OrdSvc *service.OrderService
}

func NewOrderHandler(ordsvc *service.OrderService) *OrderHandler {
	return &OrderHandler{OrdSvc: ordsvc}
}

func (o *OrderHandler) CheckOutFlow(w http.ResponseWriter, r *http.Request) {
	var checkOutReq models.CheckOutrequest
	if err := json.NewDecoder(r.Body).Decode(&checkOutReq); err != nil {
		slog.Error("Invalid json", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order, err := o.OrdSvc.CreateOrder(checkOutReq.Items)
	if err != nil {
		slog.Error("create_order_error", "err", err)
		http.Error(w, "Create Order Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)

}
