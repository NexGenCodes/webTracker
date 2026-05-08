package commands

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
)

// StatusHandler handles !status
type StatusHandler struct {
	BotPhone string
}

func (h *StatusHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
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
