package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/whatsapp"

	"go.mau.fi/whatsmeow"
)

type Worker struct {
	ID          int
	Client      *whatsmeow.Client
	Sender      *whatsapp.Sender
	DB          *supabase.Client
	Jobs        <-chan models.Job
	WG          *sync.WaitGroup
	GeminiKey   string
	AwbCmd      string
	CompanyName string
	Cmd         *commands.Dispatcher
}

func (w *Worker) Start() {
	defer w.WG.Done()
	logger.Info().Int("worker_id", w.ID).Msg("Worker started")

	for job := range w.Jobs {
		w.process(job)
	}
}

func (w *Worker) process(job models.Job) {
	logger.GlobalVitals.IncJobs()

	// 1. Check for Commands (Explicit)
	if reply, ok := w.Cmd.Dispatch(context.Background(), job.Text); ok {
		w.Sender.Reply(job.ChatJID, job.SenderJID, reply, job.MessageID)
		return
	}

	// 2. High-Performance Pre-filter
	if !w.isPotentialManifest(job.Text) {
		return
	}

	// 3. Normal Parsing (Regex first)
	m := parser.ParseRegex(job.Text)

	// AI Fallback
	if len(m.MissingFields) > 0 {
		if aiM, err := parser.ParseAI(job.Text, w.GeminiKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
			m.Validate()
		}
	}

	// 4. Validation
	if len(m.MissingFields) > 0 {
		logger.GlobalVitals.IncParseFailure()
		msg := "📝 *Information Incomplete*\n\n" +
			"The system could not parse all required fields from your message. Please provide:\n" +
			"• " + strings.Join(m.MissingFields, "\n• ")
		w.Sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID)
		return
	}
	logger.GlobalVitals.IncParseSuccess()

	// 5. Duplicate Check
	exists, tracking, err := w.DB.CheckDuplicate(m.ReceiverPhone)
	if err == nil && exists {
		logger.GlobalVitals.IncDuplicate()
		msg := fmt.Sprintf("📂 *Duplicate Record*\n\nA shipment with this phone number already exists.\nTracking ID: *%s*", tracking)
		w.Sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID)
		return
	}

	// 6. Insert
	id, err := w.DB.InsertShipment(m, job.SenderPhone)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		w.Sender.Reply(job.ChatJID, job.SenderJID, "❌ *System Error*\nSaving failed. Please contact your admin.", job.MessageID)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 7. Success
	successMsg := fmt.Sprintf("📦 *Manifest Created*\nYour shipment has been registered.\n\nID: *%s*", id)
	if m.IsAI {
		successMsg += "\n_✨ (AI Enhanced)_"
	}
	w.Sender.Reply(job.ChatJID, job.SenderJID, successMsg, job.MessageID)
}

func (w *Worker) isPotentialManifest(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"sender", "receiver", "destin", "phone", "name", "address", "country", "cargo", "ship"}
	count := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			count++
		}
	}
	return count >= 2
}
