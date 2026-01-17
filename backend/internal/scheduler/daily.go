package scheduler

import (
	"context"
	"fmt"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/supabase"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

func StartDailySummary(client *whatsmeow.Client, db *supabase.Client, timezone string, allowedGroups []string) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		logger.Warn().Str("timezone", timezone).Err(err).Msg("Invalid timezone, using UTC")
		location = time.UTC
	}

	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now().In(location)
			if now.Hour() == 23 && now.Minute() == 0 {
				logger.Info().Msg("Triggering daily summary")
				pending, transit, err := db.GetTodayStats(location)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to get daily stats")
					continue
				}
				count := pending + transit

				msg := fmt.Sprintf("📊 *Daily Summary*\nYou processed *%d* packages today.", count)

				for _, groupID := range allowedGroups {
					targetJID, _ := types.ParseJID(groupID)
					if targetJID.IsEmpty() {
						targetJID = types.NewJID(groupID, types.GroupServer)
					}

					content := &waProto.Message{Conversation: models.StrPtr(msg)}
					_, _ = client.SendMessage(context.Background(), targetJID, content)
				}
			}
		}
	}()
}
