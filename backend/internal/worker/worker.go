package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/localdb"
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
	LocalDB     *localdb.Client
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
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error().Msgf("Worker %d panicked: %v", w.ID, r)
				}
			}()
			w.process(job)
		}()
	}
}

func (w *Worker) process(job models.Job) {
	logger.GlobalVitals.IncJobs()

	// 1. Fetch Language (Moved from event listener to worker for concurrency)
	langStr, _ := w.LocalDB.GetUserLanguage(context.Background(), job.SenderJID.String())
	lang := i18n.Language(langStr)

	// 2. Check for Commands (Explicit)
	ctx := context.WithValue(context.Background(), "jid", job.SenderJID.String())
	if res, ok := w.Cmd.Dispatch(ctx, job.Text); ok {
		w.Sender.Reply(job.ChatJID, job.SenderJID, res.Message, job.MessageID, job.Text)

		// If it was an edit, we need to regenerate the receipt
		if res.EditID != "" {
			logger.Info().Str("edit_id", res.EditID).Msg("Edit detected, triggering receipt regeneration")
			w.generateAndSendReceipt(job, res.EditID, lang)
		}
		return
	}

	// 2. High-Performance Pre-filter
	isManifest, isPartial := w.isPotentialManifest(job.Text)
	if !isManifest {
		if isPartial {
			hint := "💡 *IT LOOKS LIKE SHIPMENT INFORMATION!*\n\n" +
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

	// 5. Duplicate Check & ID Retrieval
	exists, tracking, err := w.DB.CheckDuplicate(context.Background(), m.ReceiverPhone)
	id := tracking

	if err == nil && exists {
		logger.GlobalVitals.IncDuplicate()
		logger.Info().Str("tracking_id", id).Str("jid", job.SenderJID.String()).Msg("Duplicate record found, skipping image generation")

		// For duplicates, ONLY send a text confirmation to save CPU/Network
		trackingMsg := fmt.Sprintf("📂 *DUPLICATE SHIPMENT INFORMATION*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
		w.Sender.Reply(job.ChatJID, job.SenderJID, trackingMsg, job.MessageID, job.Text)
		return
	}

	// 6. Insert New
	id, err = w.DB.InsertShipment(context.Background(), m, job.SenderPhone)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		logger.Error().Err(err).Str("jid", job.SenderJID.String()).Msg("Failed to insert shipment information")
		w.Sender.Reply(job.ChatJID, job.SenderJID, "❌ *SYSTEM ERROR*\n_Saving information failed. Please contact your admin._", job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 7, 8, 9. Generate and send receipt
	w.generateAndSendReceipt(job, id, lang)

	// 10. Send tracking ID and link as follow-up message

	// 10. Send tracking ID and link as follow-up message
	trackingMsg := fmt.Sprintf("📦 *SHIPMENT INFORMATION CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\nhttps://web-tracker-iota.vercel.app?id=%s", id, id)
	if m.IsAI {
		trackingMsg += "\n\n_✨ Parsed by AI_"
	}
	w.Sender.Reply(job.ChatJID, job.SenderJID, trackingMsg, job.MessageID, job.Text)
}

func (w *Worker) generateAndSendReceipt(job models.Job, id string, lang i18n.Language) {
	// 1. Fetch
	shipment, err := w.DB.GetShipment(context.Background(), id)
	if err != nil || shipment == nil {
		logger.Warn().Err(err).Str("tracking_id", id).Msg("Failed to fetch info for receipt delivery")
		return
	}

	// 2. Render
	receiptImg, err := utils.RenderReceipt(*shipment, w.CompanyName, lang)
	if err != nil {
		logger.Error().Err(err).Str("tracking_id", id).Msg("Failed to render updated receipt")
		return
	}

	// 3. Send
	err = w.Sender.SendImage(job.ChatJID, job.SenderJID, receiptImg, "", job.MessageID, job.Text)
	if err != nil {
		logger.Warn().Err(err).Str("tracking_id", id).Msg("Failed to deliver receipt image")
	}
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
