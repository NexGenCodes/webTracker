package handler

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/utils/dbutil"

	"errors"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/whatsapp"

	"github.com/go-playground/validator/v10"
)

type ShipmentHandler struct {
	shipmentUC *usecase.ShipmentUsecase
	validate   *validator.Validate
	cfg        *config.Config
	bots       whatsapp.BotProvider
}

// NewShipmentHandler injects the Usecase
func NewShipmentHandler(shipmentUC *usecase.ShipmentUsecase, cfg *config.Config, bots whatsapp.BotProvider) *ShipmentHandler {
	return &ShipmentHandler{
		shipmentUC: shipmentUC,
		validate:   validator.New(),
		cfg:        cfg,
		bots:       bots,
	}
}

func getCompanyID(c *fiber.Ctx) uuid.UUID {
	idStr := c.Get("X-Company-ID")
	if idStr == "" {
		idStr = c.Query("company_id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func (h *ShipmentHandler) RegisterRoutes(router fiber.Router) {
	// Public Routes
	router.Get("/api/track/:id", h.Track)

	// Admin Routes (Next.js protects these via NextAuth before calling Go)
	admin := router.Group("/api/admin")

	// Stats & Telemetry
	admin.Get("/stats", h.GetStats)
	admin.Get("/telemetry", h.GetTelemetry)

	// Shipments
	shipments := admin.Group("/shipments")
	shipments.Post("/parse", h.ParseText)
	shipments.Post("/bulk_csv", h.BulkCreateCSV)
	shipments.Post("/", h.Create)
	shipments.Get("/", h.List)
	shipments.Delete("/cleanup", h.DeleteDelivered)
	shipments.Patch("/bulk_status", h.BulkUpdateStatus)
	shipments.Delete("/bulk_delete", h.BulkDelete)
	shipments.Patch("/:id", h.UpdateStatus)
	shipments.Delete("/:id", h.Delete)

}

// Track - GET /api/track/:id
func (h *ShipmentHandler) Track(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	shipment, err := h.shipmentUC.Track(c.Context(), companyID, id)
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
	SenderPhone     string  `json:"senderPhone"`
	SenderCountry   string  `json:"senderCountry" validate:"required"`
	ReceiverName    string  `json:"receiverName" validate:"required"`
	ReceiverPhone   string  `json:"receiverPhone" validate:"required"`
	ReceiverEmail   string  `json:"receiverEmail" validate:"required,email"`
	ReceiverAddress string  `json:"receiverAddress" validate:"required"`
	ReceiverCountry string  `json:"receiverCountry" validate:"required"`
	CargoType       string  `json:"cargoType"`
	Weight          float64 `json:"weight" validate:"required,gt=0"`
	Cost            float64 `json:"cost"`
	TransitTime     int     `json:"transitTime"`
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
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.validate.Struct(req); err != nil {
		var errs []string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				errs = append(errs, fmt.Sprintf("'%s' is %s", e.Field(), e.Tag()))
			}
		} else {
			errs = append(errs, err.Error())
		}

		detailedErr := strings.Join(errs, ", ")
		log.Printf("Create shipment validation error: %v", detailedErr)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed: missing or invalid field(s) - " + detailedErr,
			"details": detailedErr,
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

	if err := h.shipmentUC.Create(c.Context(), companyID, params); err != nil {
		log.Printf("Create shipment error: %v", err)
		h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_create_fail", []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment"})
	}

	h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_create_success", []byte(fmt.Sprintf(`{"id": "%s"}`, trackingID)))
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"tracking_id": trackingID})
}

