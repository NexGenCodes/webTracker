package whatsapp

import (
	"sync"
	"time"

	"webtracker-bot/internal/models"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
)

type BotInstance struct {
	CompanyID   uuid.UUID
	CompanyName string
	Prefix      string
	Tier        string
	WA          *whatsmeow.Client
	Sender      *Sender
	CurrentQR   string
	QRMu        sync.RWMutex
	Jobs        chan models.Job

	AuthCache         sync.Map // Map[string]bool (GroupJID -> isAuthorized)
	ParticipantsCache sync.Map // Map[string]map[string]bool (GroupJID -> BarePhone -> isAdmin)
	IdentityCache     IdentityCacheData
	CacheLastClear    time.Time
	CacheMu           sync.Mutex
	ReconnectCount    int
	LastReconnect     time.Time
}

type IdentityCacheData struct {
	sync.RWMutex
	BotPhone string
	BotLID   string
}

func (b *BotInstance) GetWAClient() *whatsmeow.Client   { return b.WA }
func (b *BotInstance) GetSender() models.WhatsAppSender { return b.Sender }
func (b *BotInstance) GetPrefix() string                { return b.Prefix }
func (b *BotInstance) GetCompanyName() string           { return b.CompanyName }
func (b *BotInstance) GetTier() string                  { return b.Tier }
func (b *BotInstance) GetJobs() chan models.Job         { return b.Jobs }

func (b *BotInstance) GetCurrentQR() string {
	b.QRMu.RLock()
	defer b.QRMu.RUnlock()
	return b.CurrentQR
}
