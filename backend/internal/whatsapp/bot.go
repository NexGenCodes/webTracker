package whatsapp

import (
	"context"

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
}

type BotProvider interface {
	GetBot(companyID uuid.UUID) (*BotInstance, error)
	GetAllBots() []*BotInstance
	ActivateBot(ctx context.Context, companyID uuid.UUID) error
	DeactivateBot(companyID uuid.UUID) error
	GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error)
}
