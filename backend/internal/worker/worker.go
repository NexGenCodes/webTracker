package worker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"


	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"

)

type Worker struct {
	ID              int
	Bots            models.BotProvider
	ShipmentUC      models.ShipmentUsecase
	ConfigUC        models.ConfigUsecase
	Jobs            <-chan models.Job
	WG              *sync.WaitGroup
	Cfg             *config.Config
	FrontendURL     string
	ShipmentService shipment.Service
}

func (w *Worker) Start() {
	defer w.WG.Done()
	logger.Info().Int("worker_id", w.ID).Msg("Worker started")

	for job := range w.Jobs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error().Msgf("Worker %d panicked: %v\n%s", w.ID, r, string(debug.Stack()))
				}
			}()
			w.process(job)
		}()
	}
}

func (w *Worker) process(job models.Job) {
	logger.GlobalVitals.IncJobs()

	bot, err := w.Bots.GetBot(job.CompanyID)
	if err != nil {
		logger.Error().Err(err).Str("company", job.CompanyID.String()).Msg("Bot instance not found for company")
		return
	}

	// 1. Fetch Language
	langStr, _ := w.ConfigUC.GetUserLanguage(context.Background(), job.CompanyID, job.SenderJID.String())
	lang := i18n.Language(langStr)

	// A. Initial Feedback (Typing)
	sender := bot.GetSender()
	sender.SetTyping(job.ChatJID, true)
	defer sender.SetTyping(job.ChatJID, false)

	// 2. Check for Commands
	ctx := utils.WithValues(context.Background(), job.SenderJID.String(), job.SenderPhone, job.IsAdmin, job.ChatJID.String(), job.MessageID, job.Text)

	botPhone := ""
	wa := bot.GetWAClient()
	if wa != nil && wa.Store != nil && wa.Store.ID != nil {
		botPhone = utils.GetBarePhone(wa.Store.ID.User)
	}

	dispatcher := commands.NewDispatcher(w.Cfg, w.ShipmentUC, w.ConfigUC, sender, bot.GetPrefix(), bot.GetCompanyName(), botPhone, w.Cfg.AdminTimezone, bot.GetTier())
	if res, ok := dispatcher.Dispatch(ctx, job.CompanyID, job.Text); ok {
		if len(res.Image) > 0 {
			sender.SendImage(job.ChatJID, job.SenderJID, res.Image, res.Message, job.MessageID, job.Text)
		} else if res.Message != "" {
			sender.Reply(job.ChatJID, job.SenderJID, res.Message, job.MessageID, job.Text)
		}

		// If it was an edit, we need to regenerate the receipt
		if res.EditID != "" {
			logger.Info().Str("edit_id", res.EditID).Msg("Edit detected, triggering receipt regeneration")
			w.generateAndSendReceipt(bot, job, res.EditID, lang)
		}
		return
	}

	// X. Extract Document Text (if any)
	if job.RawMessage != nil && job.RawMessage.Message.GetDocumentMessage() != nil {
		doc := job.RawMessage.Message.GetDocumentMessage()
		data, err := wa.Download(context.Background(), doc)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to download document message")
		} else {
			mimeType := doc.GetMimetype()
			extracted, err := parser.ExtractDocumentText(data, mimeType)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to parse document text")
			} else if extracted != "" {
				// Append extracted text to any existing caption/message text
				job.Text = strings.TrimSpace(job.Text + "\n" + extracted)
				logger.Info().Str("mime_type", mimeType).Msg("Successfully extracted text from document manifest")
			}
		}
	}

	// 2. Initial Checks
	isManifest, isPartial := w.isPotentialManifest(job.Text)
	if !isManifest && !isPartial {
		return // Completely unrelated message
	}

	// 3. Normal Parsing (Regex first)
	m := parser.ParseRegex(job.Text)

	// AI Fallback (Strictly bound to save costs and API limits)
	// ONLY use AI if the user provided a full manifest structure (isManifest == true)
	// BUT the regex struggled to extract all the required fields.
	// If it's just a partial message, we skip AI and immediately report the missing fields.
	if isManifest && (m.ReceiverName == "" || m.ReceiverPhone == "" || m.ReceiverAddress == "" || m.SenderName == "" || m.ReceiverCountry == "") {
		if aiM, err := parser.ParseAI(job.Text, w.Cfg.GeminiAPIKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
		}
	}

	// 4. Validation
	// Ensure Validate operates correctly after merge or regex
	missing := m.Validate()
	if len(missing) > 0 {
		logger.GlobalVitals.IncParseFailure()
		logger.Warn().
			Str("jid", job.SenderJID.String()).
			Strs("missing_fields", missing).
			Str("raw_text", job.Text).
			Msg("Information incomplete after parsing")

		// Specifically list the exact missing fields
		msg := "📝 *INFORMATION INCOMPLETE*\n\n━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"The system could not parse the following required fields:\n" +
			"• " + strings.Join(missing, "\n• ") + "\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n\n_Please provide the missing data to proceed._"
		sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID, job.Text)
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
		Weight:    m.Weight,
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
	if newShipment.Weight <= 0 {
		newShipment.Weight = 15.0 // Fallback if parser didn't extract weight
	}

	// 5b. Deduplication Check (Strict Phone Match)
	if existingID, err := w.ShipmentUC.FindSimilar(context.Background(), job.CompanyID, job.SenderJID.String(), newShipment.RecipientPhone); err == nil && existingID != "" {
		logger.Info().Str("existing_id", existingID).Msg("Duplicate shipment blocked")
		dupMsg := fmt.Sprintf("⚠️ *SHIPMENT ALREADY EXISTS*\n\nA shipment for this recipient phone is already in the system.\n\n🆔 *%s*\n\n🔹 Use `!edit %s ...` to update.\n🔹 Use `!delete %s` to remove.", existingID, existingID, existingID)
		sender.Reply(job.ChatJID, job.SenderJID, dupMsg, job.MessageID, job.Text)
		return
	}

	// 5c. Billing Limit Check
	company, err := w.ConfigUC.GetCompanyByID(context.Background(), job.CompanyID)
	if err == nil {
		remaining, err := w.ShipmentUC.CheckShipmentCap(context.Background(), w.Cfg, job.CompanyID, company.PlanType.String, company.SubscriptionExpiry)
		if err == nil && remaining == 0 {
			logger.Info().Str("company_id", job.CompanyID.String()).Msg("Shipment blocked: billing limit reached")
			sender.Reply(job.ChatJID, job.SenderJID, "⚠️ *SHIPMENT BLOCKED*\n\nYour monthly shipment limit has been reached or your subscription has expired.\n\nPlease contact your administrator to upgrade your plan via the dashboard.", job.MessageID, job.Text)
			return
		}
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

	trackingID, err := w.ShipmentUC.CreateWithPrefix(context.Background(), job.CompanyID, dbShip, bot.GetPrefix())
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		logger.Error().Err(err).Str("jid", job.SenderJID.String()).Msg("Failed to insert shipment information")
		sender.Reply(job.ChatJID, job.SenderJID, "❌ *SYSTEM ERROR*\n_Saving information failed. Please contact your admin._", job.MessageID, job.Text)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 7, 8, 9. Generate and send receipt
	w.generateAndSendReceipt(bot, job, trackingID, lang)

	logger.Info().
		Str("tracking_id", trackingID).
		Str("jid", job.SenderJID.String()).
		Msg("Shipment created successfully")

	// 10. Send tracking ID and link as follow-up message
	baseURL := w.FrontendURL
	if baseURL == "" {
		baseURL = os.Getenv("FRONTEND_URL")
	}

	trackingMsg := fmt.Sprintf("📦 *SHIPMENT INFORMATION CREATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nTracking ID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n📌 *Track your package:*\n%s/track/%s", trackingID, baseURL, trackingID)
	if m.IsAI {
		trackingMsg += "\n\n_✨ Parsed by AI_"
	}
	sender.Reply(job.ChatJID, job.SenderJID, trackingMsg, job.MessageID, job.Text)
}

func (w *Worker) generateAndSendReceipt(bot models.BotInstance, job models.Job, id string, lang i18n.Language) {
	receipt.Enqueue(receipt.Job{
		Msg:         job,
		TrackingID:  id,
		Language:    lang,
		CompanyName: bot.GetCompanyName(),
		ShipmentUC:  w.ShipmentUC,
		Sender:      bot.GetSender(),
		RenderMode:  "default",
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
	phoneKeywords := []string{"phone", "mobile", "mob", "tel", "num", "contact", "telephone", "mobil", "number"}
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
