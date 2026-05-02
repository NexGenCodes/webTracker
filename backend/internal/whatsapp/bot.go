package whatsapp

import (
	"context"
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

type BotProvider interface {
	GetBot(companyID uuid.UUID) (*BotInstance, error)
	GetAllBots() []*BotInstance
	ActivateBot(ctx context.Context, companyID uuid.UUID) error
	DeactivateBot(companyID uuid.UUID) error
	GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error)
	GetQR(ctx context.Context, companyID uuid.UUID) (string, error)
	LogoutBot(companyID uuid.UUID) error
}
