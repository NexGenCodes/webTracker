package models

import (
	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/config"
	"context"
	"database/sql"
	"time"
)

type WhatsAppSender interface {
	Reply(chat, sender types.JID, text string, quotedID string, quotedText string)
	Send(chat types.JID, text string)
	SendImage(chat, sender types.JID, imageBytes []byte, caption string, quotedID string, quotedText string) error
	SetTyping(chat types.JID, typing bool)
	GetWAClient() *whatsmeow.Client
	GetCompanyName() string
}

type BotInstance interface {
	GetWAClient() *whatsmeow.Client
	GetSender() WhatsAppSender
	GetPrefix() string
	GetCompanyName() string
	GetTier() string
	GetJobs() chan Job
	GetCurrentQR() string
}

type BotProvider interface {
	GetBot(companyID uuid.UUID) (BotInstance, error)
	GetAllBots() []BotInstance
	ActivateBot(ctx context.Context, companyID uuid.UUID) error
	DeactivateBot(companyID uuid.UUID) error
	GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error)
	GetQR(ctx context.Context, companyID uuid.UUID) (string, error)
	LogoutBot(companyID uuid.UUID) error
	PurgeBot(companyID uuid.UUID) error
	LivenessCheck()
}

type ShipmentUsecase interface {
	Track(ctx context.Context, companyID uuid.UUID, trackingID string) (*db.Shipment, error)
	Create(ctx context.Context, companyID uuid.UUID, params db.CreateShipmentParams) error
	RecordEvent(ctx context.Context, companyID uuid.UUID, eventType string, metadata []byte) error
	GetService() ShipmentService
	CountByStatus(ctx context.Context, companyID uuid.UUID) (*db.CountShipmentsByStatusRow, error)
	GetLastForUser(ctx context.Context, companyID uuid.UUID, jid string) (string, error)
	UpdateField(ctx context.Context, companyID uuid.UUID, trackingID, field, value string) error
	Delete(ctx context.Context, companyID uuid.UUID, trackingID string) error
	CountCreatedSince(ctx context.Context, companyID uuid.UUID, since time.Time) (int64, error)
	CreateWithPrefix(ctx context.Context, companyID uuid.UUID, s *db.Shipment, prefix string) (string, error)
	FindSimilar(ctx context.Context, companyID uuid.UUID, userJid, phone string) (string, error)
	CheckShipmentCap(ctx context.Context, cfg *config.Config, companyID uuid.UUID, adminEmail string, planType string, expiry sql.NullTime) (int64, error)
}

type ShipmentService interface {
	CalculateDeparture(now time.Time, originTZ string) time.Time
	CalculateArrival(departure time.Time, senderCountry, receiverCountry string) (time.Time, time.Time)
	ResolveTimezone(country string) string
}

type ConfigUsecase interface {
	GetCompanyByID(ctx context.Context, id uuid.UUID) (db.Company, error)
	GetAllCompanies(ctx context.Context) ([]uuid.UUID, error)
	GetAllActiveCompanies(ctx context.Context) ([]db.Company, error)
	UpdateCompanyAuthStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateCompanyWhatsAppPhone(ctx context.Context, id uuid.UUID, phone string) error
	GetUserLanguage(ctx context.Context, companyID uuid.UUID, jid string) (string, error)
	SetUserLanguage(ctx context.Context, companyID uuid.UUID, jid, lang string) error
	GetGroupAuthority(ctx context.Context, companyID uuid.UUID, jid string) (bool, bool, error)
	SetGroupAuthority(ctx context.Context, companyID uuid.UUID, jid string, isAuthorized bool) error
	GetSystemConfig(ctx context.Context, companyID uuid.UUID, key string) (string, error)
	SetSystemConfig(ctx context.Context, companyID uuid.UUID, key, value string) error
	HasAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (bool, error)
	GetAuthorizedGroups(ctx context.Context, companyID uuid.UUID) ([]string, error)
	CountAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (int64, error)
	Ping(ctx context.Context) error
	GetActivePlans(ctx context.Context) ([]db.GetActivePlansRow, error)
	DeleteCompany(ctx context.Context, companyID uuid.UUID) error
	GetPlanByID(ctx context.Context, id string) (db.GetPlanByIDRow, error)
	RecordPayment(ctx context.Context, companyID uuid.UUID, reference string, amount float64, status string) (int32, error)
	UpdateCompanySubscriptionStatus(ctx context.Context, companyID uuid.UUID, subStatus, planType string) error
	LogAudit(ctx context.Context, actorEmail, action string, targetCompanyID uuid.UUID, details map[string]interface{}) error
	GetAuditLogs(ctx context.Context, limit, offset int32) ([]db.AuditLog, error)
	GetPlatformAnalytics(ctx context.Context) (db.GetPlatformAnalyticsRow, error)
	UpdateCompanyPlan(ctx context.Context, companyID uuid.UUID, planType string) error
	UpdateCompanySubscription(ctx context.Context, companyID uuid.UUID, subStatus string, expiry time.Time) error
	GetCompanyPayments(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]db.Payment, error)
}

type Manifest struct {
	ReceiverName    string   `json:"receiverName"`
	ReceiverAddress string   `json:"receiverAddress"`
	ReceiverPhone   string   `json:"receiverPhone"`
	ReceiverCountry string   `json:"receiverCountry"`
	ReceiverEmail   string   `json:"receiverEmail"`
	ReceiverID      string   `json:"receiverID"`
	SenderName      string   `json:"senderName"`
	SenderCountry   string   `json:"senderCountry"`
	CargoType       string   `json:"cargoType"`
	Weight          float64  `json:"weight"`
	IsAI            bool     `json:"-"`
	MissingFields   []string `json:"-"`
}

// Merge combines this manifest with another, only filling in empty fields.
func (m *Manifest) Merge(other Manifest) {
	fillIfEmpty(&m.ReceiverName, other.ReceiverName)
	fillIfEmpty(&m.ReceiverAddress, other.ReceiverAddress)
	fillIfEmpty(&m.ReceiverPhone, other.ReceiverPhone)
	fillIfEmpty(&m.ReceiverCountry, other.ReceiverCountry)
	fillIfEmpty(&m.ReceiverEmail, other.ReceiverEmail)
	fillIfEmpty(&m.ReceiverID, other.ReceiverID)
	fillIfEmpty(&m.SenderName, other.SenderName)
	fillIfEmpty(&m.SenderCountry, other.SenderCountry)
	if m.Weight == 0 && other.Weight > 0 {
		m.Weight = other.Weight
	}
}

func fillIfEmpty(target *string, val string) {
	if *target == "" {
		*target = val
	}
}

// Validate checks for required fields and returns a list of missing ones.
func (m *Manifest) Validate() []string {
	var missing []string
	check := func(val, name string) {
		if val == "" {
			missing = append(missing, name)
		}
	}

	check(m.ReceiverName, "Receiver Name")
	check(m.ReceiverPhone, "Receiver Phone")
	check(m.ReceiverAddress, "Receiver Address")
	check(m.ReceiverCountry, "Receiver Country")
	check(m.SenderName, "Sender Name")
	check(m.SenderCountry, "Sender Country")

	m.MissingFields = missing
	return missing
}

type Job struct {
	CompanyID   uuid.UUID
	ChatJID     types.JID
	SenderJID   types.JID
	MessageID   string
	Text        string
	SenderPhone string
	Language    string
	IsAdmin     bool
	RawMessage  *events.Message
}

func StrPtr(s string) *string {
	return &s
}

func Uint64Ptr(u uint64) *uint64 {
	return &u
}
