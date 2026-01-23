package whatsapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/supabase"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "modernc.org/sqlite"
)

func NewClient(dbPath string) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// SQLite connection string with performance optimizations:
	// - WAL mode for better concurrency and speed
	// - Synchronous NORMAL for speed/durability balance
	// - Cache size -2000 (approx 2MB)
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-2000)&_pragma=foreign_keys(1)", dbPath)
	container, err := sqlstore.New(context.Background(), "sqlite", dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	client := whatsmeow.NewClient(device, waLog.Stdout("whatsapp", "INFO", true))
	logger.Info().Msg("WhatsApp client initialized and session loaded")
	return client, nil
}

// GetBarePhone extracts only the digits before any device/suffix markers.
// e.g. "23480...0:12" -> "23480...0"
func GetBarePhone(jid string) string {
	if jid == "" {
		return ""
	}
	re := regexp.MustCompile(`^(\d+)`)
	match := re.FindString(jid)
	return match
}

func HandleEvent(client *whatsmeow.Client, evt interface{}, queue chan<- models.Job, cfg *config.Config, db *supabase.Client, ldb *localdb.Client) {
	switch v := evt.(type) {
	case *events.Message:
		text := ""
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage().GetText() != "" {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		if text == "" {
			return
		}

		// Identify Chat / Sender Context
		chatJID := v.Info.Chat
		isGroup := chatJID.Server == "g.us"

		senderPhone := GetBarePhone(v.Info.Sender.User)

		// 1. Identify Bot Identity (Phone & LID)
		botPhone := ""
		if client.Store.ID != nil {
			botPhone = GetBarePhone(client.Store.ID.User)
		}

		// 2. Persistent Bot LID Mapping
		botLID, _ := ldb.GetSystemConfig(context.Background(), "bot_lid")
		if v.Info.IsFromMe && client.Store.ID != nil {
			senderPhone = botPhone
			newLID := GetBarePhone(v.Info.Sender.User)
			if newLID != "" && newLID != botLID {
				botLID = newLID
				_ = ldb.SetSystemConfig(context.Background(), "bot_lid", botLID)
				logger.Info().Str("botLID", botLID).Msg("[RBAC DIAGNOSTIC] Discovered and persisted bot LID")
			}
		}

		isAuthorized := false  // Bot's permission in group
		isSenderAdmin := false // Sender's permission to control bot

		if isGroup {
			cachedAuth, cached, _ := ldb.GetGroupAuthority(context.Background(), chatJID.String())

			// Force refresh if first time, OR if bot owner/admin is interacting but cache says unauthorized
			isCommand := strings.HasPrefix(text, "!") || strings.HasPrefix(text, "#")
			shouldRefresh := !cached || (isCommand && !cachedAuth) || (v.Info.IsFromMe && !cachedAuth)

			if shouldRefresh {
				resp, err := client.GetGroupInfo(context.Background(), chatJID)
				if err == nil {
					ownerUser := GetBarePhone(resp.OwnerJID.User)
					tempAuthorized := false

					// Bot is authorized if it's the owner or an admin
					if (botPhone != "" && ownerUser == botPhone) || (botLID != "" && ownerUser == botLID) {
						tempAuthorized = true
					}

					for _, p := range resp.Participants {
						pUser := GetBarePhone(p.JID.User)
						isBot := (botPhone != "" && pUser == botPhone) || (botLID != "" && pUser == botLID)
						if isBot {
							if p.IsAdmin || p.IsSuperAdmin {
								tempAuthorized = true
							}
						}

						// Is this participant the SENDER?
						if GetBarePhone(v.Info.Sender.User) == pUser {
							if p.IsAdmin || p.IsSuperAdmin || pUser == ownerUser {
								isSenderAdmin = true
							}
						}
					}
					isAuthorized = tempAuthorized
					ldb.SetGroupAuthority(context.Background(), chatJID.String(), isAuthorized)
				}
			} else {
				isAuthorized = cachedAuth
				if v.Info.IsFromMe {
					isSenderAdmin = true
				}
			}

			if !isAuthorized && !isSenderAdmin {
				logger.Warn().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot NOT authorized and sender is not admin")
				return
			}
		} else {
			// Private chat rules
			if cfg.AllowPrivateChat {
				isAuthorized = true
			} else {
				hasGroups, _ := ldb.HasAuthorizedGroups(context.Background())
				isAuthorized = !hasGroups
			}
			if senderPhone == botPhone || (botLID != "" && senderPhone == botLID) {
				isSenderAdmin = true
			}
		}

		if v.Info.IsFromMe {
			isSenderAdmin = true // Always trust self
		}

		// Final check - we either need the bot to be authorized in the group,
		// OR the sender to be an admin/owner who can override.
		if !isAuthorized && !isSenderAdmin {
			return
		}

		isAdmin := isSenderAdmin || (senderPhone == botPhone)

		if isAdmin {
			logger.Info().Str("sender", senderPhone).Msg("[RBAC DEBUG] Account identified as ADMIN (Bot Owner or Group Admin)")
		} else {
			logger.Debug().Str("sender", senderPhone).Str("botPhone", botPhone).Msg("[RBAC DEBUG] Account identified as REGULAR USER")
		}

		queue <- models.Job{
			ChatJID:     v.Info.Chat,
			SenderJID:   v.Info.Sender,
			MessageID:   v.Info.ID,
			Text:        strings.TrimSpace(text),
			SenderPhone: senderPhone,
			IsAdmin:     isAdmin,
		}
	}
}
