package commands

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/usecase"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

func i18nLang(s string) i18n.Language {
	return i18n.Language(strings.ToLower(s))
}

// Result represents the outcome of a command execution.
type Result struct {
	Message     string
	Language    string
	EditID      string
	List        *waProto.ListMessage
	Buttons     *waProto.ButtonsMessage
	Error       error
}

type Handler interface {
	Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result
}

type Dispatcher struct {
	cfg           *config.Config
	shipUC        *usecase.ShipmentUsecase
	configUC      *usecase.ConfigUsecase
	sender        *whatsapp.Sender // Added for broadcasting
	handlers      map[string]Handler
	AwbCmd        string
	CompanyName   string
	BotPhone      string
	AdminTimezone string
}

func NewDispatcher(cfg *config.Config, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, sender *whatsapp.Sender, awbCmd string, companyName string, botPhone string, adminTimezone string) *Dispatcher {
	d := &Dispatcher{
		cfg:           cfg,
		shipUC:        shipUC,
		configUC:      configUC,
		sender:        sender,
		handlers:      make(map[string]Handler),
		AwbCmd:        awbCmd,
		CompanyName:   companyName,
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
	d.handlers["broadcast"] = &BroadcastHandler{}
	d.handlers["status"] = &StatusHandler{}
}

func (d *Dispatcher) Dispatch(ctx context.Context, text string) (*Result, bool) {
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
			allowed, retryIn := utils.Allow(senderPhone)
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
			h.CompanyPrefix = d.AwbCmd
			h.AdminTimezone = d.AdminTimezone
			h.Sender = d.sender
			h.Cfg = d.cfg
		case *BroadcastHandler:
			h.Sender = d.sender
		case *StatusHandler:
			h.BotPhone = d.BotPhone
		}

		lang, _ := d.configUC.GetUserLanguage(ctx, jid)

		isOwnerOnlyCmd := rawCmd == "broadcast" || rawCmd == "status"
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

		res := handler.Execute(ctx, d.shipUC, d.configUC, args, lang, isAdmin)
		if res.Language != "" {
			d.configUC.SetUserLanguage(ctx, jid, res.Language)
		}

		// Handle Interactive Content
		targetJID, _ := types.ParseJID(jid)
		if res.List != nil {
			d.sender.SendList(targetJID, *res.List.Title, *res.List.Description, *res.List.ButtonText, res.List.Sections)
			return &res, true
		}
		if res.Buttons != nil {
			d.sender.SendButtons(targetJID, *res.Buttons.ContentText, res.Buttons.Buttons)
			return &res, true
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

func (h *StatsHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	if len(args) > 0 {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_INCORRECT_USAGE")}
	}

	tz := h.AdminTimezone
	if tz == "" {
		tz = "Africa/Lagos"
	}
	loc, _ := time.LoadLocation(tz)
	// Calculate since midnight of the configured timezone
	since := time.Now().In(loc)
	since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, loc)

	pending, transit, err := shipUC.CountDailyStats(ctx, since.UTC())
	if err != nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_SYSTEM_ERROR"), Error: err}
	}

	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}
	msg := i18n.T(i18nLang(lang), "MSG_STATS_HEADER", company) + "\n\n━━━━━━━━━━━━━━━━━━━━━━━\n" +
		fmt.Sprintf("📦 PENDING:    *%d*\n🚚 IN TRANSIT: *%d*\n📊 TOTAL:      *%d*\n", pending, transit, pending+transit) +
		"━━━━━━━━━━━━━━━━━━━━━━━\n\n_Total operations recorded today._"
	return Result{Message: msg}
}

// InfoHandler handles !info [ID]
type InfoHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *InfoHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
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

	dbShip, err := shipUC.Track(ctx, args[0])
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

	wb := utils.GenerateWaybill(s, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}

