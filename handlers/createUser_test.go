package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/icodeologist/poultry-backend/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}

	// Auto migrate tables
	err = db.AutoMigrate(&models.Customer{}, &models.Product{})
	if err != nil {
		panic("failed to migrate test database")
	}

	return db
}

func TestCreateNewUser(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		setupDB        func(*gorm.DB)
	}{
		{
			name: "valid customer creation",
			payload: map[string]interface{}{
				"name":         "John Doe",
				"phone_number": "1234567890",
				"balance":      100.50,
			},
			expectedStatus: http.StatusCreated,
			setupDB:        nil,
		},
		{
			name: "missing name",
			payload: map[string]interface{}{
				"phone_number": "1234567890",
				"balance":      100.50,
			},
			expectedStatus: http.StatusBadRequest,
			setupDB:        nil,
		},
		{
			name: "empty name",
			payload: map[string]interface{}{
				"name":         "",
				"phone_number": "1234567890",
				"balance":      100.50,
			},
			expectedStatus: http.StatusBadRequest,
			setupDB:        nil,
		},
		{
			name: "phone number too short",
			payload: map[string]interface{}{
				"name":         "John Doe",
				"phone_number": "12345",
				"balance":      100.50,
			},
			expectedStatus: http.StatusBadRequest,
			setupDB:        nil,
		},
		{
			name: "phone number too long",
			payload: map[string]interface{}{
				"name":         "John Doe",
				"phone_number": "12345678901234",
				"balance":      100.50,
			},
			expectedStatus: http.StatusBadRequest,
			setupDB:        nil,
		},
		{
			name: "duplicate phone number",
			payload: map[string]interface{}{
				"name":         "Jane Doe",
				"phone_number": "1234567890",
				"balance":      200.00,
			},
			expectedStatus: http.StatusConflict,
			setupDB: func(db *gorm.DB) {
				existing := models.Customer{
					Name:        "Existing User",
					PhoneNumber: "1234567890",
					Balance:     50.00,
				}
				db.Create(&existing)
			},
		},
		{
			name:           "invalid JSON",
			payload:        "not valid json{{{",
			expectedStatus: http.StatusBadRequest,
			setupDB:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh database for each test
			testDB := setupTestDB()

			// Run setup if provided
			if tt.setupDB != nil {
				tt.setupDB(testDB)
			}

			// Create handler with test DB
			handler := NewCustomerHandler(testDB)

			// Marshal payload
			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(tt.payload)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateNewUser(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// For successful creation, verify response
			if tt.expectedStatus == http.StatusCreated {
				var customer models.Customer
				if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
					t.Errorf("failed to parse response body: %v", err)
				}

				if customer.ID == 0 {
					t.Error("expected customer ID to be set")
				}

				// Verify customer was actually saved
				var saved models.Customer
				if err := testDB.First(&saved, customer.ID).Error; err != nil {
					t.Errorf("customer not found in database: %v", err)
				}
			}
		})
	}
}
