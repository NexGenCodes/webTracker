package whatsapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/localdb"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
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

var (
	authCache         sync.Map // Map[string]bool (GroupJID -> isAuthorized)
	participantsCache sync.Map // Map[string]map[string]bool (GroupJID -> BarePhone -> isAdmin)
	identityCache     struct {
		sync.RWMutex
		botPhone string
		botLID   string
	}
)

func HandleEvent(client *whatsmeow.Client, evt interface{}, queue chan<- models.Job, cfg *config.Config, ldb *localdb.Client) {
	switch v := evt.(type) {
	case *events.JoinedGroup:
		logger.Info().Str("chat", v.JID.String()).Msg("[RBAC EVENT] Joined group, re-verifying authority")
		verifyGroupAuthority(client, ldb, v.JID)

	case *events.GroupInfo:
		logger.Info().Str("chat", v.JID.String()).Msg("[RBAC EVENT] Group info updated, re-verifying authority")
		verifyGroupAuthority(client, ldb, v.JID)

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

		// 1. Identify Bot Identity (Phone & LID) from cache or store
		identityCache.RLock()
		botPhone := identityCache.botPhone
		botLID := identityCache.botLID
		identityCache.RUnlock()

		if botPhone == "" && client.Store.ID != nil {
			botPhone = GetBarePhone(client.Store.ID.User)
			identityCache.Lock()
			identityCache.botPhone = botPhone
			identityCache.Unlock()
		}
		if botLID == "" {
			botLID, _ = ldb.GetSystemConfig(context.Background(), "bot_lid")
			identityCache.Lock()
			identityCache.botLID = botLID
			identityCache.Unlock()
		}

		// 2. Persistent Bot LID Mapping (Update cache if found)
		if v.Info.IsFromMe && client.Store.ID != nil {
			senderPhone = botPhone
			newLID := GetBarePhone(v.Info.Sender.User)
			if newLID != "" && newLID != botLID {
				botLID = newLID
				identityCache.Lock()
				identityCache.botLID = botLID
				identityCache.Unlock()
				_ = ldb.SetSystemConfig(context.Background(), "bot_lid", botLID)
			}
		}

		isAuthorized := false  // Bot's permission in group
		isSenderAdmin := false // Sender's permission to control bot

		if isGroup {
			// FAST CHECK: Use in-memory Go Map (sync.Map)
			if val, ok := authCache.Load(chatJID.String()); ok {
				isAuthorized = val.(bool)
			} else {
				// Not in memory? Re-verify and populate cache
				isAuthorized = verifyGroupAuthority(client, ldb, chatJID)
			}

			// DIAGNOSTIC LOG (As requested: non-blocking, just log)
			logger.Info().
				Str("group", chatJID.String()).
				Bool("is_authorized", isAuthorized).
				Msg("[RBAC DIAGNOSTIC] Current authorization status")

			// If Bot IS authorized, check Sender permissions
			if isAuthorized {
				if v.Info.IsFromMe {
					isSenderAdmin = true
				} else if strings.HasPrefix(text, "!") || strings.HasPrefix(text, "#") {
					// Check Cache First
					senderBare := GetBarePhone(v.Info.Sender.User)
					if groupAdmins, ok := participantsCache.Load(chatJID.String()); ok {
						admins := groupAdmins.(map[string]bool)
						if isAdminEntry, exist := admins[senderBare]; exist {
							isSenderAdmin = isAdminEntry
						} else {
							// If not in cache, re-verify (group might have changed)
							verifyGroupAuthority(client, ldb, chatJID)
							if groupAdminsNew, okNew := participantsCache.Load(chatJID.String()); okNew {
								isSenderAdmin = groupAdminsNew.(map[string]bool)[senderBare]
							}
						}
					} else {
						// No cache? Re-verify
						verifyGroupAuthority(client, ldb, chatJID)
						if groupAdminsNew, okNew := participantsCache.Load(chatJID.String()); okNew {
							isSenderAdmin = groupAdminsNew.(map[string]bool)[senderBare]
						}
					}
				}
			}

			// STRICT RULE: If bot is not Admin/Owner, it completely ignores the group.
			if !isAuthorized {
				logger.Debug().Str("group", chatJID.String()).Msg("[RBAC DEBUG] Bot ignored group: Not an Admin/Owner")
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

		// In Diagnostic mode, we only block if it's strictly required for command authorization
		// But for now, we'll let everything through to the dispatcher as requested
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

// verifyGroupAuthority performs a real-time check. Updates both DB and in-memory cache.
func verifyGroupAuthority(client *whatsmeow.Client, ldb *localdb.Client, chat types.JID) bool {
	resp, err := client.GetGroupInfo(context.Background(), chat)
	if err != nil {
		logger.Error().Err(err).Str("chat", chat.String()).Msg("[RBAC EVENT] Failed to fetch group info")
		return false
	}

	identityCache.RLock()
	botPhone := identityCache.botPhone
	botLID := identityCache.botLID
	identityCache.RUnlock()

	if botPhone == "" && client.Store.ID != nil {
		botPhone = GetBarePhone(client.Store.ID.User)
	}

	ownerUserJID := GetBarePhone(resp.OwnerJID.User)
	isAuth := (botPhone != "" && ownerUserJID == botPhone) || (botLID != "" && ownerUserJID == botLID)

	admins := make(map[string]bool)
	for _, p := range resp.Participants {
		pBare := GetBarePhone(p.JID.User)
		isAdmin := p.IsAdmin || p.IsSuperAdmin || pBare == ownerUserJID
		admins[pBare] = isAdmin

		if !isAuth && (pBare == botPhone || (botLID != "" && pBare == botLID)) {
			if isAdmin {
				isAuth = true
			}
		}
	}

	// Update Memory AND Database
	authCache.Store(chat.String(), isAuth)
	participantsCache.Store(chat.String(), admins)
	ldb.SetGroupAuthority(context.Background(), chat.String(), isAuth)

	logger.Info().
		Str("group", chat.String()).
		Bool("is_authorized", isAuth).
		Int("cached_members", len(admins)).
		Msg("[RBAC EVENT] Authority status synchronized (Memory & DB)")

	return isAuth
}
