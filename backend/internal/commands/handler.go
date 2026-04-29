package commands

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"
)

func i18nLang(s string) i18n.Language {
	return i18n.Language(strings.ToLower(s))
}

// Result represents the outcome of a command execution.
type Result struct {
	Message  string
	Language string
	EditID   string
	Image    []byte // Binary payload for direct responses
	Error    error
}

type Handler interface {
	Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result
}

type Dispatcher struct {
	cfg           *config.Config
	shipUC        *shipment.Usecase
	configUC      *config.Usecase
	sender        *whatsapp.Sender // Added for broadcasting
	handlers      map[string]Handler
	AwbCmd        string
	CompanyName   string
	Tier          string
	BotPhone      string
	AdminTimezone string
}

func NewDispatcher(cfg *config.Config, shipUC *shipment.Usecase, configUC *config.Usecase, sender *whatsapp.Sender, awbCmd string, companyName string, botPhone string, adminTimezone string, tier string) *Dispatcher {
	d := &Dispatcher{
		cfg:           cfg,
		shipUC:        shipUC,
		configUC:      configUC,
		sender:        sender,
		handlers:      make(map[string]Handler),
		AwbCmd:        awbCmd,
		CompanyName:   companyName,
		Tier:          tier,
		BotPhone:      botPhone,
		AdminTimezone: adminTimezone,
	}
	d.registerDefaults()
	return d
}

func (d *Dispatcher) registerDefaults() {
	d.handlers["stats"] = &StatsHandler{}
	d.handlers["info"] = &InfoHandler{}
	d.handlers["help"] = &HelpHandler{}
	d.handlers["lang"] = &LangHandler{}
	d.handlers["edit"] = &EditHandler{}
	d.handlers["delete"] = &DeleteHandler{}
	d.handlers["status"] = &StatusHandler{}
	d.handlers["receipt"] = &ReceiptHandler{}
}

func (d *Dispatcher) Dispatch(ctx context.Context, companyID uuid.UUID, text string) (*Result, bool) {
	if !presentsAsCommand(text) {
		return nil, false
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, false
	}
	rawCmd := strings.ToLower(parts[0][1:]) // Remove "!" prefix
	args := parts[1:]

	if handler, ok := d.handlers[rawCmd]; ok {
		jid := utils.GetJID(ctx)
		senderPhone := utils.GetSenderPhone(ctx)
		isAdmin := utils.IsAdmin(ctx)

		isOwner := senderPhone == d.BotPhone

		if !isOwner {
			allowed, retryIn := utils.Allow(senderPhone, d.Tier)
			if !allowed {
				return &Result{Message: fmt.Sprintf("⏳ *RATE LIMIT REACHED*\n\n_Please wait %d seconds before sending another command._", int(retryIn.Seconds()))}, true
			}
		}

		switch h := handler.(type) {
		case *StatsHandler:
			h.CompanyName = d.CompanyName
			h.AdminTimezone = d.AdminTimezone
		case *InfoHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *HelpHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *EditHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
			h.AdminTimezone = d.AdminTimezone
			h.Sender = d.sender
			h.Cfg = d.cfg
		case *StatusHandler:
			h.BotPhone = d.BotPhone
		case *ReceiptHandler:
			h.Sender = d.sender
		}

		lang, _ := d.configUC.GetUserLanguage(ctx, companyID, jid)

		isOwnerOnlyCmd := rawCmd == "status"
		if isOwnerOnlyCmd {
			if !isOwner {
				logger.Warn().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Owner-only command blocked")
				return &Result{Message: i18n.T(i18nLang(lang), "ERR_OWNER_ONLY")}, true
			}
		}

		isPublicCmd := rawCmd == "info" || rawCmd == "help"
		if !isPublicCmd && !isOwnerOnlyCmd {
			if isAdmin {
				logger.Info().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Admin command authorized")
			} else {
				return &Result{Message: i18n.T(i18nLang(lang), "ERR_ACCESS_DENIED")}, true
			}
		}

		res := handler.Execute(ctx, d.shipUC, d.configUC, companyID, args, lang, isAdmin)
		if res.Language != "" {
			d.configUC.SetUserLanguage(ctx, companyID, jid, res.Language)
		}

		return &res, true
	}

	return nil, false
}

func presentsAsCommand(text string) bool {
	return len(text) > 1 && (text[0] == '!' || text[0] == '#')
}

// StatsHandler handles !stats
type StatsHandler struct {
	CompanyName   string
	AdminTimezone string
}

