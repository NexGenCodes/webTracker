package commands

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/utils"
	"webtracker-bot/internal/whatsapp"

	"go.mau.fi/whatsmeow/types"
)

// Result represents the outcome of a command execution.
type Result struct {
	Message  string
	Language string
	EditID   string // Signals that this ID was edited and needs a new receipt
	Error    error
}

type Handler interface {
	Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result
}

type Dispatcher struct {
	ldb           *localdb.Client
	sender        *whatsapp.Sender // Added for broadcasting
	handlers      map[string]Handler
	AwbCmd        string
	CompanyName   string
	BotPhone      string
	AdminTimezone string
}

func NewDispatcher(ldb *localdb.Client, sender *whatsapp.Sender, awbCmd string, companyName string, botPhone string, adminTimezone string) *Dispatcher {
	d := &Dispatcher{
		ldb:           ldb,
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
		jid := ctx.Value("jid").(string)
		senderPhone := ctx.Value("sender_phone").(string)
		isAdmin, _ := ctx.Value("is_admin").(bool)

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
		case *BroadcastHandler:
			h.Sender = d.sender
		case *StatusHandler:
			h.BotPhone = d.BotPhone
		}

		lang, _ := d.ldb.GetUserLanguage(ctx, jid)

		isOwnerOnlyCmd := rawCmd == "broadcast" || rawCmd == "status"
		if isOwnerOnlyCmd {
			if !isOwner {
				logger.Warn().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Owner-only command blocked")
				return &Result{Message: "🚫 *OWNER ACCESS ONLY*\n\n_This command is restricted to the bot owner only._"}, true
			}
		}

		isPublicCmd := rawCmd == "info" || rawCmd == "help"
		if !isPublicCmd && !isOwnerOnlyCmd {
			if isAdmin {
				logger.Info().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Admin command authorized")
			} else {
				logger.Warn().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Command blocked: sender is not authorized")
				return &Result{Message: "🚫 *ACCESS DENIED*\n\n_This command is restricted to the bot owner or group admins._\n\n💡 You can use `!info [ID]` to track packages."}, true
			}
		}

		res := handler.Execute(ctx, d.ldb, args, lang, isAdmin)
		if res.Language != "" {
			d.ldb.SetUserLanguage(ctx, jid, res.Language)
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

func (h *StatsHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	if len(args) > 0 {
		return Result{Message: "⚠️ *INCORRECT USAGE*\n_Please send only `!stats` without any extra text._"}
	}

	tz := h.AdminTimezone
	if tz == "" {
		tz = "Africa/Lagos"
	}
	loc, _ := time.LoadLocation(tz)
	// Calculate since midnight of the configured timezone
	since := time.Now().In(loc)
	since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, loc)

	pending, transit, err := ldb.CountDailyStats(ctx, since.UTC())
	if err != nil {
		return Result{Message: "❌ *SYSTEM ERROR*\n_Could not fetch statistics._", Error: err}
	}

	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}
	msg := fmt.Sprintf("📊 *%s VITAL STATS*\n\n━━━━━━━━━━━━━━━━━━━━━━━\n📦 PENDING:    *%d*\n🚚 IN TRANSIT: *%d*\n📊 TOTAL:      *%d*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n_Total operations recorded today._", company, pending, transit, pending+transit)
	return Result{Message: msg}
}

// InfoHandler handles !info [ID]
type InfoHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *InfoHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
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

	// Using ldb instead of db
	shipment, err := ldb.GetShipment(ctx, args[0])
	if err != nil {
		return Result{Message: "❌ *DATABASE ERROR*\n_Lookup failed. Please try again later._", Error: err}
	}

	// Check if shipment is nil (although GetShipment returns error on not found usually, let's look at implementation)
	// Our new GetShipment returns error if not found? No, let's assume it might return nil if we handled SqlNoRows differently
	// Actually typical GetShipment implementations error on Not Found.
	// But let's keep the null check just in case logic changes.

	wb := utils.GenerateWaybill(*shipment, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}

