package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"webtracker-bot/internal/auth"
	"webtracker-bot/internal/cache"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthApp() (*fiber.App, *MockQuerier) {
	cfg := &config.Config{
		FrontendURL: os.Getenv("FRONTEND_URL"),
		JWTSecret:   "test-secret",
	}

	repo := new(MockQuerier)
	authService := auth.NewService(cfg, repo)
	authHandler := auth.NewHandler(authService)

	// Prevent nil pointer panic in tests
	cache.RedisClient = redis.NewClient(&redis.Options{Addr: "localhost:1"})

	app := fiber.New()
	authHandler.RegisterRoutes(app)
	return app, repo
}

func TestAuthAPI_RegisterIntent(t *testing.T) {
	app, repo := setupAuthApp()

	reqBody := []byte(`{"company_name":"Test Inc", "email":"test@example.com", "password":"Password123!"}`)

	// Mock the user not existing
	repo.On("GetCompanyByEmail", mock.Anything, "test@example.com").Return(db.Company{}, sql.ErrNoRows).Once()

	req := httptest.NewRequest("POST", "/api/auth/register-intent", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	t.Logf("Response Status: %d, Body: %v", resp.StatusCode, body)
	
	assert.Equal(t, 500, resp.StatusCode)
	// Because Redis is not available, it fails to store the pending user
	repo.AssertExpectations(t)
}

func TestAuthAPI_LoginFailure(t *testing.T) {
	app, repo := setupAuthApp()

	reqBody := []byte(`{"email":"nonexistent@example.com", "password":"Password123!"}`)

	repo.On("GetCompanyByEmail", mock.Anything, "nonexistent@example.com").Return(db.Company{}, sql.ErrNoRows).Once()

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode) // Unauthorized
	repo.AssertExpectations(t)
}
