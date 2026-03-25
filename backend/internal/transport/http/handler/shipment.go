package handler

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/utils/dbutil"

	"github.com/go-playground/validator/v10"
	"errors"
	"strings"
)

type ShipmentHandler struct {
	shipmentUC *usecase.ShipmentUsecase
	validate   *validator.Validate
}

func NewShipmentHandler(shipmentUC *usecase.ShipmentUsecase) *ShipmentHandler {
	return &ShipmentHandler{
		shipmentUC: shipmentUC,
		validate:   validator.New(),
	}
}

func (h *ShipmentHandler) RegisterRoutes(router fiber.Router) {
	// Public Routes
	router.Get("/api/track/:id", h.Track)

	// Admin Routes (Next.js protects these via NextAuth before calling Go)
	admin := router.Group("/api/admin")
	
	// Stats
	admin.Get("/stats", h.GetStats)

	// Shipments
	shipments := admin.Group("/shipments")
	shipments.Post("/", h.Create)
	shipments.Get("/", h.List)
	shipments.Delete("/cleanup", h.DeleteDelivered)
	shipments.Patch("/:id", h.UpdateStatus)
	shipments.Delete("/:id", h.Delete)
}

// Track - GET /api/track/:id
func (h *ShipmentHandler) Track(c *fiber.Ctx) error {
	id := c.Params("id")
	shipment, err := h.shipmentUC.Track(c.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
		}
		log.Printf("Track error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(shipment)
}

// CreateRequest matches Next.js CreateShipmentDto
type CreateRequest struct {
	SenderName      string  `json:"senderName" validate:"required"`
	SenderPhone     string  `json:"senderPhone" validate:"required"`
	SenderCountry   string  `json:"senderCountry" validate:"required"`
	ReceiverName    string  `json:"receiverName" validate:"required"`
	ReceiverPhone   string  `json:"receiverPhone" validate:"required"`
	ReceiverEmail   string  `json:"receiverEmail" validate:"required,email"`
	ReceiverAddress string  `json:"receiverAddress" validate:"required"`
	ReceiverCountry string  `json:"receiverCountry" validate:"required"`
	CargoType       string  `json:"cargoType" validate:"required"`
	Weight          float64 `json:"weight" validate:"required,gt=0"`
	Cost            float64 `json:"cost" validate:"required,gt=0"`
	TransitTime     int     `json:"transitTime" validate:"required,gt=0"`
}

func toNullTime(t time.Time) sql.NullTime {
	return dbutil.ToNullTime(t)
}

func toNullString(s string) sql.NullString {
	return dbutil.ToNullString(s)
}

func toNullFloat64(f float64) sql.NullFloat64 {
	return dbutil.ToNullFloat64(f)
}

// Create - POST /api/admin/shipments
func (h *ShipmentHandler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Generate a collision-resistant tracking ID
	randVal, _ := rand.Int(rand.Reader, big.NewInt(999999999))
	trackingID := fmt.Sprintf("AWB-%09d", randVal.Int64())
	now := time.Now()

	// Use industrial status calculation algorithms instead of hardcoded defaults
	departure := h.shipmentUC.Service.CalculateDeparture(now, "Africa/Lagos") // Default origin TZ
	arrival, outForDelivery := h.shipmentUC.Service.CalculateArrival(departure, req.SenderCountry, req.ReceiverCountry)

	params := db.CreateShipmentParams{
		TrackingID:           trackingID,
		UserJid:              "admin_portal", 
		Status:               toNullString("pending"),
		CreatedAt:            toNullTime(now),
		ScheduledTransitTime: toNullTime(departure),
		OutfordeliveryTime:   toNullTime(outForDelivery),
		ExpectedDeliveryTime: toNullTime(arrival),
		SenderTimezone:       toNullString("Africa/Lagos"),
		RecipientTimezone:    toNullString(h.shipmentUC.Service.ResolveTimezone(req.ReceiverCountry)),
		SenderName:           toNullString(req.SenderName),
		SenderPhone:          toNullString(req.SenderPhone),
		Origin:               toNullString(req.SenderCountry),
		RecipientName:        toNullString(req.ReceiverName),
		RecipientPhone:       toNullString(req.ReceiverPhone),
		RecipientEmail:       toNullString(req.ReceiverEmail),
		RecipientAddress:     toNullString(req.ReceiverAddress),
		Destination:          toNullString(req.ReceiverCountry),
		CargoType:            toNullString(req.CargoType),
		Weight:               toNullFloat64(req.Weight),
		Cost:                 toNullFloat64(req.Cost),
		UpdatedAt:            toNullTime(now),
	}

	if err := h.shipmentUC.Create(c.Context(), params); err != nil {
		log.Printf("Create error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"tracking_id": trackingID})
}

// List - GET /api/admin/shipments
func (h *ShipmentHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 10 }
	
	offset := (page - 1) * limit

	shipments, err := h.shipmentUC.ListPaginated(c.Context(), int32(limit), int32(offset))
	if err != nil {
		log.Printf("List error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Get real total count for pagination
	total, countErr := h.shipmentUC.CountAll(c.Context())
	if countErr != nil {
		log.Printf("Count error: %v", countErr)
		total = 0
	}

	return c.JSON(fiber.Map{
		"data": shipments,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// UpdateStatusRequest for Patch
type UpdateStatusRequest struct {
	Status      string `json:"status" validate:"required"`
	Destination string `json:"destination" validate:"required"`
}

// UpdateStatus - PATCH /api/admin/shipments/:id
func (h *ShipmentHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	if err := h.shipmentUC.UpdateStatus(c.Context(), id, req.Status, req.Destination); err != nil {
		log.Printf("Update status error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"success": true})
}

// Delete - DELETE /api/admin/shipments/:id
func (h *ShipmentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.shipmentUC.Delete(c.Context(), id); err != nil {
		log.Printf("Delete error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete shipment"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// DeleteDelivered - DELETE /api/admin/shipments/cleanup
func (h *ShipmentHandler) DeleteDelivered(c *fiber.Ctx) error {
	if err := h.shipmentUC.DeleteDelivered(c.Context()); err != nil {
		log.Printf("Cleanup error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cleanup"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// GetStats - GET /api/admin/stats
func (h *ShipmentHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.shipmentUC.CountByStatus(c.Context())
	if err != nil {
		log.Printf("Stats error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stats"})
	}
	return c.JSON(fiber.Map{
		"total":          stats.Total,
		"intransit":      stats.Intransit,
		"outfordelivery": stats.Outfordelivery,
		"delivered":      stats.Delivered,
		"pending":        stats.Pending,
		"canceled":       stats.Canceled,
	})
}
