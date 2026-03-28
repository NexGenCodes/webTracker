package receipt

import (
	"context"
	"sync"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
)

// Job represents a receipt rendering job.
type Job struct {
	Msg         models.Job
	TrackingID  string
	Language    i18n.Language
	CompanyName string
	ShipmentUC  *usecase.ShipmentUsecase
	Sender      *whatsapp.Sender
	RenderMode  string // "legacy" or "default"
}

var (
	queue chan Job
	once  sync.Once
)

const QueueSize = 100

// InitProcessor initializes the singleton queue and starts the worker goroutines.
func InitProcessor(companyName string, shipUC *usecase.ShipmentUsecase, sender *whatsapp.Sender) {
	once.Do(func() {
		queue = make(chan Job, QueueSize)
		go startWorker(companyName, shipUC, sender)
		go startWorker(companyName, shipUC, sender)
	})
}

// Enqueue adds a receipt rendering job to the queue.
func Enqueue(job Job) {
	if queue == nil {
		logger.Error().Msg("Receipt processor not initialized")
		return
	}
	select {
	case queue <- job:
		logger.Debug().Str("tracking_id", job.TrackingID).Msg("Receipt enqueued")
	default:
		logger.Warn().Str("tracking_id", job.TrackingID).Msg("Receipt queue full, dropping job")
	}
}

func startWorker(companyName string, shipUC *usecase.ShipmentUsecase, sender *whatsapp.Sender) {
	logger.Info().Msg("Singleton Receipt Processor started (Optimized mode)")
	for rJob := range queue {
		processReceipt(rJob)
	}
}

func processReceipt(rj Job) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error().Msgf("Receipt processor panicked: %v", r)
		}
	}()

	logger.Info().Str("id", rj.TrackingID).Msg("Rendering receipt (singleton queue)")

	if rj.TrackingID == "" {
		logger.Error().Msg("Cannot process receipt: tracking ID is empty")
		return
	}

	// 1. Fetch Shipment
	dbShip, err := rj.ShipmentUC.Track(context.Background(), rj.TrackingID)
	if err != nil || dbShip == nil {
		logger.Warn().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to fetch info for receipt delivery")
		return
	}

	// 2. Map to Domain Model for Rendering
	s := &shipment.Shipment{
		TrackingID:        dbShip.TrackingID,
		UserJID:           dbShip.UserJid,
		Status:            dbShip.Status.String,
		CreatedAt:         dbShip.CreatedAt.Time,
		SenderTimezone:    dbShip.SenderTimezone.String,
		RecipientTimezone: dbShip.RecipientTimezone.String,
		SenderName:        dbShip.SenderName.String,
		SenderPhone:       dbShip.SenderPhone.String,
		Origin:            dbShip.Origin.String,
		RecipientName:     dbShip.RecipientName.String,
		RecipientPhone:    dbShip.RecipientPhone.String,
		RecipientID:       dbShip.RecipientID.String,
		RecipientEmail:    dbShip.RecipientEmail.String,
		RecipientAddress:  dbShip.RecipientAddress.String,
		Destination:       dbShip.Destination.String,
		CargoType:         dbShip.CargoType.String,
		Weight:            dbShip.Weight.Float64,
		Cost:              dbShip.Cost.Float64,
	}

	if dbShip.ScheduledTransitTime.Valid {
		s.ScheduledTransitTime = &dbShip.ScheduledTransitTime.Time
	}
	if dbShip.OutfordeliveryTime.Valid {
		s.OutForDeliveryTime = &dbShip.OutfordeliveryTime.Time
	}
	if dbShip.ExpectedDeliveryTime.Valid {
		s.ExpectedDeliveryTime = &dbShip.ExpectedDeliveryTime.Time
	}

	// 3. Render (Memory Intensive Step - Only one at a time)
	receiptImg, err := utils.RenderReceipt(*s, rj.CompanyName, rj.Language)
	if err != nil {
		logger.Error().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to render receipt")
		return
	}

	// 4. Send
	err = rj.Sender.SendImage(rj.Msg.ChatJID, rj.Msg.SenderJID, receiptImg, "", rj.Msg.MessageID, rj.Msg.Text)
	if err != nil {
		logger.Warn().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to deliver receipt image")
	}

	logger.Info().Str("id", rj.TrackingID).Msg("Receipt processing complete")
}
