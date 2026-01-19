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
	"webtracker-bot/internal/utils"
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
		w.Sender.Reply(job.ChatJID, job.SenderJID, reply, job.MessageID, job.Text)
		return
	}

	// 2. High-Performance Pre-filter
	isManifest, isPartial := w.isPotentialManifest(job.Text)
	if !isManifest {
		if isPartial {
			hint := "💡 *IT LOOKS LIKE A SHIPMENT!*\n\n" +
				"━━━━━━━━━━━━━━━━━━━━━━━\n" +
				"To register a package, please ensure your message includes:\n" +
				"• Sender Name\n" +
				"• Receiver Name\n" +
				"• Receiver Phone\n" +
				"━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
				"_Type `!help` for a full example._"
			w.Sender.Reply(job.ChatJID, job.SenderJID, hint, job.MessageID, job.Text)
		}
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
		msg := "📝 *INFORMATION INCOMPLETE*\n\n━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"The system could not parse the following required fields:\n" +
			"• " + strings.Join(m.MissingFields, "\n• ") + "\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n\n_Please provide the missing data to proceed._"
		w.Sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncParseSuccess()

	// 5. Duplicate Check
	exists, tracking, err := w.DB.CheckDuplicate(context.Background(), m.ReceiverPhone)
	if err == nil && exists {
		logger.GlobalVitals.IncDuplicate()
		msg := fmt.Sprintf("📂 *DUPLICATE RECORD*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n_A shipment with this phone number already exists._", tracking)
		w.Sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID, job.Text)
		return
	}

	// 6. Insert
	id, err := w.DB.InsertShipment(context.Background(), m, job.SenderPhone)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		w.Sender.Reply(job.ChatJID, job.SenderJID, "❌ *SYSTEM ERROR*\n_Saving failed. Please contact your admin._", job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 7. Fetch full shipment for receipt
	shipment, err := w.DB.GetShipment(context.Background(), id)
	if err != nil || shipment == nil {
		logger.Warn().Err(err).Str("tracking_id", id).Msg("Failed to fetch shipment for receipt")
		// Fallback to text-only success
		textMsg := fmt.Sprintf("📦 *PACKAGE SHIPPING CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
		if m.IsAI {
			textMsg += "\n\n_✨ Parsed by AI_"
		}
		w.Sender.Reply(job.ChatJID, job.SenderJID, textMsg, job.MessageID, job.Text)
		return
	}

	// 8. Generate receipt image
	receiptImg, err := utils.RenderReceipt(*shipment)
	if err != nil {
		logger.Warn().Err(err).Str("tracking_id", id).Msg("Receipt render failed, sending text")
		// Fallback to text-only success
		textMsg := fmt.Sprintf("📦 *PACKAGE SHIPPING CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
		if m.IsAI {
			textMsg += "\n\n_✨ Parsed by AI_"
		}
		w.Sender.Reply(job.ChatJID, job.SenderJID, textMsg, job.MessageID, job.Text)
		return
	}

	// 9. Send receipt image (quoted reply to manifest)
	caption := "📦 *PACKAGE SHIPPING CREATED*"
	if m.IsAI {
		caption += "\n\n_✨ Parsed by AI_"
	}

	err = w.Sender.SendImage(job.ChatJID, job.SenderJID, receiptImg, caption, job.MessageID, job.Text)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to send receipt image, sending text fallback")
		// Fallback to text if image send fails
		textMsg := fmt.Sprintf("📦 *PACKAGE SHIPPING CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
		if m.IsAI {
			textMsg += "\n\n_✨ Parsed by AI_"
		}
		w.Sender.Reply(job.ChatJID, job.SenderJID, textMsg, job.MessageID, job.Text)
		return
	}

	// 10. Send tracking ID as follow-up message (Group Chat) quote the user's message
	trackingMsg := fmt.Sprintf("🔗 *Tracking ID:* %s\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
	w.Sender.Reply(job.ChatJID, job.SenderJID, trackingMsg, job.MessageID, job.Text)
}

func (w *Worker) isPotentialManifest(text string) (bool, bool) {
	lower := strings.ToLower(text)

	// Sender Check
	hasSender := strings.Contains(lower, "sender") || strings.Contains(lower, "origin") || strings.Contains(lower, "from")

	// Receiver Variants Check
	hasReceiver := false
	receiverKeywords := []string{"receiver", "reciver", "receive", "recieve", "resiver", "recever", "receivers", "recievers", "reciever"}
	for _, kw := range receiverKeywords {
		if strings.Contains(lower, kw) {
			hasReceiver = true
			break
		}
	}

	// Phone Variants Check
	hasPhone := false
	phoneKeywords := []string{"phone", "mobile", "tel", "num", "contact", "telephone", "mobil", "number"}
	for _, kw := range phoneKeywords {
		if strings.Contains(lower, kw) {
			hasPhone = true
			break
		}
	}

	// Name Check
	hasName := strings.Contains(lower, "name")

	// Strict: All 4
	if hasSender && hasReceiver && hasPhone && hasName {
		return true, false
	}

	// Partial: At least 3
	count := 0
	if hasSender {
		count++
	}
	if hasReceiver {
		count++
	}
	if hasPhone {
		count++
	}
	if hasName {
		count++
	}

	return false, count >= 3
}
