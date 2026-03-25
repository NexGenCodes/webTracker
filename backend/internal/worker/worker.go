package worker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/whatsapp"

	"go.mau.fi/whatsmeow"
)

type Worker struct {
	ID              int
	Client          *whatsmeow.Client
	Sender          *whatsapp.Sender
	ShipmentUC      *usecase.ShipmentUsecase
	ConfigUC        *usecase.ConfigUsecase
	Jobs            <-chan models.Job
	WG              *sync.WaitGroup
	Cfg             *config.Config
	AwbCmd          string
	CompanyName     string
	TrackingBaseURL string
	Cmd             *commands.Dispatcher
	ShipmentService shipment.Service
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

	// 1. Fetch Language
	langStr, _ := w.ConfigUC.GetUserLanguage(context.Background(), job.SenderJID.String())
	lang := i18n.Language(langStr)

	// A. Initial Feedback (Typing)
	w.Sender.SetTyping(job.ChatJID, true)
	defer w.Sender.SetTyping(job.ChatJID, false)

	// 2. Check for Commands
	ctx := context.WithValue(context.Background(), "jid", job.SenderJID.String())
	ctx = context.WithValue(ctx, "sender_phone", job.SenderPhone)
	ctx = context.WithValue(ctx, "is_admin", job.IsAdmin)

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

	// AI Fallback (Minimized: Only if critical fields are missing to save costs)
	if m.ReceiverName == "" || m.ReceiverPhone == "" || m.ReceiverAddress == "" {
		if aiM, err := parser.ParseAI(job.Text, w.Cfg.GeminiAPIKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
			m.Validate()
		}
	}

	// 4. Validation
	if len(m.MissingFields) > 0 {
		logger.GlobalVitals.IncParseFailure()
		logger.Warn().
			Str("jid", job.SenderJID.String()).
			Strs("missing_fields", m.MissingFields).
			Msg("Information incomplete after parsing")

		msg := "📝 *INFORMATION INCOMPLETE*\n\n━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"The system could not parse the following required fields:\n" +
			"• " + strings.Join(m.MissingFields, "\n• ") + "\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n\n_Please provide the missing data to proceed._"
		w.Sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncParseSuccess()

	orig := m.SenderCountry
	dest := m.ReceiverCountry

	newShipment := &shipment.Shipment{
		UserJID:           job.SenderJID.String(),
		Status:            shipment.StatusPending,
		SenderTimezone:    w.ShipmentService.ResolveTimezone(orig),
		RecipientTimezone: w.ShipmentService.ResolveTimezone(dest),

		SenderName:       m.SenderName,
		SenderPhone:      job.SenderPhone,
		Origin:           m.SenderCountry,
		RecipientName:    m.ReceiverName,
		RecipientPhone:   m.ReceiverPhone,
		RecipientEmail:   m.ReceiverEmail,
		RecipientID:      m.ReceiverID,
		RecipientAddress: m.ReceiverAddress,
		Destination:      m.ReceiverCountry,

		CargoType: m.CargoType,
		Weight:    15.0, // STRICT: Always 15kg as per policy
		Cost:      0.0,
	}

	if newShipment.CargoType == "" {
		newShipment.CargoType = "consignment box"
	}
	if newShipment.Origin == "" {
		newShipment.Origin = "Processing Center"
	}
	if newShipment.Destination == "" {
		newShipment.Destination = "Local Delivery"
	}

	// 5b. Deduplication Check (Strict Phone, Email, or ID Match)
	if existingID, err := w.ShipmentUC.FindSimilar(context.Background(), job.SenderJID.String(), newShipment.RecipientPhone, newShipment.RecipientEmail, newShipment.RecipientID); err == nil && existingID != "" {
		logger.Info().Str("existing_id", existingID).Msg("Duplicate shipment blocked")
		dupMsg := fmt.Sprintf("⚠️ *SHIPMENT ALREADY EXISTS*\n\nA shipment for this recipient phone is already in the system.\n\n🆔 *%s*\n\n🔹 Use `!edit %s ...` to update.\n🔹 Use `!delete %s` to remove.", existingID, existingID, existingID)
		w.Sender.Reply(job.ChatJID, job.SenderJID, dupMsg, job.MessageID, job.Text)
		return
	}

	// Generate schedule dates using the new Smart Anchor Algorithm (A & B)
	now := time.Now().UTC()
	departure := w.ShipmentService.CalculateDeparture(now, w.Cfg.AdminTimezone)
	// Enforce "Afghanistan" as sender as per user request
	arrival, outForDelivery := w.ShipmentService.CalculateArrival(departure, newShipment.Origin, newShipment.Destination)

	dbShip := &db.Shipment{
		UserJid:              newShipment.UserJID,
		Status:               sql.NullString{String: newShipment.Status, Valid: true},
		ScheduledTransitTime: sql.NullTime{Time: departure, Valid: true},
		OutfordeliveryTime:   sql.NullTime{Time: outForDelivery, Valid: true},
		ExpectedDeliveryTime: sql.NullTime{Time: arrival, Valid: true},
		SenderTimezone:       sql.NullString{String: newShipment.SenderTimezone, Valid: true},
		RecipientTimezone:    sql.NullString{String: newShipment.RecipientTimezone, Valid: true},
		SenderName:           sql.NullString{String: newShipment.SenderName, Valid: true},
		SenderPhone:          sql.NullString{String: newShipment.SenderPhone, Valid: true},
		Origin:               sql.NullString{String: newShipment.Origin, Valid: true},
		RecipientName:        sql.NullString{String: newShipment.RecipientName, Valid: true},
		RecipientPhone:       sql.NullString{String: newShipment.RecipientPhone, Valid: true},
		RecipientEmail:       sql.NullString{String: newShipment.RecipientEmail, Valid: true},
		RecipientID:          sql.NullString{String: newShipment.RecipientID, Valid: true},
		RecipientAddress:     sql.NullString{String: newShipment.RecipientAddress, Valid: true},
		Destination:          sql.NullString{String: newShipment.Destination, Valid: true},
		CargoType:            sql.NullString{String: newShipment.CargoType, Valid: true},
		Weight:               sql.NullFloat64{Float64: newShipment.Weight, Valid: true},
		Cost:                 sql.NullFloat64{Float64: newShipment.Cost, Valid: true},
	}

	trackingID, err := w.ShipmentUC.CreateWithPrefix(context.Background(), dbShip, w.Cfg.CompanyPrefix)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		logger.Error().Err(err).Str("jid", job.SenderJID.String()).Msg("Failed to insert shipment information")
		w.Sender.Reply(job.ChatJID, job.SenderJID, "❌ *SYSTEM ERROR*\n_Saving information failed. Please contact your admin._", job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 7, 8, 9. Generate and send receipt
	w.generateAndSendReceipt(job, trackingID, lang)

	logger.Info().
		Str("tracking_id", trackingID).
		Str("jid", job.SenderJID.String()).
		Msg("Shipment created successfully")

	// 10. Send tracking ID and link as follow-up message
	baseURL := w.TrackingBaseURL
	if baseURL == "" {
		baseURL = "https://web-tracker-iota.vercel.app"
	}

	trackingMsg := fmt.Sprintf("📦 *SHIPMENT INFORMATION CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\n%s?id=%s", trackingID, baseURL, trackingID)
	if m.IsAI {
		trackingMsg += "\n\n_✨ Parsed by AI_"
	}
	w.Sender.Reply(job.ChatJID, job.SenderJID, trackingMsg, job.MessageID, job.Text)

	// 11. Send Email Notification (if email provided)
	if newShipment.RecipientEmail != "" {
		trackingURL := fmt.Sprintf("%s?id=%s", baseURL, trackingID)
		notif.SendShipmentEmail(w.Cfg, newShipment, trackingURL)
	}
}

func (w *Worker) generateAndSendReceipt(job models.Job, id string, lang i18n.Language) {
	EnqueueReceipt(ReceiptJob{
		Job:         job,
		TrackingID:  id,
		Language:    lang,
		CompanyName: w.CompanyName,
		ShipmentUC:  w.ShipmentUC,
		Sender:      w.Sender,
	})
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