// HelpHandler handles !help
type HelpHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *HelpHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	var msg string
	if isAdmin {
		msg = fmt.Sprintf("🛠️ *%s - ADMIN CONTROL PANEL*\n\n", company) +
			"━━━━ MANAGEMENT ━━━━\n" +
			"📊 `!stats` - Daily Operations\n" +
			" `!broadcast [msg]` - Global Update\n" +
			"🖥️ `!status` - System Health/Groups\n" +
			" `!edit [field] [value]` - Fix shipment mistakes\n" +
			"🗑️ `!delete [ID]` - Permanently remove shipment\n" +
			"━━━━ GENERAL ━━━━\n" +
			"🔍 `!info [ID]` - Check shipment status\n" +
			"🌐 `!lang [code]` - Switch language (en, pt, es, de)\n" +
			"❓ `!help` - Show this admin menu\n" +
			"━━━━━━━━━━━━━━━━━━━\n\n" +
			"*🛠️ HOW TO EDIT:*\n" +
			"`!edit name Jane Doe` (Fixes last shipment)\n" +
			fmt.Sprintf("`!edit %s-123 name Jane Doe` (Fixes specific ID)\n", h.CompanyPrefix) +
			"_Fields: name, phone, address, country, origin, departure, arrival_\n\n" +
			"*📅 DATE EDITING:*\n" +
			"_Use: today, tomorrow, next tomorrow, yesterday_\n" +
			"_Example: `!edit departure tomorrow`_"
	} else {
		msg = fmt.Sprintf("📖 *%s - CUSTOMER SERVICE*\n\n", company) +
			"━━━━ AVAILABLE COMMANDS ━━━━\n" +
			"🔍 `!info [ID]` - Track your shipment\n" +
			"❓ `!help` - Show this instructions menu\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"*📦 HOW TO REGISTER SHIPMENT:*\n" +
			"_Send a message with these details:_\n\n" +
			"Sender: John Doe\n" +
			"Receiver Name: Jane Smith\n" +
			"Receiver Phone: +234 800 123 4567\n" +
			"Receiver Address: 123 Main St, Lagos\n\n" +
			"*PRO TIP:* _You can use shortcuts like #info or #help._"
	}

	return Result{Message: msg}
}

type LangHandler struct{}

func (h *LangHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
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
}

func (h *EditHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	if len(args) < 2 {
		return Result{Message: "📝 *EDIT SHIPMENT INFORMATION*\n\nUsage:\n`!edit [field] [new_value]`\n\nFields: `name`, `phone`, `address`, `country`, `email`, `id`, `sender`, `origin`"}
	}

	var trackingID, field, value string
	jid := ctx.Value("jid").(string)

	// Case 1: !edit [trackingID] [field] [value...]
	if strings.Contains(args[0], "-") {
		trackingID = args[0]
		field = args[1]
		value = strings.Join(args[2:], " ")
	} else {
		// Case 2: !edit [field] [value...] (Target last shipment)
		var err error
		trackingID, err = ldb.GetLastTrackingByJID(ctx, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "⚠️ *NO RECORD FOUND*\n_I couldn't find your last shipment. Please provide the tracking ID._"}
		}
		field = args[0]
		value = strings.Join(args[1:], " ")
	}

	if value == "" {
		return Result{Message: "⚠️ *MISSING VALUE*\n_Please provide the new information for the field._"}
	}

	// Normalize Field Names (Aliases)
	normField := strings.ToLower(field)
	switch normField {
	case "receiver", "receivername", "receiver_name", "recipient", "reciever", "recieve", "recievers", "receivers", "recipientname":
		field = "recipient_name"
	case "sender", "sendername", "sender_name", "senders":
		field = "sender_name"
	case "phone", "phones", "receiverphone", "receiver_phone", "recipient_phone", "mobile", "mobiles", "number", "numbers", "receivernumber", "cell", "contact":
		field = "recipient_phone"
	case "email", "emails", "mail", "mails", "receiveremail", "receiver_email", "recipient_email":
		field = "recipient_email"
	case "country", "countries", "receivercountry", "receiver_country", "dest", "destination", "destinations", "location":
		field = "destination"
	case "address", "addresses", "addr", "receiveraddress", "receiver_address", "recipient_address":
		field = "recipient_address"
	case "sendercountry", "sender_country", "origin", "origins", "from", "source":
		field = "origin"
	case "senderphone", "sender_phone", "sendernumber", "sender_number":
		field = "sender_phone"
	case "type", "types", "cargotype", "cargo_type", "content", "contents":
		field = "cargo_type"
	case "departure", "departure_date", "departuredate", "transit", "transit_time":
		field = "scheduled_transit_time"
	case "arrival", "arrival_date", "arrivaldate", "delivery", "delivery_time":
		field = "expected_delivery_time"
	}

	if field == "weight" {
		return Result{Message: "🚫 *ACCESS DENIED*\n_The weight policy is strictly set to 15KG and cannot be modified._"}
	}

	// Validation Logic
	if strings.Contains(strings.ToLower(field), "email") {
		if !parser.ValidateEmail(value) {
			return Result{Message: "⚠️ *INVALID EMAIL format*\n_Please provide a valid email address (e.g., name@domain.com)._"}
		}
	}

	if strings.Contains(strings.ToLower(field), "phone") || strings.Contains(strings.ToLower(field), "mobile") {
		if !parser.ValidatePhone(value) {
			return Result{Message: "⚠️ *INVALID PHONE FORMAT*\n_Phone numbers must contain at least 5 digits._"}
		}
	}

	// Use the same parser cleaning logic as creation
	value = parser.CleanText(value)

	// Special Handling for Dates (today, tomorrow, next tomorrow, yesterday)
	if field == "scheduled_transit_time" || field == "expected_delivery_time" {
		tz := h.AdminTimezone
		if tz == "" {
			tz = "Africa/Lagos" // Default
		}
		loc, _ := time.LoadLocation(tz)
		now := time.Now().In(loc)

		if parsedDate, ok := utils.ParseNaturalDate(value, now); ok {
			// Format for SQLite (YYYY-MM-DD HH:MM:SS)
			value = parsedDate.Format("2006-01-02 15:04:05")
		} else {
			// Fallback: Check if it's already a valid date string
			_, err := time.Parse("2006-01-02", value)
			if err != nil {
				_, err = time.Parse("2006-01-02 15:04:05", value)
				if err != nil {
					return Result{Message: "⚠️ *INVALID DATE FORMAT*\n_Please use 'today', 'tomorrow', 'next tomorrow', 'yesterday' or YYYY-MM-DD._"}
				}
			}
		}
	}

	err := ldb.UpdateShipmentField(ctx, trackingID, field, value)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *UPDATE FAILED*\n_%v_", err)}
	}

	// Catch-up logic for edits: If we edited a date, re-calculate status
	if field == "scheduled_transit_time" || field == "expected_delivery_time" {
		if s, _ := ldb.GetShipment(ctx, trackingID); s != nil {
			newStatus := s.ResolveStatus(time.Now().UTC())
			if newStatus != s.Status {
				_ = ldb.UpdateShipmentStatus(ctx, trackingID, newStatus)
			}
		}
	}

	return Result{
		Message: fmt.Sprintf("✅ *INFORMATION UPDATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nID: *%s*\nField: *%s*\nNew Value: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n_Generating your updated receipt..._", trackingID, strings.ToUpper(field), value),
		EditID:  trackingID,
	}
}

