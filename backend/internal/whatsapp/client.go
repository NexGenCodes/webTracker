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

		// Identify Chat Context
		chatJID := v.Info.Chat
		isGroup := chatJID.Server == "g.us"
		isPrivate := !isGroup

		allowed := false

		if isGroup {
			// Rule: Log-only for now as requested.
			// Check authority (Admin Status) but don't block.
			isAuthorized, cached, _ := ldb.GetGroupAuthority(context.Background(), chatJID.String())
			if !cached {
				resp, err := client.GetGroupInfo(context.Background(), chatJID)
				if err == nil {
					botUser := client.Store.ID.User
					ownerUser := resp.OwnerJID.User
					tempAuthorized := false
					for _, participant := range resp.Participants {
						if participant.JID.User == botUser {
							if participant.IsAdmin || participant.IsSuperAdmin || ownerUser == botUser {
								tempAuthorized = true
							}
							break
						}
					}
					isAuthorized = tempAuthorized
					ldb.SetGroupAuthority(context.Background(), chatJID.String(), isAuthorized)
				}
			}

			if isAuthorized {
				logger.Info().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot is Admin/Owner in this group")
			} else {
				logger.Warn().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot is NOT Admin in this group (Not blocking - Log Only)")
			}
			allowed = true // Log only: always allow group processing for now
		} else if isPrivate {
			// Rule:
			// 1. If AllowPrivateChat is true -> Always allowed
			// 2. If AllowPrivateChat is false -> Allowed ONLY if bot is NOT admin in ANY group
			if cfg.AllowPrivateChat {
				allowed = true
			} else {
				hasGroups, _ := ldb.HasAuthorizedGroups(context.Background())
				if !hasGroups {
					allowed = true // Fallback
					logger.Debug().Msg("[RBAC DEBUG] Private chat allowed (Fallback: Bot is not Admin in any group)")
				} else {
					allowed = false
					logger.Debug().Msg("[RBAC DEBUG] Private chat blocked: Bot is Admin in some groups and AllowPrivateChat is false")
				}
			}
		}

		if !allowed {
			return
		}

		// Queue job (Language fetch moved to worker to avoid blocking event listener)
		senderPhone := GetBarePhone(v.Info.Sender.User)
		if v.Info.IsFromMe {
			// If it's from me (linked device or main), ensure we use the bot's pairing phone
			if client.Store.ID != nil {
				senderPhone = GetBarePhone(client.Store.ID.User)
			}
		}

		// RBAC DEBUG LOGGING (Verification Phase)
		isAdmin := false
		for _, admin := range cfg.AdminPhones {
			if senderPhone == admin {
				isAdmin = true
				break
			}
		}
		if isAdmin {
			logger.Info().Str("sender", senderPhone).Msg("[RBAC DEBUG] Account identified as ADMIN")
		} else {
			logger.Debug().Str("sender", senderPhone).Interface("adminList", cfg.AdminPhones).Msg("[RBAC DEBUG] Account identified as REGULAR USER")
		}

		queue <- models.Job{
			ChatJID:     v.Info.Chat,
			SenderJID:   v.Info.Sender,
			MessageID:   v.Info.ID,
			Text:        strings.TrimSpace(text),
			SenderPhone: senderPhone,
		}
	}
}
