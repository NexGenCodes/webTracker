package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/database/dbutil"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/tasks"
	"webtracker-bot/internal/utils"
)

var (
	senderPattern   = regexp.MustCompile(`(?i)sender|origin|from|absender|remetente|remitente`)
	receiverPattern = regexp.MustCompile(`(?i)receiver|reciver|receive|recieve|resiver|recever|consignee|destinatario|destinatário|empfänger|destinataire`)
	phonePattern    = regexp.MustCompile(`(?i)phone|mobile|mob|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp|handy`)
	namePattern     = regexp.MustCompile(`(?i)name|nombre|nome|nom`)
)

// Worker processes incoming WhatsApp messages and executes commands.
type Worker struct {
	ID              int
	Bots            models.BotProvider
	ShipmentUC      models.ShipmentUsecase
	ConfigUC        models.ConfigUsecase
	Cfg             *config.Config
	FrontendURL     string
	ShipmentService shipment.Service
	Context         context.Context
	AsynqClient     *asynq.Client
	Queries         db.Querier
}

// Start registers the Asynq handlers and begins processing.
func (w *Worker) Start(mux *asynq.ServeMux) {
	logger.Info().Int("worker_id", w.ID).Msg("Worker registering Asynq handlers")
	// Registration happens externally in app.go, but we could do it here
}

// HandleWhatsAppMessage is the Asynq Handler interface method
func (w *Worker) HandleWhatsAppMessage(ctx context.Context, t *asynq.Task) error {
	var payload tasks.WhatsAppMessagePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	chatJID, _ := types.ParseJID(payload.ChatJID)
	senderJID, _ := types.ParseJID(payload.SenderJID)

	job := models.Job{
		CompanyID:   payload.CompanyID,
		ChatJID:     chatJID,
		SenderJID:   senderJID,
		MessageID:   payload.MessageID,
		Text:        payload.Text,
		SenderPhone: payload.SenderPhone,
		Language:    payload.Language,
		IsAdmin:     payload.IsAdmin,
	}

	// Restore the protobuf message if it exists
	if len(payload.RawMessageBytes) > 0 {
		msg := &waProto.Message{}
		if err := proto.Unmarshal(payload.RawMessageBytes, msg); err == nil {
			job.RawMessage = &events.Message{
				Message: msg,
			}
		}
	}

	w.process(ctx, job)
	return nil
}

