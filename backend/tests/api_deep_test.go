package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/shipment"
	httpapi "webtracker-bot/internal/api"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPI_Deep(t *testing.T) {
	// Setup dependencies
	cfg := &config.Config{
		TrackingBaseURL: "http://localhost:3000",
		APISecretKey:    "test-secret",
	}

	setupApp := func() (*fiber.App, *MockQuerier) {
		repo := new(MockQuerier)
		uc := shipment.NewUsecase(repo, nil)
		server := httpapi.NewServer(cfg, uc, nil, nil, nil)
		server.SetupRoutes()
		return server.GetAppForTest(), repo
	}

	t.Run("GetShipment_Found", func(t *testing.T) {
		app, repo := setupApp()
		repo.On("GetShipment", mock.Anything, mock.AnythingOfType("db.GetShipmentParams")).Return(db.Shipment{
			TrackingID: "AWB-123",
			Status:     sql.NullString{String: "pending", Valid: true},
		}, nil).Once()

		req := httptest.NewRequest("GET", "/api/track/AWB-123?company_id="+testCompanyID.String(), nil)
		req.Header.Set("X-API-Key", "test-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "AWB-123", body["tracking_id"])
		repo.AssertExpectations(t)
	})

	t.Run("GetShipment_NotFound", func(t *testing.T) {
		app, repo := setupApp()
		repo.On("GetShipment", mock.Anything, mock.AnythingOfType("db.GetShipmentParams")).Return(db.Shipment{}, sql.ErrNoRows).Once()

		req := httptest.NewRequest("GET", "/api/track/MISSING?company_id="+testCompanyID.String(), nil)
		req.Header.Set("X-API-Key", "test-secret")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
		repo.AssertExpectations(t)
	})

	t.Run("Auth_Failure", func(t *testing.T) {
		app, _ := setupApp()
		req := httptest.NewRequest("GET", "/api/admin/shipments/?company_id="+testCompanyID.String(), nil)
		req.Header.Set("X-API-Key", "wrong-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("CreateShipment_BrutalValidation", func(t *testing.T) {
		app, _ := setupApp()
		// Validating malformed JSON
		req := httptest.NewRequest("POST", "/api/admin/shipments/?company_id="+testCompanyID.String(), bytes.NewBufferString("{invalid-json}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-secret")

		resp, _ := app.Test(req)
		assert.Equal(t, 400, resp.StatusCode)
	})
}
