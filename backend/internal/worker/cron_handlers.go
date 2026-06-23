package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"go.mau.fi/whatsmeow/types"

	"webtracker-bot/internal/billing"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/tasks"
)

// HandleCronPulse acts as a dispatcher, enqueuing a per-company pulse task to parallelize status processing.
func (w *Worker) HandleCronPulse(ctx context.Context, t *asynq.Task) error {
	companies, err := w.ConfigUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Pulse Dispatcher: Failed to get companies")
		return err
	}

	if w.AsynqClient == nil {
		return fmt.Errorf("Asynq client not initialized in worker")
	}

	for _, companyID := range companies {
		if err := tasks.EnqueueCronCompanyPulse(w.AsynqClient, companyID); err != nil {
			logger.Error().Err(err).Str("company", companyID.String()).Msg("Pulse Dispatcher: Failed to enqueue company pulse")
		}
	}

	logger.Info().Int("count", len(companies)).Msg("Pulse Dispatcher: Fanned out company pulse tasks")
	return nil
}

// HandleCronCompanyPulse processes status transitions for a single company.
// This is now fanned out and runs in parallel across the worker fleet.
func (w *Worker) HandleCronCompanyPulse(ctx context.Context, t *asynq.Task) error {
	var payload tasks.CronCompanyPulsePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to decode company pulse payload: %w", asynq.SkipRetry)
	}

	companyID := payload.CompanyID
	now := time.Now().UTC()

	// 1. Subscription Check (Proactive Deactivation)
	// Regular admins should be disconnected if their trial ends, but the super admin remains free.
	company, err := w.ConfigUC.GetCompanyByID(ctx, companyID)
	if err == nil {
		isSuperAdmin, _ := billing.IsSuperAdminByEmail(ctx, w.ConfigUC, company.AdminEmail)
		if !isSuperAdmin {
			expired := company.SubscriptionExpiry.Valid && company.SubscriptionExpiry.Time.Before(now)
			inactive := company.SubscriptionStatus.String != "active" && company.SubscriptionStatus.String != "trialing"

			if expired || inactive {
				logger.Warn().
					Str("company", companyID.String()).
					Str("email", company.AdminEmail).
					Msg("Subscription expired or inactive. Proactively deactivating bot.")

				// Forcefully disconnect/deactivate the bot instance
				if err := w.Bots.DeactivateBot(companyID); err != nil {
					logger.Error().Err(err).Str("company", companyID.String()).Msg("Cron pulse: failed to deactivate bot for expired subscription")
				}
				return nil // Stop processing for this company
			}
		}
	}

	// Multi-round processing (up to 3 rounds) to catch cascading transitions
	// (e.g. OutForDelivery -> Delivered if the logic allows it, though usually 1-by-1)
	maxRounds := 3
	for i := 0; i < maxRounds; i++ {
		transitions, err := w.ShipmentUC.ProcessTransitions(ctx, companyID, now)
		if err != nil {
			logger.Error().Err(err).Str("company", companyID.String()).Msg("Company Pulse: Failed to process transitions")
			return err
		}

		if len(transitions) == 0 {
			break
		}

		bot, err := w.Bots.GetBot(companyID)
		if err != nil {
			// Try to activate if not found
			if err := w.Bots.ActivateBot(ctx, companyID); err != nil {
				logger.Warn().Err(err).Str("company", companyID.String()).Msg("Company Pulse: Skipping alert, bot activation failed")
				continue
			}
			var getErr error
			bot, getErr = w.Bots.GetBot(companyID)
			if getErr != nil {
				logger.Error().Err(getErr).Str("company", companyID.String()).Msg("Company Pulse: Failed to get bot after activation")
				continue
			}
		}

		if bot == nil {
			logger.Warn().Str("company", companyID.String()).Msg("Company Pulse: Bot instance not found after activation attempt")
			continue
		}

		for _, tr := range transitions {
			logger.Info().Str("id", tr.TrackingID).Str("new_status", tr.NewStatus).Msg("Company Pulse: Shipment status updated")

			// Always use the Asynq-backed OutboundAlert task for high reliability
			if w.AsynqClient != nil {
				if err := tasks.EnqueueOutboundAlert(w.AsynqClient, companyID, bot.GetCompanyName(), tr.UserJID, tr.TrackingID, tr.NewStatus, tr.RecipientEmail); err != nil {
					logger.Error().Err(err).Str("id", tr.TrackingID).Msg("Company Pulse: Failed to enqueue outbound alert task")
				}
			} else {
				// Fallback to direct sender queue if Asynq is unavailable (emergency path)
				shipData, _ := w.ShipmentUC.Track(ctx, companyID, tr.TrackingID)
				dest := ""
				if shipData != nil {
					dest = shipData.Destination.String
				}
				notif.SendStatusAlert(ctx, bot.GetSender(), w.Cfg, bot.GetCompanyName(), tr.UserJID, tr.TrackingID, tr.NewStatus, tr.RecipientEmail, dest)
			}
		}
	}

	return nil
}