func (h *StatsHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) > 0 {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_INCORRECT_USAGE")}
	}

	stats, err := shipUC.CountByStatus(ctx, companyID)
	if err != nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_SYSTEM_ERROR"), Error: err}
	}

	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	msg := i18n.T(i18nLang(lang), "MSG_STATS_HEADER", company) + "\n\n━━━━━━━━━━━━━━━━━━━━━━━\n" +
		fmt.Sprintf("📦 PENDING:    *%d*\n", stats.Pending) +
		fmt.Sprintf("🚚 IN TRANSIT: *%d*\n", stats.Intransit) +
		fmt.Sprintf("🏠 AT DEST:    *%d*\n", stats.Outfordelivery) +
		fmt.Sprintf("🏁 DELIVERED:  *%d*\n", stats.Delivered) +
		fmt.Sprintf("📊 TOTAL:      *%d*\n", stats.Total) +
		"━━━━━━━━━━━━━━━━━━━━━━━\n\n_Real-time operational dashboard._"

	return Result{Message: msg}
}

// InfoHandler handles !info [ID]
type InfoHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *InfoHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		// Attempt contextual lookup
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			company := strings.ToUpper(h.CompanyName)
			if company == "" {
				company = "COMMAND"
			}
			msg := fmt.Sprintf("🚀 *%s COMMAND CENTER*\n\n", company) +
				"━━━━━━━━━━━━━━━━━━━━━━━\n"

			if isAdmin {
				msg += "1️⃣ `!stats` - Daily Operations Summary\n"
			}

			msg += "2️⃣ `!info [TrackingID]` - Shipment Information Tracker\n" +
				"━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
				"*PRO TIP:*\n" +
				fmt.Sprintf("_Use `!info %s-123456789` for full details._", h.CompanyPrefix)
			return Result{Message: msg}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	dbShip, err := shipUC.Track(ctx, companyID, trackingID)
	if err != nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_DB_ERROR"), Error: err}
	}

	if dbShip == nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_NOT_FOUND")}
	}

	// Map DB model to Domain model for waybill generation
	s := shipment.Shipment{
		TrackingID:        dbShip.TrackingID,
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

	wb := receipt.GenerateWaybill(s, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}

// HelpHandler handles !help
type HelpHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *HelpHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	if isAdmin {
		// Admin Help Menu as Plain Text
		msg := fmt.Sprintf("🛡️ *%s ADMIN COMMANDS*\n\n", company) +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"📊 `!stats` - Today's operations\n" +
			"🌡️ `!status` - System health & vitals\n" +
			"✏️ `!edit [ID] [updates]` - Update shipment\n" +
			"🗑️ `!delete [ID]` - Remove shipment\n" +
			"📦 `!info [ID]` - Detailed waybill\n" +
			"🌐 `!lang [en|pt|es|de]` - Switch language\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"_Use these commands strictly within the authorized groups._"
		return Result{Message: msg}
	}

	// Customer Help Menu as Plain Text
	msg := fmt.Sprintf("📖 *%s CUSTOMER SERVICE*\n\n", company) +
		"How can we help you today?\n\n" +
		"🔎 `!info [ID]` - Track your shipment\n" +
		"📖 `!help` - View this menu\n" +
		"🌐 `!lang [code]` - Change language\n\n" +
		"_Please type the command manually to interact with the bot._"
	return Result{Message: msg}
}

type LangHandler struct{}

func (h *LangHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "🌐 *LANGUAGE MENU*\n\nUsage: `!lang [en|pt|es|de]`\n\nExample: `!lang pt` para Português"}
	}

	newLang := strings.ToLower(args[0])
	switch newLang {
	case "en", "pt", "es", "de":
		// Handled by dispatcher update
		return Result{
			Message:  fmt.Sprintf("🌐 *LANGUAGE UPDATED*\n\nYour language is now set to *%s*.", strings.ToUpper(newLang)),
			Language: newLang,
		}
	default:
		return Result{Message: "❌ *UNSUPPORTED LANGUAGE*\n\nAvailable: `en`, `pt`, `es`, `de`"}
	}
}

// EditHandler handles !edit [trackingID] [field] [value] or !edit [field] [value]
type EditHandler struct {
	CompanyName   string
	CompanyPrefix string
	AdminTimezone string
	Sender        *whatsapp.Sender
	Cfg           *config.Config
}

