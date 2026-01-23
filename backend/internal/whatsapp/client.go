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
		isPrivate := !isGroup

		senderPhone := GetBarePhone(v.Info.Sender.User)
		botReference := "" // Alternative JID (LID) if available

		if v.Info.IsFromMe {
			if client.Store.ID != nil {
				senderPhone = GetBarePhone(client.Store.ID.User)
				botReference = GetBarePhone(v.Info.Sender.User)
				logger.Info().Str("senderJID", v.Info.Sender.String()).Str("botPhone", senderPhone).Str("botLID", botReference).Msg("[RBAC DIAGNOSTIC] Attributed sender to bot account")
			}
		}

		allowed := false

		if isGroup {
			// Rule: Log-only for now as requested.
			isAuthorized, cached, _ := ldb.GetGroupAuthority(context.Background(), chatJID.String())

			// Self-healing: If we are the sender (IsFromMe) but cache says we aren't authorized, force a refresh.
			shouldRefresh := !cached || (v.Info.IsFromMe && !isAuthorized)

			if shouldRefresh {
				if cached {
					logger.Info().Str("group", chatJID.String()).Msg("[RBAC DIAGNOSTIC] Forcing refresh: Bot sent message but cache says NOT authorized")
				}

				resp, err := client.GetGroupInfo(context.Background(), chatJID)
				if err == nil {
					botPhone := GetBarePhone(client.Store.ID.User)
					ownerUser := GetBarePhone(resp.OwnerJID.User)

					tempAuthorized := false
					for _, participant := range resp.Participants {
						pUser := GetBarePhone(participant.JID.User)
						isMatch := pUser == botPhone || (botReference != "" && pUser == botReference)
						if isMatch {
							if participant.IsAdmin || participant.IsSuperAdmin || ownerUser == botPhone || (botReference != "" && ownerUser == botReference) {
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
			allowed = true // Log only for groups
		} else if isPrivate {
			if cfg.AllowPrivateChat {
				allowed = true
			} else {
				hasGroups, _ := ldb.HasAuthorizedGroups(context.Background())
				if !hasGroups {
					allowed = true // Fallback
					logger.Debug().Msg("[RBAC DEBUG] Private chat allowed (Fallback: Bot is not Admin in any group)")
				} else {
					allowed = false
					logger.Debug().Msg("[RBAC DEBUG] Private chat blocked: Bot is Admin in some groups")
				}
			}
		}

		if !allowed {
			return
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
