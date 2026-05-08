package whatsapp

import (
	"context"
	"sync"
	"time"

	"webtracker-bot/internal/models"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.mau.fi/whatsmeow"
	"golang.org/x/sync/singleflight"
)

// BotInstance represents a single connected WhatsApp bot for a company.
type BotInstance struct {
	CompanyID   uuid.UUID
	CompanyName string
	Prefix      string
	Tier        string
	WA          *whatsmeow.Client
	Sender      *Sender
	CurrentQR   string
	QRMu        sync.RWMutex
	AsynqClient *asynq.Client

	AuthCache         sync.Map // Map[string]bool (GroupJID -> isAuthorized)
	ParticipantsCache sync.Map // Map[string]map[string]bool (GroupJID -> BarePhone -> isAdmin)
	IdentityCache     IdentityCacheData
	CacheLastClear    time.Time
	CacheMu           sync.Mutex
	GroupAuthFlight   singleflight.Group
	ReconnectCount    int
	LastReconnect     time.Time
	KeepaliveCancel   context.CancelFunc // stops the keepalive goroutine
}

// IdentityCacheData holds cached identity information for the bot.
type IdentityCacheData struct {
	sync.RWMutex
	BotPhone string
	BotLID   string
}

// GetWAClient returns the underlying whatsmeow client.
func (b *BotInstance) GetWAClient() *whatsmeow.Client { return b.WA }

// GetSender returns the message sender instance.
func (b *BotInstance) GetSender() models.WhatsAppSender { return b.Sender }

// GetPrefix returns the bot's command prefix.
func (b *BotInstance) GetPrefix() string { return b.Prefix }

// GetCompanyName returns the name of the company.
func (b *BotInstance) GetCompanyName() string { return b.CompanyName }

// GetTier returns the subscription tier.
func (b *BotInstance) GetTier() string { return b.Tier }

// GetCurrentQR returns the current pairing QR code.
func (b *BotInstance) GetCurrentQR() string {
	b.QRMu.RLock()
	defer b.QRMu.RUnlock()
	return b.CurrentQR
}