// List - GET /api/admin/shipments
func (h *ShipmentHandler) List(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	shipments, err := h.shipmentUC.ListPaginated(c.Context(), companyID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("List error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Get real total count for pagination
	total, countErr := h.shipmentUC.CountAll(c.Context(), companyID)
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
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

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

	// Fetch to capture User JID and Email before passing to notifier
	ship, err := h.shipmentUC.Track(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	if err := h.shipmentUC.UpdateStatus(c.Context(), companyID, id, req.Status, req.Destination); err != nil {
		log.Printf("Update status error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}

	// Send instant alert for manual admin overrides
	if h.bots != nil {
		if bot, err := h.bots.GetBot(companyID); err == nil {
			notif.SendStatusAlert(c.Context(), bot.WA, h.cfg, bot.CompanyName, ship.UserJid, ship.TrackingID, req.Status, ship.RecipientEmail.String)
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

// Delete - DELETE /api/admin/shipments/:id
func (h *ShipmentHandler) Delete(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	id := c.Params("id")
	if err := h.shipmentUC.Delete(c.Context(), companyID, id); err != nil {
		log.Printf("Delete error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete shipment"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// DeleteDelivered - DELETE /api/admin/shipments/cleanup
func (h *ShipmentHandler) DeleteDelivered(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	if err := h.shipmentUC.DeleteDelivered(c.Context(), companyID); err != nil {
		log.Printf("Cleanup error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to cleanup"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// GetStats - GET /api/admin/stats
func (h *ShipmentHandler) GetStats(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	stats, err := h.shipmentUC.CountByStatus(c.Context(), companyID)
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

// ParseRequest for ParseText endpoint
type ParseRequest struct {
	Text string `json:"text" validate:"required"`
}

// ParseText - POST /api/admin/shipments/parse
func (h *ShipmentHandler) ParseText(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req ParseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// 1. Regex Parse
	m := parser.ParseRegex(req.Text)

	// 2. AI Fallback Parse
	if m.ReceiverName == "" || m.ReceiverPhone == "" || m.ReceiverAddress == "" {
		if aiM, err := parser.ParseAI(req.Text, h.cfg.GeminiAPIKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
			m.Validate()
			h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_parse_ai", []byte(fmt.Sprintf(`{"text_len": %d}`, len(req.Text))))
		} else {
			log.Printf("AI Parse error: %v", err)
			h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_parse_fail", []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		}
	} else {
		h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_parse_regex", []byte(fmt.Sprintf(`{"text_len": %d}`, len(req.Text))))
	}

	return c.JSON(m)
}

// BulkCreateCSV - POST /api/admin/shipments/bulk_csv
func (h *ShipmentHandler) BulkCreateCSV(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req ParseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	manifests, err := parser.ParseCSV(req.Text)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if len(manifests) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No shipments found in CSV"})
	}

	createdIds := []string{}
	failed := 0

	for _, m := range manifests {
		randVal, _ := rand.Int(rand.Reader, big.NewInt(999999999))
		trackingID := fmt.Sprintf("AWB-%09d", randVal.Int64())
		now := time.Now()

		departure := h.shipmentUC.Service.CalculateDeparture(now, "Africa/Lagos")
		arrival, outForDelivery := h.shipmentUC.Service.CalculateArrival(departure, m.SenderCountry, m.ReceiverCountry)

		if err := h.shipmentUC.Create(c.Context(), companyID, db.CreateShipmentParams{
			TrackingID:           trackingID,
			UserJid:              "admin_portal",
			Status:               toNullString("pending"),
			CreatedAt:            toNullTime(now),
			ScheduledTransitTime: toNullTime(departure),
			OutfordeliveryTime:   toNullTime(outForDelivery),
			ExpectedDeliveryTime: toNullTime(arrival),
			SenderTimezone:       toNullString("Africa/Lagos"),
			RecipientTimezone:    toNullString(h.shipmentUC.Service.ResolveTimezone(m.ReceiverCountry)),
			SenderName:           toNullString(m.SenderName),
			Origin:               toNullString(m.SenderCountry),
			RecipientName:        toNullString(m.ReceiverName),
			RecipientPhone:       toNullString(m.ReceiverPhone),
			RecipientEmail:       toNullString(m.ReceiverEmail),
			RecipientAddress:     toNullString(m.ReceiverAddress),
			Destination:          toNullString(m.ReceiverCountry),
			CargoType:            toNullString(m.CargoType),
			Weight:               toNullFloat64(m.Weight),
			UpdatedAt:            toNullTime(now),
		}); err != nil {
			log.Printf("Bulk create error: %v", err)
			failed++
		} else {
			createdIds = append(createdIds, trackingID)
		}
	}

	h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_bulk_csv", []byte(fmt.Sprintf(`{"created": %d, "failed": %d}`, len(createdIds), failed)))

	return c.JSON(fiber.Map{
		"success": true,
		"created": len(createdIds),
		"failed":  failed,
		"ids":     createdIds,
	})
}

// GetTelemetry - GET /api/admin/telemetry
func (h *ShipmentHandler) GetTelemetry(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	stats, err := h.shipmentUC.GetTelemetryStats(c.Context(), companyID, time.Now().Add(-24*7*time.Hour))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	recent, _ := h.shipmentUC.GetRecentEvents(c.Context(), companyID, 50)

	return c.JSON(fiber.Map{
		"stats":  stats,
		"recent": recent,
	})
}

// BulkUpdateStatusRequest for BulkUpdateStatus endpoint
type BulkUpdateStatusRequest struct {
	IDs    []string `json:"ids" validate:"required,min=1"`
	Status string   `json:"status" validate:"required"`
}

// BulkUpdateStatus - PATCH /api/admin/shipments/bulk_status
func (h *ShipmentHandler) BulkUpdateStatus(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req BulkUpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.shipmentUC.BulkUpdateStatus(c.Context(), companyID, req.IDs, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_bulk_update", []byte(fmt.Sprintf(`{"count": %d, "status": "%s"}`, len(req.IDs), req.Status)))

	// Send instant alerts for manual admin bulk overrides
	if h.bots != nil {
		if bot, err := h.bots.GetBot(companyID); err == nil {
			for _, id := range req.IDs {
				// Ignore track errors in bulk sending (if one fails, don't halt everything)
				if ship, err := h.shipmentUC.Track(c.Context(), companyID, id); err == nil {
					notif.SendStatusAlert(c.Context(), bot.WA, h.cfg, bot.CompanyName, ship.UserJid, ship.TrackingID, req.Status, ship.RecipientEmail.String)
				}
			}
		}
	}

	return c.JSON(fiber.Map{"success": true, "updated": len(req.IDs)})
}

// BulkDeleteRequest for BulkDelete endpoint
type BulkDeleteRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// BulkDelete - DELETE /api/admin/shipments/bulk_delete
func (h *ShipmentHandler) BulkDelete(c *fiber.Ctx) error {
	companyID := getCompanyID(c)
	if companyID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid company_id"})
	}

	var req BulkDeleteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	for _, id := range req.IDs {
		_ = h.shipmentUC.Delete(c.Context(), companyID, id)
	}

	h.shipmentUC.RecordEvent(c.Context(), companyID, "admin_bulk_delete", []byte(fmt.Sprintf(`{"count": %d}`, len(req.IDs))))

	return c.JSON(fiber.Map{"success": true, "deleted": len(req.IDs)})
}
