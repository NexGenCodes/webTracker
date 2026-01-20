package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/utils"
)

// Result represents the outcome of a command execution.
type Result struct {
	Message  string
	Language string
	EditID   string // Signals that this ID was edited and needs a new receipt
	Error    error
}

// Handler defines the interface all bot commands must implement.
type Handler interface {
	Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result
}

// Dispatcher routes messages starting with "!" to the appropriate handler.
type Dispatcher struct {
	db          *supabase.Client
	handlers    map[string]Handler
	AwbCmd      string
	CompanyName string
}

func NewDispatcher(db *supabase.Client, awbCmd string, companyName string) *Dispatcher {
	d := &Dispatcher{
		db:          db,
		handlers:    make(map[string]Handler),
		AwbCmd:      awbCmd,
		CompanyName: companyName,
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
		// Stats, Info, and Help all need branding
		switch h := handler.(type) {
		case *StatsHandler:
			h.CompanyName = d.CompanyName
		case *InfoHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *HelpHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *EditHandler:
			h.CompanyPrefix = d.AwbCmd
		}

		jid := ctx.Value("jid").(string)
		lang, _ := d.db.GetUserLanguage(ctx, jid)

		res := handler.Execute(ctx, d.db, args, lang)
		if res.Language != "" {
			d.db.SetUserLanguage(ctx, jid, res.Language)
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
	CompanyName string
}

func (h *StatsHandler) Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result {
	if len(args) > 0 {
		return Result{Message: "⚠️ *INCORRECT USAGE*\n_Please send only `!stats` without any extra text._"}
	}

	loc, _ := time.LoadLocation("Africa/Lagos")
	pending, transit, err := db.GetTodayStats(loc)
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

func (h *InfoHandler) Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result {
	if len(args) < 1 {
		company := strings.ToUpper(h.CompanyName)
		if company == "" {
			company = "COMMAND"
		}
		msg := fmt.Sprintf("🚀 *%s COMMAND CENTER*\n\n", company) +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"1️⃣ `!stats` - Daily Operations\n" +
			"2️⃣ `!info [TrackingID]` - Shipment Information Tracker\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"*PRO TIP:*\n" +
			fmt.Sprintf("_Use `!info %s-123456789` for full details._", h.CompanyPrefix)
		return Result{Message: msg}
	}

	shipment, err := db.GetShipment(ctx, args[0])
	if err != nil {
		return Result{Message: "❌ *DATABASE ERROR*\n_Lookup failed. Please try again later._", Error: err}
	}

	if shipment == nil {
		return Result{Message: fmt.Sprintf("⚠️ *RECORD NOT FOUND*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nID: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n_This tracking ID does not exist in our registry._", args[0])}
	}

	wb := utils.GenerateWaybill(*shipment, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}

// HelpHandler handles !help
type HelpHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *HelpHandler) Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	msg := fmt.Sprintf("📖 *%s SERVICE MENU*\n\n", company) +
		"━━━━ COMMANDS ━━━━\n" +
		"1️⃣ `!stats` - Daily Operations\n" +
		"2️⃣ `!info [ID]` - Check Status\n" +
		"3️⃣ `!edit [field] [value]` - Fix mistakes\n" +
		"4️⃣ `!help` - Show this menu\n" +
		"━━━━━━━━━━━━━━━━━━━\n\n" +
		"*📦 HOW TO REGISTER SHIPMENT INFORMATION:*\n" +
		"_Simply send a message with these details:_\n\n" +
		"Sender: John Doe\n" +
		"Receiver Name: Jane Smith\n" +
		"Receiver Phone: +234 800 123 4567\n" +
		"Receiver Address: 123 Main St, Lagos\n" +
		"Receiver Country: Nigeria\n" +
		"Sender Country: UK\n\n" +
		"*🛠️ HOW TO EDIT:*\n" +
		"`!edit name Jane Doe` (Fixes last shipment)\n" +
		fmt.Sprintf("`!edit %s-123 name Jane Doe` (Fixes specific ID)\n\n", h.CompanyPrefix) +
		"*PRO TIP:* _You can use shortcuts like #stats or #edit._"

	return Result{Message: msg}
}

type LangHandler struct{}

func (h *LangHandler) Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result {
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
}

func (h *EditHandler) Execute(ctx context.Context, db *supabase.Client, args []string, lang string) Result {
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
		trackingID, err = db.GetLastTrackingByJID(ctx, jid)
		if err != nil || trackingID == "" {
			return Result{Message: "⚠️ *NO RECORD FOUND*\n_I couldn't find your last shipment. Please provide the tracking ID._"}
		}
		field = args[0]
		value = strings.Join(args[1:], " ")
	}

	if value == "" {
		return Result{Message: "⚠️ *MISSING VALUE*\n_Please provide the new information for the field._"}
	}

	err := db.UpdateShipmentField(ctx, trackingID, field, value)
	if err != nil {
		return Result{Message: fmt.Sprintf("❌ *UPDATE FAILED*\n_%v_", err)}
	}

	return Result{
		Message: fmt.Sprintf("✅ *INFORMATION UPDATED*\n\n━━━━━━━━━━━━━━━━━━━━━━━\nID: *%s*\nField: *%s*\nNew Value: *%s*\n━━━━━━━━━━━━━━━━━━━━━━━\n\n_Generating your updated receipt..._", trackingID, strings.ToUpper(field), value),
		EditID:  trackingID,
	}
}
