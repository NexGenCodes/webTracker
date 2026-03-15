package worker

import (
	"context"
	"sync"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
)

type ReceiptJob struct {
	Job         models.Job
	TrackingID  string
	Language    i18n.Language
	CompanyName string
	LocalDB     *localdb.Client
	Sender      *whatsapp.Sender
}

var (
	receiptQueue chan ReceiptJob
	once         sync.Once
)

const (
	QueueSize = 100
)

// InitReceiptProcessor initializes the singleton queue and starts the worker.
func InitReceiptProcessor(companyName string, ldb *localdb.Client, sender *whatsapp.Sender) {
	once.Do(func() {
		receiptQueue = make(chan ReceiptJob, QueueSize)
		go startWorker(companyName, ldb, sender)
	})
}

// EnqueueReceipt adds a receipt rendering job to the queue.
func EnqueueReceipt(job ReceiptJob) {
	if receiptQueue == nil {
		logger.Error().Msg("Receipt processor not initialized")
		return
	}
	select {
	case receiptQueue <- job:
		logger.Debug().Str("tracking_id", job.TrackingID).Msg("Receipt enqueued")
	default:
		logger.Warn().Str("tracking_id", job.TrackingID).Msg("Receipt queue full, dropping job")
	}
}

func startWorker(companyName string, ldb *localdb.Client, sender *whatsapp.Sender) {
	logger.Info().Msg("Singleton Receipt Processor started (Concurrency: 1)")
	for rJob := range receiptQueue {
		processReceipt(rJob)
	}
}

func processReceipt(rj ReceiptJob) {
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
	s, err := rj.LocalDB.GetShipment(context.Background(), rj.TrackingID)
	if err != nil || s == nil {
		logger.Warn().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to fetch info for receipt delivery")
		return
	}

	// 2. Render (Memory Intensive Step - Only one at a time)
	receiptImg, err := utils.RenderReceipt(*s, rj.CompanyName, rj.Language)
	if err != nil {
		logger.Error().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to render receipt")
		return
	}

	// 3. Send
	err = rj.Sender.SendImage(rj.Job.ChatJID, rj.Job.SenderJID, receiptImg, "", rj.Job.MessageID, rj.Job.Text)
	if err != nil {
		logger.Warn().Err(err).Str("tracking_id", rj.TrackingID).Msg("Failed to deliver receipt image")
	}

	logger.Info().Str("id", rj.TrackingID).Msg("Receipt processing complete")
}