// HandleCronDailyStats processes daily stats report
func (w *Worker) HandleCronDailyStats(ctx context.Context, t *asynq.Task) error {
	since := time.Now().Add(-24 * time.Hour)

	companies, err := w.ConfigUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Stats: Failed to get companies")
		return err
	}

	for _, companyID := range companies {
		created, delivered, err := w.ShipmentUC.CountDailyStats(ctx, companyID, since)
		if err != nil {
			logger.Error().Err(err).Msg("Stats: Failed to count")
			continue
		}

		msg := fmt.Sprintf("📊 *DAILY STATS*\n\n✅ Created: %d\n📦 Delivered: %d", created, delivered)

		bot, err := w.Bots.GetBot(companyID)
		if err != nil {
			if actErr := w.Bots.ActivateBot(ctx, companyID); actErr != nil {
				logger.Warn().Err(actErr).Str("company", companyID.String()).Msg("Stats: failed to activate bot")
				continue
			}
			var getErr error
			bot, getErr = w.Bots.GetBot(companyID)
			if getErr != nil {
				logger.Warn().Err(getErr).Str("company", companyID.String()).Msg("Stats: failed to get bot after activation")
				continue
			}
		}
		if bot == nil {
			logger.Warn().Str("company", companyID.String()).Msg("Stats: bot is nil after activation")
			continue
		}

		wc := bot.GetWAClient()
		if wc == nil || wc.Store == nil || wc.Store.ID == nil {
			logger.Warn().Str("company", companyID.String()).Msg("Stats: Skipping report, bot session not initialized")
			continue
		}

		groups, _ := w.ConfigUC.GetAuthorizedGroups(ctx, companyID)
		for _, g := range groups {
			jid, err := types.ParseJID(g)
			if err != nil {
				continue
			}
			bareJid := types.JID{User: jid.User, Server: jid.Server}
			bot.GetSender().Send(bareJid, msg)
			logger.Debug().Str("group", g).Msg("Stats: Daily report enqueued")
		}
	}
	return nil
}

// HandleCronPruning processes daily aged data cleanup
func (w *Worker) HandleCronPruning(ctx context.Context, t *asynq.Task) error {
	deliveredCutoff := time.Now().AddDate(0, 0, -7)
	allCutoff := time.Now().AddDate(0, 0, -14)

	companies, err := w.ConfigUC.GetAllCompanies(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Pruning: Failed to get companies")
		return err
	}

	for _, companyID := range companies {
		deleted, err := w.ShipmentUC.RunAgedCleanup(ctx, companyID, deliveredCutoff, allCutoff)
		if err != nil {
			logger.Error().Err(err).Msg("Pruning: Failed to run aged cleanup")
			continue
		}
		logger.Info().Int64("deleted_count", deleted).Msg("Pruning: Aged cleanup completed successfully")
	}
	return nil
}

// HandleCronHealthCheck pings the external healthcheck
func (w *Worker) HandleCronHealthCheck(ctx context.Context, t *asynq.Task) error {
	if w.Cfg.HealthcheckURL == "" {
		return nil
	}
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(w.Cfg.HealthcheckURL)
	if err != nil {
		logger.Error().Err(err).Msg("Health check ping failed")
		return err
	}
	return nil
}

// HandleCronBotLiveness triggers bot reconnections if down
func (w *Worker) HandleCronBotLiveness(ctx context.Context, t *asynq.Task) error {
	w.Bots.LivenessCheck()
	return nil
}

// HandleOutboundAlert processes outbound WhatsApp notifications reliably
func (w *Worker) HandleOutboundAlert(ctx context.Context, t *asynq.Task) error {
	var payload tasks.OutboundAlertPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to decode outbound alert payload: %w", asynq.SkipRetry)
	}

	bot, err := w.Bots.GetBot(payload.CompanyID)
	if err != nil {
		// Just re-activate and try to get it again
		if actErr := w.Bots.ActivateBot(ctx, payload.CompanyID); actErr != nil {
			logger.Error().Err(actErr).Str("company_id", payload.CompanyID.String()).Msg("Outbound alert: failed to reactivate bot")
		}
		bot, err = w.Bots.GetBot(payload.CompanyID)
		if err != nil || bot == nil {
			return fmt.Errorf("bot instance not found or failed to hydrate: %w", asynq.SkipRetry)
		}
	}

	wc := bot.GetWAClient()
	if wc == nil || wc.Store.ID == nil {
		return fmt.Errorf("bot session not initialized")
	}

	shipmentData, _ := w.ShipmentUC.Track(ctx, payload.CompanyID, payload.TrackingID)
	dest := ""
	if shipmentData != nil {
		dest = shipmentData.Destination.String
	}

	// Enqueues to the jitter queue via Sender (non-blocking, retried by Asynq on failure)
	notif.SendStatusAlert(ctx, bot.GetSender(), w.Cfg, payload.CompanyName, payload.JIDStr, payload.TrackingID, payload.Status, payload.Email, dest)

	return nil
}
