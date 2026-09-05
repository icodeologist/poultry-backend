package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ReportPaymentChannel struct {
	Method string  `json:"method"`
	Amount float64 `json:"amount"`
}

type ReportTopProduct struct {
	ID       uint    `json:"id"`
	Title    string  `json:"title"`
	Quantity float64 `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

type ReportTimingPeriod struct {
	Period string `json:"period"`
	Label  string `json:"label"`
	Orders int64  `json:"orders"`
}

type SalesReport struct {
	GrossBilled       float64                `json:"gross_billed"`
	TotalOrders       int64                  `json:"total_orders"`
	ItemsSold         float64                `json:"items_sold"`
	AverageOrderValue float64                `json:"average_order_value"`
	PaymentChannels   []ReportPaymentChannel `json:"payment_channels"`
	TopProducts       []ReportTopProduct     `json:"top_products"`
	CustomerTiming    []ReportTimingPeriod   `json:"customer_timing"`
}

// GetSalesReport builds reporting aggregates directly from orders, payments,
// order items, and products so the dashboard never relies on preview data.
func (o *OrderHandler) GetSalesReport(w http.ResponseWriter, _ *http.Request) {
	report := SalesReport{
		PaymentChannels: make([]ReportPaymentChannel, 0),
		TopProducts:     make([]ReportTopProduct, 0),
		CustomerTiming: []ReportTimingPeriod{
			{Period: "morning", Label: "Morning · 6am–11am"},
			{Period: "afternoon", Label: "Afternoon · 12pm–4pm"},
			{Period: "evening", Label: "Evening · 5pm–9pm"},
			{Period: "night", Label: "Night · 10pm–5am"},
		},
	}

	type reportTotals struct {
		GrossBilled       float64
		TotalOrders       int64
		AverageOrderValue float64
	}
	var totals reportTotals
	if err := o.OrdSvc.DB.Table("orders").
		Select("COALESCE(SUM(total_amount), 0) AS gross_billed, COUNT(*) AS total_orders, COALESCE(AVG(total_amount), 0) AS average_order_value").
		Scan(&totals).Error; err != nil {
		writeReportError(w, err)
		return
	}
	report.GrossBilled = totals.GrossBilled
	report.TotalOrders = totals.TotalOrders
	report.AverageOrderValue = totals.AverageOrderValue
	if err := o.OrdSvc.DB.Table("order_items").Select("COALESCE(SUM(quantity), 0)").Scan(&report.ItemsSold).Error; err != nil {
		writeReportError(w, err)
		return
	}
	if err := o.OrdSvc.DB.Table("payments").
		Select("COALESCE(NULLIF(method, ''), 'Other') AS method, COALESCE(SUM(amount_tendered - change_given), 0) AS amount").
		Group("COALESCE(NULLIF(method, ''), 'Other')").Order("amount DESC").
		Scan(&report.PaymentChannels).Error; err != nil {
		writeReportError(w, err)
		return
	}
	if err := o.OrdSvc.DB.Table("order_items").
		Select("products.id, products.title, SUM(order_items.quantity) AS quantity, SUM(order_items.quantity * order_items.price_at_sale) AS revenue").
		Joins("JOIN products ON products.id = order_items.product_id").
		Group("products.id, products.title").Order("quantity DESC").Limit(5).
		Scan(&report.TopProducts).Error; err != nil {
		writeReportError(w, err)
		return
	}

	type timingCount struct {
		Period string
		Orders int64
	}
	var timingCounts []timingCount
	if err := o.OrdSvc.DB.Table("orders").
		Select(`CASE
			WHEN EXTRACT(HOUR FROM time_stamp) BETWEEN 6 AND 11 THEN 'morning'
			WHEN EXTRACT(HOUR FROM time_stamp) BETWEEN 12 AND 16 THEN 'afternoon'
			WHEN EXTRACT(HOUR FROM time_stamp) BETWEEN 17 AND 21 THEN 'evening'
			ELSE 'night' END AS period, COUNT(*) AS orders`).
		Group("period").Scan(&timingCounts).Error; err != nil {
		writeReportError(w, err)
		return
	}
	for _, count := range timingCounts {
		for index := range report.CustomerTiming {
			if report.CustomerTiming[index].Period == count.Period {
				report.CustomerTiming[index].Orders = count.Orders
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func writeReportError(w http.ResponseWriter, err error) {
	slog.Error("failed to build sales report", "err", err)
	http.Error(w, "failed to load sales report", http.StatusInternalServerError)
}
