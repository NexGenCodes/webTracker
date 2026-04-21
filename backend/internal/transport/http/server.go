package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"database/sql"
	"strings"
	"time"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/transport/http/handler"
	"webtracker-bot/internal/transport/http/middleware"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/whatsapp"
)

type Server struct {
	app        *fiber.App
	cfg        *config.Config
	shipmentUC *usecase.ShipmentUsecase
	configUC   *usecase.ConfigUsecase
	db         *sql.DB
	bots       whatsapp.BotProvider
	startTime  time.Time
}

func NewServer(cfg *config.Config, shipmentUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, db *sql.DB, bots whatsapp.BotProvider) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Global Middlewares
	app.Use(logger.New())

	// Dynamic CORS Origins based on Config (Production URL + Localhost)
	allowOrigins := cfg.TrackingBaseURL
	if !strings.Contains(allowOrigins, "localhost") {
		allowOrigins += ", http://localhost:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Company-ID",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))

	// API Key Authentication (skip if no key is configured, e.g. local dev)
	if cfg.APISecretKey != "" {
		app.Use(middleware.APIKeyAuth(cfg.APISecretKey))
	}

	return &Server{
		app:        app,
		cfg:        cfg,
		shipmentUC: shipmentUC,
		configUC:   configUC,
		db:         db,
		bots:       bots,
		startTime:  time.Now(),
	}
}

func (s *Server) SetupRoutes() {
	shipmentHandler := handler.NewShipmentHandler(s.shipmentUC, s.cfg, s.bots)
	shipmentHandler.RegisterRoutes(s.app)

	companyHandler := handler.NewCompanyHandler(s.cfg, s.configUC, s.bots)
	companyHandler.RegisterRoutes(s.app)

	// Enhanced Healthcheck
	s.app.Get("/health", func(c *fiber.Ctx) error {
		status := "OK"
		dbStatus := "connected"

		if s.db != nil {
			if err := s.db.Ping(); err != nil {
				status = "Error"
				dbStatus = "disconnected: " + err.Error()
			}
		} else {
			status = "Warning"
			dbStatus = "not initialized"
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    status,
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"database": dbStatus,
				"uptime":   time.Since(s.startTime).Truncate(time.Second).String(),
			},
		})
	})
}

func (s *Server) Start(port string) error {
	s.SetupRoutes()
	log.Printf("Starting HTTP REST API on port %s", port)
	return s.app.Listen(":" + port)
}

func (s *Server) Stop() error {
	log.Println("Stopping HTTP REST API...")
	return s.app.Shutdown()
}

// GetAppForTest returns the Fiber app for internal testing.
func (s *Server) GetAppForTest() *fiber.App {
	return s.app
}