func (h *EditHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "✏️ *EDIT SHIPMENT*\n\nUsage: `!edit [TrackingID] [Updates...]` or `!edit [Updates...]` (targets last shipment)\n\n*Example:* `!edit LGS-1234 name: John, departure: tomorrow`"}
	}

	jid := utils.GetJID(ctx)
	var trackingID string
	var startIdx int

	// 1. Identify Target Shipment
	// Pattern: 3 uppercase letters, hyphen, 4 digits
	idPattern := regexp.MustCompile(`^[A-Z]{3}-\d{4,9}$`)
	if idPattern.MatchString(args[0]) {
		trackingID = args[0]
		startIdx = 1
	} else {
		// Contextual Lookup: Fetch last shipment for this user
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: i18n.T(i18nLang(lang), "ERR_CONTEXT_ERROR")}
		}
		startIdx = 0
	}

	// 2. Parse Updates
	updateText := strings.Join(args[startIdx:], " ")
	updates := parser.ParseEditPairs(updateText)

	// Fallback for single field (e.g., !edit name Mark) if parser didn't find clear anchors
	if len(updates) == 0 && len(args[startIdx:]) >= 2 {
		// This handles the old style: !edit field value
		field := args[startIdx]
		value := strings.Join(args[startIdx+1:], " ")
		// Reuse canonical mapping logic
		normField := strings.ToLower(field)
		dbField := ""
		switch normField {
		case "receiver", "name", "recipient", "reciever":
			dbField = "recipient_name"
		case "phone", "mobile", "number", "contact":
			dbField = "recipient_phone"
		case "address", "addr":
			dbField = "recipient_address"
		case "destination", "country", "to":
			dbField = "destination"
		case "origin", "from", "source":
			dbField = "origin"
		case "id", "passport", "identification":
			dbField = "recipient_id"
		case "email", "mail":
			dbField = "recipient_email"
		case "departure", "transit":
			dbField = "scheduled_transit_time"
		case "arrival", "delivery":
			dbField = "expected_delivery_time"
		case "outfordelivery", "out_for_delivery":
			dbField = "outfordelivery_time"
		case "cargo", "type", "content":
			dbField = "cargo_type"
		case "weight":
			dbField = "weight"
		}
		if dbField != "" {
			updates[dbField] = value
		}
	}

	if len(updates) == 0 {
		return Result{Message: "⚠️ *NO UPDATES FOUND*\n_Please specify what you want to change (e.g., 'name: John' or 'departure: tomorrow')._"}
	}

	// 3. Apply Updates
	var updatedFields []string
	departureUpdated := false
	var newDeparture time.Time
	arrivalExplicitlyUpdated := false

	for field, value := range updates {
		// Strict Policy: Weight is fixed
		if field == "weight" {
			continue
		}

		// Validation
		if strings.Contains(field, "email") && !parser.ValidateEmail(value) {
			continue
		}
		if strings.Contains(field, "phone") && !parser.ValidatePhone(value) {
			continue
		}

		// Special Date Parsing
		if field == "scheduled_transit_time" || field == "expected_delivery_time" || field == "outfordelivery_time" {
			tz := h.AdminTimezone
			if tz == "" {
				tz = "Africa/Lagos"
			}
			loc, _ := time.LoadLocation(tz)
			now := time.Now().In(loc)

			if parsedDate, ok := utils.ParseNaturalDate(value, now); ok {
				value = parsedDate.UTC().Format("2006-01-02 15:04:05")
			} else {
				// Fallback to strict format
				_, err := time.Parse("2006-01-02", value)
				if err != nil {
					_, err = time.Parse("2006-01-02 15:04:05", value)
					if err != nil {
						continue // Skip invalid date
					}
				}
			}

			if field == "scheduled_transit_time" {
				departureUpdated = true
				newDeparture, _ = time.Parse("2006-01-02 15:04:05", value)
			}
			if field == "expected_delivery_time" {
				arrivalExplicitlyUpdated = true
			}
		}

		err := shipUC.UpdateField(ctx, companyID, trackingID, field, value)
		if err == nil {
			updatedFields = append(updatedFields, strings.ToUpper(strings.ReplaceAll(field, "_", " ")))
		}
	}

	// 4. Automatic Arrival Sync (Algorithm B)
	if departureUpdated && !arrivalExplicitlyUpdated {
		dbShip, _ := shipUC.Track(ctx, companyID, trackingID)
		if dbShip != nil {
			// Recalculate Arrival based on new Departure
			arrival, outForDelivery := shipUC.Service.CalculateArrival(newDeparture, dbShip.Origin.String, dbShip.Destination.String)

			_ = shipUC.UpdateField(ctx, companyID, trackingID, "expected_delivery_time", arrival.Format("2006-01-02 15:04:05"))
			_ = shipUC.UpdateField(ctx, companyID, trackingID, "outfordelivery_time", outForDelivery.Format("2006-01-02 15:04:05"))

			updatedFields = append(updatedFields, "EXPECTED DELIVERY TIME (AUTO-SYNC)", "OUTFORDELIVERY TIME (AUTO-SYNC)")
		}
	}

	if len(updatedFields) == 0 {
		return Result{Message: "⚠️ *UPDATE FAILED*\n_None of the fields could be updated. Check your format (e.g., label: value)._"}
	}

	// 4. Persistence & Schedule Sync
	dbShip, _ := shipUC.Track(ctx, companyID, trackingID)
	if dbShip != nil {
		// Resolve status
		s := shipment.Shipment{
			Status: dbShip.Status.String,
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

		newStatus := s.ResolveStatus(time.Now().UTC())
		if newStatus != dbShip.Status.String {
			_ = shipUC.UpdateField(ctx, companyID, trackingID, "status", newStatus)

			// If it transitions, optionally trigger the notification explicitly!
			if h.Sender != nil && h.Sender.Client != nil {
				notif.SendStatusAlert(ctx, h.Sender.Client, h.Cfg, h.CompanyName, dbShip.UserJid, trackingID, newStatus, dbShip.RecipientEmail.String)
			}
		}
	}

	summary := fmt.Sprintf("✅ *INFORMATION UPDATED*\n\n🆔 *%s*\n\n📝 *FIELDS MODIFIED:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Updates have been successfully persisted to the cloud._",
		trackingID, strings.Join(updatedFields, "\n• "))

	return Result{
		Message: summary,
		EditID:  trackingID,
	}
}

