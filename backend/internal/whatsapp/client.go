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

		// Access Control Logic
		allowed := false

		// Rule: "If the message is a private chat, allow it only if AllowPrivateChat is true."
		if isPrivate {
			if cfg.AllowPrivateChat {
				allowed = true
			} else if len(cfg.AllowedGroups) == 0 {
				// User Rule: "Exception: If AllowedGroups is empty AND AllowPrivateChat is false, the bot SHOULD work in private chats anyway."
				allowed = true
			} else {
				allowed = false // Specifically blocked because AllowedGroups has entries
			}
		}

		// Rule: "Group Chats: The bot must only respond if in AllowedGroups list"
		if isGroup {
			if len(cfg.AllowedGroups) == 0 {
				allowed = false
			} else {
				// Check if group is in the allowed list
				for _, g := range cfg.AllowedGroups {
					if chatJID.String() == g {
						allowed = true
						break
					}
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

		// Group Authority Check (Log-Only for now)
		if isGroup {
			isAuthorized, cached, _ := ldb.GetGroupAuthority(context.Background(), chatJID.String())
			if !cached {
				// Fetch from WhatsApp
				resp, err := client.GetGroupInfo(context.Background(), chatJID)
				if err == nil {
					// Check if bot is Admin or SuperAdmin or Owner
					botUser := client.Store.ID.User
					ownerUser := resp.OwnerJID.User

					for _, participant := range resp.Participants {
						if participant.JID.User == botUser {
							if participant.IsAdmin || participant.IsSuperAdmin || ownerUser == botUser {
								isAuthorized = true
							}
							break
						}
					}
					ldb.SetGroupAuthority(context.Background(), chatJID.String(), isAuthorized)
				}
			}

			if isAuthorized {
				logger.Info().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot authorized in group (Admin/Owner)")
			} else {
				logger.Warn().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot NOT authorized in group (Regular Member)")
				// return // Not blocking yet, just logging
			}
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
