package commands

import (
	"context"
	"fmt"
	"strings"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/models"

	"github.com/google/uuid"
)

// StatsHandler handles !stats
type StatsHandler struct {
	CompanyName   string
	AdminTimezone string
}

func (h *StatsHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
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
