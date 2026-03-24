package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/transport/http/handler"
	"webtracker-bot/internal/usecase"
	"database/sql"
	"time"
)

type Server struct {
	app        *fiber.App
	cfg        *config.Config
	shipmentUC *usecase.ShipmentUsecase
	db         *sql.DB
}

func NewServer(cfg *config.Config, shipmentUC *usecase.ShipmentUsecase, db *sql.DB) *Server {
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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))

	return &Server{
		app:        app,
		cfg:        cfg,
		shipmentUC: shipmentUC,
		db:         db,
	}
}

func (s *Server) SetupRoutes() {
	shipmentHandler := handler.NewShipmentHandler(s.shipmentUC)
	shipmentHandler.RegisterRoutes(s.app)
	
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
				"uptime":   time.Since(time.Now()).String(), // Placeholder for real uptime if needed
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
