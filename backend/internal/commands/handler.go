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
	Message string
	Error   error
}

// Handler defines the interface all bot commands must implement.
type Handler interface {
	Execute(ctx context.Context, db *supabase.Client, args []string) Result
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
}

func (d *Dispatcher) Dispatch(ctx context.Context, text string) (string, bool) {
	if !presentsAsCommand(text) {
		return "", false
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", false
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
		}
		res := handler.Execute(ctx, d.db, args)
		return res.Message, true
	}

	return "", false
}

func presentsAsCommand(text string) bool {
	return len(text) > 1 && (text[0] == '!' || text[0] == '#')
}

// StatsHandler handles !stats
type StatsHandler struct {
	CompanyName string
}

func (h *StatsHandler) Execute(ctx context.Context, db *supabase.Client, args []string) Result {
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

func (h *InfoHandler) Execute(ctx context.Context, db *supabase.Client, args []string) Result {
	if len(args) < 1 {
		company := strings.ToUpper(h.CompanyName)
		if company == "" {
			company = "COMMAND"
		}
		msg := fmt.Sprintf("🚀 *%s COMMAND CENTER*\n\n", company) +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"1️⃣ `!stats` - Daily Operations\n" +
			"2️⃣ `!info [TrackingID]` - Shipment Tracker\n" +
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

func (h *HelpHandler) Execute(ctx context.Context, db *supabase.Client, args []string) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "LOGISTICS"
	}

	msg := fmt.Sprintf("📖 *%s SERVICE MENU*\n\n", company) +
		"━━━━ COMMANDS ━━━━\n" +
		"1️⃣ `!stats` - Daily Operations\n" +
		"2️⃣ `!info [ID]` - Check Status\n" +
		"3️⃣ `!help` - Show this menu\n" +
		"━━━━━━━━━━━━━━━━━━━\n\n" +
		"*📦 HOW TO REGISTER A PACKAGE:*\n" +
		"_Simply send a message with these details:_\n\n" +
		"Sender: John Doe\n" +
		"Receiver Name: Jane Smith\n" +
		"Receiver Phone: +234 800 123 4567\n" +
		"Receiver Address: 123 Main St, Lagos\n" +
		"Receiver Country: Nigeria\n" +
		"Sender Country: UK\n\n" +
		"*PRO TIP:* _You can use shortcuts like #stats or #info._"

	return Result{Message: msg}
}