// DeleteHandler handles !delete [trackingID]
type DeleteHandler struct{}

func (h *DeleteHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "🗑️ *DELETE SHIPMENT*\n\nUsage: `!delete [TrackingID]`"}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	err := shipUC.Delete(ctx, companyID, trackingID)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *DELETE FAILED*\n_%v_", err)}
	}

	return Result{Message: fmt.Sprintf("🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.", trackingID)}
}

// StatusHandler handles !status
type StatusHandler struct {
	BotPhone string
}

func (h *StatusHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	// Performance Telemetry
	uptime := time.Since(logger.GlobalVitals.StartTime)
	jobs := atomic.LoadInt64(&logger.GlobalVitals.JobsProcessed)
	success := atomic.LoadInt64(&logger.GlobalVitals.ParseSuccess)

	// Memory usage tracking
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memMB := m.Alloc / 1024 / 1024

	// System Health
	dbStatus := "🟢 ONLINE"
	if err := configUC.Ping(ctx); err != nil {
		dbStatus = "🔴 OFFLINE"
		logger.Error().Err(err).Msg("Database ping failed in !status")
	}

	groupsCount, _ := configUC.CountAuthorizedGroups(ctx, companyID)

	msg := i18n.T(i18nLang(lang), "MSG_STATUS_DASHBOARD") + "\n\n" +
		fmt.Sprintf("📊 UPTIME:    *%s*\n", uptime.Truncate(time.Second)) +
		fmt.Sprintf("🔋 MEMORY:    *%d MB* / 1024 MB\n", memMB) +
		fmt.Sprintf("🗄️ DATABASE:  *%s*\n", dbStatus) +
		fmt.Sprintf("👥 GROUPS:    *%d authorized*\n", groupsCount) +
		fmt.Sprintf("📦 PROCESSED: *%d jobs* (%d success)\n\n", jobs, success) +
		"_System is running within safe 1GB RAM margins._"

	return Result{Message: msg}
}

// ReceiptHandler handles !receipt [ID]
type ReceiptHandler struct {
	Sender *whatsapp.Sender
}

func (h *ReceiptHandler) Execute(ctx context.Context, shipUC *shipment.Usecase, configUC *config.Usecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if !isAdmin {
		return Result{Message: "🔒 *ACCESS DENIED*\nOnly admins can regenerate receipts."}
	}

	var trackingID string
	jid := utils.GetJID(ctx)

	if len(args) < 1 {
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "🧾 *RECEIPT REGENERATION*\n\nUsage: `!receipt [TrackingID]`"}
		}
	} else {
		trackingID = strings.ToUpper(args[0])
	}

	// Fetch shipment to get JID
	dbShip, err := shipUC.Track(ctx, companyID, trackingID)
	if err != nil || dbShip == nil {
		return Result{Message: "❌ *NOT FOUND*\nCould not find a shipment with that ID."}
	}

	// Map to Domain Model for Rendering
	s := shipment.Shipment{
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

	// Render synchronous
	receiptImg, err := receipt.RenderReceipt(s, h.Sender.CompanyName, i18n.Language(lang))
	if err != nil {
		return Result{Message: "❌ *RENDER FAILED*", Error: err}
	}

	return Result{
		Message: "", // Binary response handled by worker
		Image:   receiptImg,
	}
}

