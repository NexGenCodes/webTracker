package api

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"database/sql"
	"time"
	"webtracker-bot/internal/auth"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/shipment"
)

type Server struct {
	app        *fiber.App
	cfg        *config.Config
	shipmentUC *shipment.Usecase
	configUC   *config.Usecase
	db         *sql.DB
	bots       models.BotProvider
	startTime  time.Time
}

func NewServer(cfg *config.Config, shipmentUC *shipment.Usecase, configUC *config.Usecase, db *sql.DB, bots models.BotProvider) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             1 * 1024 * 1024, // 1MB — prevents OOM on free tier
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
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
	app.Use(recover.New()) // Catch panics — prevents full process crash
	app.Use(logger.New())

	// Safely restrict CORS to the configured frontend URL and local development
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == cfg.FrontendURL || origin == "http://localhost:3000" {
				return true
			}
			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With, Cache-Control, Pragma, X-OTP-Token, X-Reset-Token, X-Company-ID",
		AllowMethods:     "GET, POST, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Global JWT Authentication (Zero-Trust Token Relay)
	if cfg.JWTPublicKeyPath != "" {
		app.Use(auth.JWTAuth(cfg.JWTPublicKeyPath))
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
	// Initialize auth service and handler
	queries := db.New(s.db)
	authService := auth.NewService(s.cfg, queries)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(s.app)

	shipmentHandler := NewShipmentHandler(s.shipmentUC, s.configUC, s.cfg, s.bots)
	shipmentHandler.RegisterRoutes(s.app)

	companyHandler := NewCompanyHandler(s.cfg, s.configUC, s.bots)
	companyHandler.RegisterRoutes(s.app)

	billingHandler := NewBillingHandler(s.cfg, s.configUC)
	billingHandler.RegisterRoutes(s.app)

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

