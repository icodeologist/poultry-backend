package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type CreditSummary struct {
	TotalCredit               float64 `json:"total_credit"`
	PendingCollections        int64   `json:"pending_collections"`
	TotalCustomers            int64   `json:"total_customers"`
	PendingCustomerPercentage float64 `json:"pending_customer_percentage"`
	SettlementHealth          string  `json:"settlement_health"`
}

type CreditLedgerCustomer struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	PhoneNumber  string     `json:"phone_number"`
	Balance      float64    `json:"balance"`
	LastPurchase *time.Time `json:"last_purchase"`
}

// GetCreditSummary returns the current customer-credit exposure. A balance is
// considered settled at one paisa or less so floating-point residue does not
// turn a settled account into a pending collection.
func (ch *CustomerHandler) GetCreditSummary(w http.ResponseWriter, _ *http.Request) {
	var summary CreditSummary
	result := ch.DB.Model(&struct{ ID uint }{}).
		// TODO: understand this and write it yourself
		Table("customers").
		Select(`COALESCE(SUM(CASE WHEN balance > 0.01 THEN balance ELSE 0 END), 0) AS total_credit,
			COUNT(*) FILTER (WHERE balance > 0.01) AS pending_collections,
			COUNT(*) AS total_customers`).
		Scan(&summary)
	if result.Error != nil {
		slog.Error("failed to build credit summary", "err", result.Error)
		http.Error(w, "failed to load credit summary", http.StatusInternalServerError)
		return
	}

	if summary.TotalCustomers == 0 || summary.PendingCollections == 0 {
		summary.SettlementHealth = "Clear"
	} else {
		summary.PendingCustomerPercentage = float64(summary.PendingCollections) / float64(summary.TotalCustomers) * 100
		switch {
		case summary.PendingCustomerPercentage < 25:
			summary.SettlementHealth = "Healthy"
		case summary.PendingCustomerPercentage < 50:
			summary.SettlementHealth = "Watch"
		default:
			summary.SettlementHealth = "At risk"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetCreditLedger returns real customer balances together with each
// customer's latest order date. Customers without orders are retained.
func (ch *CustomerHandler) GetCreditLedger(w http.ResponseWriter, _ *http.Request) {
	customers := make([]CreditLedgerCustomer, 0)
	result := ch.DB.Table("customers").
		Select(`customers.id, customers.name, customers.phone_number, customers.balance,
			(SELECT MAX(orders.time_stamp) FROM orders WHERE orders.customer_id = customers.id) AS last_purchase`).
		Order("customers.balance DESC, customers.name ASC").
		Scan(&customers)
	if result.Error != nil {
		slog.Error("failed to load credit ledger", "err", result.Error)
		http.Error(w, "failed to load credit ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}
