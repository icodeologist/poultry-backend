package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

func TestAddProduct(t *testing.T) {
	tests := []struct {
		name           string
		payload        any
		expectedStatus int
		setupDb        func(*gorm.DB)
	}{
		{
			name: "valid product creation",
			payload: map[string]any{
				"title":   "chikken masala",
				"price":   1200.00,
				"instock": 112,
			},
			expectedStatus: http.StatusCreated,
			setupDb:        nil,
		},
		{
			name: "missing title error",
			payload: map[string]any{
				"price":   1200.00,
				"instock": 112,
			},
			expectedStatus: http.StatusBadRequest,
			setupDb:        nil,
		},
		{
			name: "empty title",
			payload: map[string]any{
				"title":   "",
				"price":   1200.00,
				"instock": 112,
			},
			expectedStatus: http.StatusBadRequest,
			setupDb:        nil,
		},
		{
			name: "title already reported",
			payload: map[string]any{
				"title":   "Hello world",
				"price":   1200.00,
				"instock": 112,
			},
			expectedStatus: http.StatusBadRequest,
			setupDb: func(*gorm.DB) {
				existing := models.Product{
					Title:   "Hello world",
					Price:   1200.00,
					InStock: 10,
				}
				db := setupTestDB()
				db.Create(&existing)
			},
		},
		{
			name: "invalid price",
			payload: map[string]any{
				"title":   "Chikken mystic",
				"price":   120,
				"instock": 112,
			},
			expectedStatus: http.StatusBadRequest,
			setupDb:        nil,
		},
		{
			name:           "invalid json",
			payload:        "json",
			expectedStatus: http.StatusBadRequest,
			setupDb:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := setupTestDB()
			if tt.setupDb != nil {
				tt.setupDb(testDB)
			}
			handler := NewProductHandler(testDB)

			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(tt.payload)
			}

			req := httptest.NewRequest(http.MethodPost, "/add/product", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.AddProduct(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body : %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectedStatus == http.StatusCreated {
				var product models.Product
				if err := json.Unmarshal(w.Body.Bytes(), &product); err != nil {
					t.Errorf("failed to parse response body : %v", err)
				}
				if product.ID == 0 {
					t.Error("expected product ID to be set")
				}
				var saved models.Customer
				if err := testDB.First(&saved, product.ID).Error; err != nil {
					t.Errorf("product not found in database : %v", err)
				}
			}
		})
	}
}
