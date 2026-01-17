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
		// Stats doesn't need info, but Info does
		if infoH, ok := handler.(*InfoHandler); ok {
			infoH.CompanyName = d.CompanyName
		}
		res := handler.Execute(ctx, d.db, args)
		return res.Message, true
	}

	return "", false
}

func presentsAsCommand(text string) bool {
	return len(text) > 1 && text[0] == '!'
}

// StatsHandler handles !stats
type StatsHandler struct{}

func (h *StatsHandler) Execute(ctx context.Context, db *supabase.Client, args []string) Result {
	loc, _ := time.LoadLocation("Africa/Lagos")
	pending, transit, err := db.GetTodayStats(loc)
	if err != nil {
		return Result{Message: "❌ System Error: Could not fetch statistics.", Error: err}
	}

	msg := fmt.Sprintf("📊 *Today's Logistics*\n\n• PENDING: %d\n• IN_TRANSIT: %d\n\n_Total Created Today: %d_", pending, transit, pending+transit)
	return Result{Message: msg}
}

// InfoHandler handles !info [ID]
type InfoHandler struct {
	CompanyName string
}

func (h *InfoHandler) Execute(ctx context.Context, db *supabase.Client, args []string) Result {
	if len(args) < 1 {
		return Result{Message: "💡 *Usage Guide*\nPlease provide a tracking ID:\n`!INFO [ID]`"}
	}

	shipment, err := db.GetShipment(args[0])
	if err != nil {
		return Result{Message: "❌ System Error: Database lookup failed.", Error: err}
	}
	if shipment == nil {
		return Result{Message: fmt.Sprintf("⚠️ *Not Found*\nTracking ID *%s* does not exist in our system.", args[0])}
	}

	wb := utils.GenerateWaybill(*shipment, h.CompanyName)
	return Result{Message: "```\n" + wb + "\n```"}
}
