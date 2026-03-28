package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/transport/http"
	"webtracker-bot/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPI_Deep(t *testing.T) {
	// Setup dependencies
	cfg := &config.Config{
		TrackingBaseURL: "http://localhost:3000",
		APISecretKey:    "test-secret",
	}

	repo := new(MockQuerier)
	uc := usecase.NewShipmentUsecase(repo, nil)

	// Initialize Fiber App via Server
	server := http.NewServer(cfg, uc, nil, nil)
	server.SetupRoutes()
	app := server.GetAppForTest() // I need to add this helper to server.go or use field if public

	t.Run("GetShipment_Found", func(t *testing.T) {
		repo.On("GetShipment", mock.Anything, "AWB-123").Return(db.Shipment{
			TrackingID: "AWB-123",
			Status:     sql.NullString{String: "pending", Valid: true},
		}, nil).Once()

		req := httptest.NewRequest("GET", "/api/shipments/AWB-123", nil)
		req.Header.Set("X-API-Key", "test-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "AWB-123", body["tracking_id"])
		repo.AssertExpectations(t)
	})

	t.Run("GetShipment_NotFound", func(t *testing.T) {
		repo.On("GetShipment", mock.Anything, "MISSING").Return(db.Shipment{}, sql.ErrNoRows).Once()

		req := httptest.NewRequest("GET", "/api/shipments/MISSING", nil)
		req.Header.Set("X-API-Key", "test-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 404, resp.StatusCode)
		repo.AssertExpectations(t)
	})

	t.Run("Auth_Failure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/shipments/AWB-123", nil)
		req.Header.Set("X-API-Key", "wrong-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("CreateShipment_BrutalValidation", func(t *testing.T) {
		// Validating malformed JSON
		req := httptest.NewRequest("POST", "/api/shipments", bytes.NewBufferString("{invalid-json}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 400, resp.StatusCode)
	})
}