func (w *Worker) process(workerCtx context.Context, job models.Job) {
	logger.GlobalVitals.IncJobs()

	bot, err := w.Bots.GetBot(job.CompanyID)
	if err != nil {
		logger.Error().Err(err).Str("company", job.CompanyID.String()).Msg("Bot instance not found for company")
		return
	}

	// 1. Fetch Metadata in Parallel
	var (
		langStr    string
		existingID string
		company    db.Company
		remaining  int64 = -1 // -1 means not checked yet
	)

	// Add a 15 second timeout for all operations inside process
	processCtx, processCancel := context.WithTimeout(w.Context, 15*time.Second)
	defer processCancel()

	g, gctx := errgroup.WithContext(processCtx)

	// A. User Language
	g.Go(func() error {
		langStr, _ = w.ConfigUC.GetUserLanguage(gctx, job.CompanyID, job.SenderJID.String())
		return nil
	})

	// B. Billing Status
	g.Go(func() error {
		var err error
		company, err = w.ConfigUC.GetCompanyByID(gctx, job.CompanyID)
		return err
	})

	if err := g.Wait(); err != nil {
		logger.Error().Err(err).Str("company", job.CompanyID.String()).Msg("Failed to fetch initial metadata")
		return
	}

	lang := i18n.Language(langStr)

	// Subscription Guard: Block expired or inactive tenants from using the bot
	// The super admin is always free to access anything and bypasses all billing checks.
	isSuperAdmin := billing.IsSuperAdminByEmail(w.Cfg, company.AdminEmail)
	
	if !isSuperAdmin {
		if company.SubscriptionStatus.String != "active" && company.SubscriptionStatus.String != "trialing" {
			logger.Info().Str("company_id", job.CompanyID.String()).Str("status", company.SubscriptionStatus.String).Msg("Ignoring message from inactive subscription")
			return
		}
		if company.SubscriptionExpiry.Valid && company.SubscriptionExpiry.Time.Before(time.Now()) {
			logger.Info().Str("company_id", job.CompanyID.String()).Msg("Ignoring message from expired subscription")
			return
		}
	}

	// 1. Intent Detection (Pre-reaction filtering)
	isCommand := len(job.Text) > 1 && (job.Text[0] == '!' || job.Text[0] == '#' || job.Text[0] == '/')
	isManifest, isPartial := w.isPotentialManifest(job.Text)
	hasDocument := job.RawMessage != nil && job.RawMessage.Message.GetDocumentMessage() != nil

	// If it's not a command, not a manifest, and doesn't have a document, ignore it immediately.
	// This prevents the bot from reacting to normal admin conversations.
	if !isCommand && !isManifest && !isPartial && !hasDocument {
		return
	}

	// A. Initial Feedback (Typing, Reading, Reacting)
	// We only get here if the message is something the bot should actually process.
	sender := bot.GetSender()

	sender.MarkRead(job.ChatJID, job.SenderJID, job.MessageID)
	sender.React(job.ChatJID, job.SenderJID, job.MessageID, "🔍")

	sender.SetTyping(job.ChatJID, true)
	defer sender.SetTyping(job.ChatJID, false)

	// 2. Check for Commands
	ctx := utils.WithValues(processCtx, job.SenderJID.String(), job.SenderPhone, job.IsAdmin, job.ChatJID.String(), job.MessageID, job.Text)

	botPhone := ""
	wa := bot.GetWAClient()
	if wa != nil && wa.Store != nil && wa.Store.ID != nil {
		botPhone = utils.GetBarePhone(wa.Store.ID.User)
	}

	dispatcher := commands.NewDispatcher(w.Cfg, w.ShipmentUC, w.ConfigUC, sender, bot.GetPrefix(), bot.GetCompanyName(), botPhone, bot.GetTier())
	if res, ok := dispatcher.Dispatch(ctx, job.CompanyID, job.Text); ok {
		if len(res.Image) > 0 {
			sender.SendImage(job.ChatJID, job.SenderJID, res.Image, res.Message, job.MessageID, job.Text)
		} else if res.Message != "" {
			sender.Reply(job.ChatJID, job.SenderJID, res.Message, job.MessageID, job.Text)
		}

		// If it was an edit, we need to regenerate the receipt
		if res.EditID != "" {
			logger.Info().Str("edit_id", res.EditID).Msg("Edit detected, triggering receipt regeneration")
			// Clear the dedup guard so the edited receipt always gets generated
			receipt.ClearInflight(res.EditID)
			w.generateAndSendReceipt(bot, job, res.EditID, lang)
		}

		// Change reaction to success since command was handled
		sender.React(job.ChatJID, job.SenderJID, job.MessageID, "✅")
		return
	}

	// X. Extract Document Text (if any)
	if job.RawMessage != nil && job.RawMessage.Message.GetDocumentMessage() != nil {
		doc := job.RawMessage.Message.GetDocumentMessage()
		data, err := wa.Download(processCtx, doc)
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
	isManifest, isPartial = w.isPotentialManifest(job.Text)
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
		aiCtx, aiCancel := context.WithTimeout(processCtx, 7*time.Second)
		defer aiCancel()
		if aiM, err := parser.ParseAI(aiCtx, job.Text, w.Cfg.GeminiAPIKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
		} else {
			if aiCtx.Err() == context.DeadlineExceeded {
				logger.Warn().Str("jid", job.SenderJID.String()).Msg("AI parsing timed out (7s)")
			} else {
				logger.Error().Err(err).Str("jid", job.SenderJID.String()).Msg("AI parsing failed")
			}
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
		missingStr := "• " + strings.Join(missing, "\n• ")
		msg := i18n.T(lang, "MSG_INFO_INCOMPLETE", missingStr)
		sender.Reply(job.ChatJID, job.SenderJID, msg, job.MessageID, job.Text)
		sender.React(job.ChatJID, job.SenderJID, job.MessageID, "⚠️")
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
		newShipment.Weight = 15.0
	}

	//  Deduplication & Billing Check in Parallel
	g, gctx = errgroup.WithContext(processCtx)

	g.Go(func() error {
		var err error
		existingID, err = w.ShipmentUC.FindSimilar(gctx, job.CompanyID, job.SenderJID.String(), newShipment.RecipientPhone)
		return err
	})

	g.Go(func() error {
		var err error
		remaining, err = w.ShipmentUC.CheckShipmentCap(gctx, job.CompanyID, isSuperAdmin, company.PlanType.String, company.SubscriptionExpiry)
		return err
	})

	if err := g.Wait(); err != nil {
		logger.Error().Err(err).Msg("Failed to perform secondary checks")
	}

	if existingID != "" {
		logger.Info().Str("existing_id", existingID).Msg("Duplicate shipment blocked")
		dupMsg := fmt.Sprintf("⚠️ *SHIPMENT ALREADY EXISTS*\n\nA shipment for this recipient phone is already in the system.\n\n🆔 *%s*\n\n🔹 Use `!edit %s ...` to update.\n🔹 Use `!delete %s` to remove.", existingID, existingID, existingID)
		sender.Reply(job.ChatJID, job.SenderJID, dupMsg, job.MessageID, job.Text)
		sender.React(job.ChatJID, job.SenderJID, job.MessageID, "⚠️")
		return
	}

	if remaining == 0 {
		logger.Info().Str("company_id", job.CompanyID.String()).Msg("Shipment blocked: billing limit reached")
		sender.Reply(job.ChatJID, job.SenderJID, "⚠️ *SHIPMENT BLOCKED*\n\nYour monthly shipment limit has been reached or your subscription has expired.\n\nPlease contact your administrator to upgrade your plan via the dashboard.", job.MessageID, job.Text)
		sender.React(job.ChatJID, job.SenderJID, job.MessageID, "⛔")
		return
	}

	// Generate schedule dates using the new Smart Anchor Algorithm (A & B)
	now := time.Now().UTC()
	departure := w.ShipmentService.CalculateDeparture(now, newShipment.Origin)
	arrival, outForDelivery := w.ShipmentService.CalculateArrival(departure, newShipment.Destination)

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
		Weight:               dbutil.FloatToNullNumeric(newShipment.Weight),
		Cost:                 dbutil.FloatToNullNumeric(newShipment.Cost),
	}

	trackingID, err := w.ShipmentUC.CreateWithPrefix(processCtx, job.CompanyID, dbShip, bot.GetPrefix())
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		logger.Error().Err(err).Str("jid", job.SenderJID.String()).Msg("Failed to insert shipment information")
		sender.Reply(job.ChatJID, job.SenderJID, "❌ *SYSTEM ERROR*\n_Saving information failed. Please contact your admin._", job.MessageID, job.Text)
		sender.React(job.ChatJID, job.SenderJID, job.MessageID, "❌")
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	logger.Info().
		Str("tracking_id", trackingID).
		Str("jid", job.SenderJID.String()).
		Msg("Shipment created successfully")

	// React ✅ immediately — before the (slow) receipt render
	sender.React(job.ChatJID, job.SenderJID, job.MessageID, "✅")

	// Generate and send receipt (async queue — does not block)
	w.generateAndSendReceipt(bot, job, trackingID, lang)

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
	hasSender := senderPattern.MatchString(text)
	hasReceiver := receiverPattern.MatchString(text)
	hasPhone := phonePattern.MatchString(text)
	hasName := namePattern.MatchString(text)

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