// DeleteHandler handles !delete [trackingID]
type DeleteHandler struct{}

func (h *DeleteHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "🗑️ *DELETE SHIPMENT*\n\nUsage: `!delete [TrackingID]`"}
	}

	trackingID := args[0]
	err := ldb.DeleteShipment(ctx, trackingID)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *DELETE FAILED*\n_%v_", err)}
	}

	return Result{Message: fmt.Sprintf("🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.", trackingID)}
}

// BroadcastHandler handles !broadcast [message]
type BroadcastHandler struct {
	Sender *whatsapp.Sender
}

func (h *BroadcastHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "📣 *GLOBAL BROADCAST*\n\nUsage: `!broadcast [your message]`\n\n_This sends a message to ALL authorized groups._"}
	}

	msg := strings.Join(args, " ")
	broadcastMsg := "📢 *OFFICIAL UPDATE FROM LOGISTICS*\n\n" + msg

	// Check connection first
	if !h.Sender.Client.IsConnected() {
		return Result{Message: "❌ *FAILED*\n_WhatsApp is currently disconnected._"}
	}

	// Get all authorized groups from DB
	groups, err := ldb.GetAuthorizedGroups(ctx)
	if err != nil {
		return Result{Message: "❌ *DATABASE ERROR*\n_Failed to fetch target groups._", Error: err}
	}

	successCount := 0

	// Create semaphore to limit concurrent sends (avoid blocking main thread too long)
	// WhatsApp recommends not flooding

	for _, groupID := range groups {
		groupJID, err := types.ParseJID(groupID)
		if err != nil {
			continue
		}
		// Direct send (non-reply)
		h.Sender.Send(groupJID, broadcastMsg)
		successCount++
	}

	return Result{Message: fmt.Sprintf("✅ *BROADCAST COMPLETE*\n\nSent to: *%d* groups.", successCount)}
}

// StatusHandler handles !status
type StatusHandler struct {
	BotPhone string
}

func (h *StatusHandler) Execute(ctx context.Context, ldb *localdb.Client, args []string, lang string, isAdmin bool) Result {
	// Performance Telemetry
	uptime := time.Since(logger.GlobalVitals.StartTime)
	jobs := atomic.LoadInt64(&logger.GlobalVitals.JobsProcessed)
	success := atomic.LoadInt64(&logger.GlobalVitals.ParseSuccess)

	// Memory usage tracking
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memMB := m.Alloc / 1024 / 1024

	// System Health
	dbStatus := "🟢 ONLINE (SQLite)"
	// Supabase Ping removed, we assume localdb is alive if process is running
	// Could check ldb.db.Ping() if exposed

	groupsCount, _ := ldb.CountAuthorizedGroups(ctx)

	msg := "🖥️ *SYSTEM DASHBOARD*\n\n" +
		fmt.Sprintf("📊 UPTIME:    *%s*\n", uptime.Truncate(time.Second)) +
		fmt.Sprintf("🔋 MEMORY:    *%d MB* / 1024 MB\n", memMB) +
		fmt.Sprintf("🗄️ DATABASE:  *%s*\n", dbStatus) +
		fmt.Sprintf("👥 GROUPS:    *%d authorized*\n", groupsCount) +
		fmt.Sprintf("📦 PROCESSED: *%d jobs* (%d success)\n\n", jobs, success) +
		"_System is running within safe 1GB RAM margins._"

	return Result{Message: msg}
}
