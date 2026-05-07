package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/models"
)

// HelpHandler handles !help
type HelpHandler struct {
	CompanyName   string
	CompanyPrefix string
}

func (h *HelpHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	company := strings.ToUpper(h.CompanyName)
	if company == "" {
		company = "airwaybill"
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