// HelpHandler handles !help
type HelpHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *HelpHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	if isAdmin {
		// Admin Help Menu as Interactive List
		return Result{
			List: &waProto.ListMessage{
				Title:       models.StrPtr(fmt.Sprintf("%s ADMIN", company)),
				Description: models.StrPtr("Select a management action below:"),
				ButtonText:  models.StrPtr("Open Menu"),
				Sections: []*waProto.ListMessage_Section{
					{
						Title: models.StrPtr("MANAGEMENT"),
						Rows: []*waProto.ListMessage_Row{
							{Title: models.StrPtr("Stats"), Description: models.StrPtr("Daily operations summary"), RowID: models.StrPtr("!stats")},
							{Title: models.StrPtr("Status"), Description: models.StrPtr("System health & vitals"), RowID: models.StrPtr("!status")},
							{Title: models.StrPtr("Broadcast"), Description: models.StrPtr("Send msg to all groups"), RowID: models.StrPtr("!broadcast")},
						},
					},
					{
						Title: models.StrPtr("SHIPMENT CONTROL"),
						Rows: []*waProto.ListMessage_Row{
							{Title: models.StrPtr("Edit"), Description: models.StrPtr("!edit [field] [value]"), RowID: models.StrPtr("!help_edit")},
							{Title: models.StrPtr("Delete"), Description: models.StrPtr("Permanently remove shipment"), RowID: models.StrPtr("!help_delete")},
							{Title: models.StrPtr("Info"), Description: models.StrPtr("Detailed tracking data"), RowID: models.StrPtr("!info")},
						},
					},
					{
						Title: models.StrPtr("SETTINGS"),
						Rows: []*waProto.ListMessage_Row{
							{Title: models.StrPtr("Language"), Description: models.StrPtr("Switch bot language"), RowID: models.StrPtr("!lang")},
						},
					},
				},
			},
		}
	}

	// Customer Help Menu as Buttons
	return Result{
		Buttons: &waProto.ButtonsMessage{
			ContentText: models.StrPtr(fmt.Sprintf("📖 *%s CUSTOMER SERVICE*\n\nHow can we help you today?", company)),
			Buttons: []*waProto.ButtonsMessage_Button{
				{ButtonID: models.StrPtr("!info"), ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: models.StrPtr("Track Shipment")}},
				{ButtonID: models.StrPtr("!help"), ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: models.StrPtr("Instructions")}},
				{ButtonID: models.StrPtr("!lang"), ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: models.StrPtr("Change Language")}},
			},
		},
	}
}

type LangHandler struct{}

func (h *LangHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
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
	CompanyPrefix string
	AdminTimezone string
	Sender        *whatsapp.Sender
	Cfg           *config.Config
}

func (h *EditHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "✏️ *EDIT SHIPMENT*\n\nUsage: `!edit [TrackingID] [Updates...]` or `!edit [Updates...]` (targets last shipment)\n\n*Example:* `!edit LGS-1234 name: John, departure: tomorrow`"}
	}

	jid := utils.GetJID(ctx)
	var trackingID string
	var startIdx int

	// 1. Identify Target Shipment
	// Pattern: 3 uppercase letters, hyphen, 4 digits
	idPattern := regexp.MustCompile(`^[A-Z]{3}-\d{4}$`)
	if idPattern.MatchString(args[0]) {
		trackingID = args[0]
		startIdx = 1
	} else {
		// Contextual Lookup: Fetch last shipment for this user
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, jid)
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

		err := shipUC.UpdateField(ctx, trackingID, field, value)
		if err == nil {
			updatedFields = append(updatedFields, strings.ToUpper(strings.ReplaceAll(field, "_", " ")))
		}
	}

	// 4. Automatic Arrival Sync (Algorithm B)
	if departureUpdated && !arrivalExplicitlyUpdated {
		dbShip, _ := shipUC.Track(ctx, trackingID)
		if dbShip != nil {
			// Recalculate Arrival based on new Departure
			arrival, outForDelivery := shipUC.Service.CalculateArrival(newDeparture, dbShip.Origin.String, dbShip.Destination.String)

			_ = shipUC.UpdateField(ctx, trackingID, "expected_delivery_time", arrival.Format("2006-01-02 15:04:05"))
			_ = shipUC.UpdateField(ctx, trackingID, "outfordelivery_time", outForDelivery.Format("2006-01-02 15:04:05"))

			updatedFields = append(updatedFields, "EXPECTED DELIVERY TIME (AUTO-SYNC)", "OUTFORDELIVERY TIME (AUTO-SYNC)")
		}
	}

	if len(updatedFields) == 0 {
		return Result{Message: "⚠️ *UPDATE FAILED*\n_None of the fields could be updated. Check your format (e.g., label: value)._"}
	}

	// 4. Persistence & Schedule Sync
	// The DB trigger fn_shipment_auto_schedule() auto-recalculates
	// scheduled_transit_time, outfordelivery_time, expected_delivery_time
	// when origin or destination changes on UPDATE.

	dbShip, _ := shipUC.Track(ctx, trackingID)
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
			_ = shipUC.UpdateField(ctx, trackingID, "status", newStatus)

			// If it transitions, optionally trigger the notification explicitly!
			if h.Sender != nil && h.Sender.Client != nil {
				notif.SendStatusAlert(ctx, h.Sender.Client, h.Cfg, dbShip.UserJid, trackingID, newStatus, dbShip.RecipientEmail.String)
			}
		}
	}

	summary := fmt.Sprintf("✅ *INFORMATION UPDATED*\n\n🆔 *%s*\n\n📝 *FIELDS MODIFIED:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Updates have been successfully persisted to the cloud._",
		trackingID, strings.Join(updatedFields, "\n• "))

	return Result{
		Buttons: &waProto.ButtonsMessage{
			ContentText: models.StrPtr(summary),
			Buttons: []*waProto.ButtonsMessage_Button{
				{ButtonID: models.StrPtr(fmt.Sprintf("!info %s", trackingID)), ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: models.StrPtr("View Final Receipt")}},
				{ButtonID: models.StrPtr("!help"), ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: models.StrPtr("Back to Menu")}},
			},
		},
		EditID: trackingID,
	}
}

