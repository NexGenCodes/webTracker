package whatsapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewStore(dsn string) (*sqlstore.Container, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	if !strings.Contains(dsn, "default_query_exec_mode") {
		if strings.Contains(dsn, "?") {
			dsn += "&default_query_exec_mode=simple_protocol"
		} else {
			dsn += "?default_query_exec_mode=simple_protocol"
		}
	}
	container, err := sqlstore.New(context.Background(), "pgx", dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}
	return container, nil
}

func NewClientForDevice(device *store.Device) *whatsmeow.Client {
	client := whatsmeow.NewClient(device, waLog.Stdout("whatsapp", "INFO", true))
	logger.Info().Msg("WhatsApp client initialized and session loaded")
	return client
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
	cacheLastClear time.Time
)

const maxCacheAge = 1 * time.Hour

func checkCacheCleanup() {
	if time.Since(cacheLastClear) > maxCacheAge {
		logger.Info().Msg("[RBAC] Clearing participant and authority caches (TTL reached)")
		authCache = sync.Map{}
		participantsCache = sync.Map{}
		cacheLastClear = time.Now()
	}
}

func HandleEvent(client *whatsmeow.Client, evt interface{}, queue chan<- models.Job, cfg *config.Config, configUC *config.Usecase, companyID uuid.UUID) {
	switch v := evt.(type) {
	case *events.JoinedGroup:
		logger.Info().Str("chat", v.JID.String()).Msg("[RBAC EVENT] Joined group, re-verifying authority")
		verifyGroupAuthority(client, configUC, companyID, v.JID)

	case *events.GroupInfo:
		logger.Info().Str("chat", v.JID.String()).Msg("[RBAC EVENT] Group info updated, re-verifying authority")
		verifyGroupAuthority(client, configUC, companyID, v.JID)

	case *events.Message:
		checkCacheCleanup()
		text := ""
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage().GetText() != "" {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		docMsg := v.Message.GetDocumentMessage()
		if text == "" && docMsg == nil {
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
			botLID, _ = configUC.GetSystemConfig(context.Background(), companyID, "bot_lid")
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
				_ = configUC.SetSystemConfig(context.Background(), companyID, "bot_lid", botLID)
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
				isAuthorized = verifyGroupAuthority(client, configUC, companyID, chatJID)
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
							verifyGroupAuthority(client, configUC, companyID, chatJID)
							if groupAdminsNew, okNew := participantsCache.Load(chatJID.String()); okNew {
								isSenderAdmin = groupAdminsNew.(map[string]bool)[senderBare]
							}
						}
					} else {
						// No cache? Re-verify
						verifyGroupAuthority(client, configUC, companyID, chatJID)
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
			isSelfChat := chatJID.User == botPhone || (botLID != "" && chatJID.User == botLID)

			if isSelfChat {
				isAuthorized = true // Always allow Note to Self
			} else if cfg.AllowPrivateChat {
				isAuthorized = true
			} else {
				hasGroups, _ := configUC.HasAuthorizedGroups(context.Background(), companyID)
				isAuthorized = !hasGroups // Failover: Allow private if no groups exist
			}

			if !isAuthorized {
				logger.Debug().
					Str("chat", chatJID.String()).
					Str("sender", senderPhone).
					Msg("[RBAC DEBUG] Ignoring unauthorized private chat")
				return
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
			CompanyID:   companyID,
			ChatJID:     v.Info.Chat,
			SenderJID:   v.Info.Sender,
			MessageID:   v.Info.ID,
			Text:        strings.TrimSpace(text),
			SenderPhone: senderPhone,
			IsAdmin:     isAdmin,
			RawMessage:  v,
		}
	}
}

// verifyGroupAuthority performs a real-time check. Updates both DB and in-memory cache.
func verifyGroupAuthority(client *whatsmeow.Client, configUC *config.Usecase, companyID uuid.UUID, chat types.JID) bool {
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
	configUC.SetGroupAuthority(context.Background(), companyID, chat.String(), isAuth)

	logger.Info().
		Str("group", chat.String()).
		Bool("is_authorized", isAuth).
		Int("cached_members", len(admins)).
		Msg("[RBAC EVENT] Authority status synchronized (Memory & DB)")

	return isAuth
}