// DeleteHandler handles !delete [trackingID]
type DeleteHandler struct{}

func (h *DeleteHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "🗑️ *DELETE SHIPMENT*\n\nUsage: `!delete [TrackingID]`"}
	}

	trackingID := args[0]
	err := shipUC.Delete(ctx, trackingID)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *DELETE FAILED*\n_%v_", err)}
	}

	return Result{Message: fmt.Sprintf("🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.", trackingID)}
}

// BroadcastHandler handles !broadcast [message]
type BroadcastHandler struct {
	Sender *whatsapp.Sender
}

func (h *BroadcastHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "📣 *GLOBAL BROADCAST*\n\nUsage: `!broadcast [your message]`\n\n_This sends a message to ALL authorized groups._"}
	}

	msg := strings.Join(args, " ")
	company := h.Sender.CompanyName
	if company == "" {
		company = "LOGISTICS"
	}
	broadcastMsg := fmt.Sprintf("📢 *OFFICIAL UPDATE FROM %s*\n\n", strings.ToUpper(company)) + msg

	// Fetch authorized groups from DB
	groups, err := configUC.GetAuthorizedGroups(ctx)
	if err != nil {
		return Result{Message: i18n.T(i18nLang(lang), "ERR_DB_ERROR")}
	}

	if len(groups) == 0 {
		return Result{Message: "ℹ️ *NO TARGETS*\n\nThere are no authorized groups to broadcast to."}
	}

	// Perform broadcast in background to avoid blocking worker
	go func() {
		successCount := 0
		for _, groupID := range groups {
			groupJID, err := types.ParseJID(groupID)
			if err != nil {
				continue
			}
			// Add extra jitter between groups for safety
			time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

			h.Sender.Send(groupJID, broadcastMsg)
			successCount++
		}
		logger.Info().Int("success_count", successCount).Msg("Background broadcast complete")
	}()

	return Result{Message: fmt.Sprintf("📣 *BROADCAST INITIATED*\n\nSending to *%d* authorized groups in the background.\n\n_Progress will be logged to system vitals._", len(groups))}
}

// StatusHandler handles !status
type StatusHandler struct {
	BotPhone string
}

func (h *StatusHandler) Execute(ctx context.Context, shipUC *usecase.ShipmentUsecase, configUC *usecase.ConfigUsecase, args []string, lang string, isAdmin bool) Result {
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

	groupsCount, _ := configUC.CountAuthorizedGroups(ctx)

	msg := i18n.T(i18nLang(lang), "MSG_STATUS_DASHBOARD") + "\n\n" +
		fmt.Sprintf("📊 UPTIME:    *%s*\n", uptime.Truncate(time.Second)) +
		fmt.Sprintf("🔋 MEMORY:    *%d MB* / 1024 MB\n", memMB) +
		fmt.Sprintf("🗄️ DATABASE:  *%s*\n", dbStatus) +
		fmt.Sprintf("👥 GROUPS:    *%d authorized*\n", groupsCount) +
		fmt.Sprintf("📦 PROCESSED: *%d jobs* (%d success)\n\n", jobs, success) +
		"_System is running within safe 1GB RAM margins._"

	return Result{Message: msg}
}
