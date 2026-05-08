package whatsapp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/receipt"
	"webtracker-bot/internal/utils"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// Manager orchestrates multiple BotInstances across different companies.
type Manager struct {
	Cfg         *config.Config
	ShipmentUC  models.ShipmentUsecase
	ConfigUC    models.ConfigUsecase
	WAStore     *sqlstore.Container
	Bots        map[uuid.UUID]*BotInstance
	BotsMu      sync.RWMutex
	PairLocks   map[uuid.UUID]*sync.Mutex
	PairMu      sync.Mutex
	WG          *sync.WaitGroup
	Context     context.Context
	AsynqClient *asynq.Client
}

// NewManager creates a new multi-tenant WhatsApp manager.
func NewManager(ctx context.Context, cfg *config.Config, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, store *sqlstore.Container, wg *sync.WaitGroup, asynqClient *asynq.Client) *Manager {
	return &Manager{
		Cfg:         cfg,
		ShipmentUC:  shipUC,
		ConfigUC:    configUC,
		WAStore:     store,
		Bots:        make(map[uuid.UUID]*BotInstance),
		PairLocks:   make(map[uuid.UUID]*sync.Mutex),
		WG:          wg,
		Context:     ctx,
		AsynqClient: asynqClient,
	}
}

func (m *Manager) getPairLock(companyID uuid.UUID) *sync.Mutex {
	m.PairMu.Lock()
	defer m.PairMu.Unlock()
	mu, ok := m.PairLocks[companyID]
	if !ok {
		mu = &sync.Mutex{}
		m.PairLocks[companyID] = mu
	}
	return mu
}

// GetBot retrieves a bot instance for a specific company.
func (m *Manager) GetBot(companyID uuid.UUID) (models.BotInstance, error) {
	m.BotsMu.RLock()
	defer m.BotsMu.RUnlock()
	bot, ok := m.Bots[companyID]
	if !ok {
		return nil, fmt.Errorf("bot not found for company %s", companyID)
	}
	return bot, nil
}

// GetAllBots returns all currently active bot instances.
func (m *Manager) GetAllBots() []models.BotInstance {
	m.BotsMu.RLock()
	defer m.BotsMu.RUnlock()
	var bots []models.BotInstance
	for _, b := range m.Bots {
		bots = append(bots, b)
	}
	return bots
}

// ActivateBot initializes and starts a bot for a company.
func (m *Manager) ActivateBot(ctx context.Context, companyID uuid.UUID) error {
	company, err := m.ConfigUC.GetCompanyByID(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	if !company.AuthStatus.Valid || company.AuthStatus.String != "active" {
		return fmt.Errorf("company is not active")
	}

	mu := m.getPairLock(companyID)
	mu.Lock()
	defer mu.Unlock()

	m.BotsMu.RLock()
	if _, exists := m.Bots[companyID]; exists {
		m.BotsMu.RUnlock()
		return fmt.Errorf("bot already active for company")
	}
	m.BotsMu.RUnlock()

	return m.InitBotForCompany(company)
}

// DeactivateBot stops and removes a bot from memory.
func (m *Manager) DeactivateBot(companyID uuid.UUID) error {
	m.BotsMu.Lock()
	bot, exists := m.Bots[companyID]
	if !exists {
		m.BotsMu.Unlock()
		return fmt.Errorf("bot not found")
	}

	bot.GetWAClient().Disconnect()

	delete(m.Bots, companyID)
	m.BotsMu.Unlock()

	// Clean up pairing lock to prevent memory leaks over time
	m.PairMu.Lock()
	delete(m.PairLocks, companyID)
	m.PairMu.Unlock()

	logger.Info().Str("company_id", companyID.String()).Msg("Bot dynamically deactivated")
	return nil
}

// LogoutBot unpairs the bot device and marks it as pending.
func (m *Manager) LogoutBot(companyID uuid.UUID) error {
	// 1. Update Database immediately to "pending" to provide fast UI feedback
	err := m.ConfigUC.UpdateCompanyAuthStatus(m.Context, companyID, "pending")
	if err != nil {
		logger.Error().Err(err).Str("company_id", companyID.String()).Msg("Failed to update auth status to pending")
		return fmt.Errorf("failed to update auth status: %w", err)
	}

	err = m.ConfigUC.UpdateCompanyWhatsAppPhone(m.Context, companyID, "")
	if err != nil {
		logger.Error().Err(err).Str("company_id", companyID.String()).Msg("Failed to clear WhatsApp phone")
	}

	// 2. Try to find the bot in memory for a clean remote logout
	bot, err := m.GetBot(companyID)
	if err != nil {
		// If bot is not in memory, we still succeeded in resetting the DB state
		return nil
	}

	client := bot.GetWAClient()
	if !client.IsConnected() {
		// Attempt a quick reconnection to send the logout signal to unpair on the phone side
		_ = client.Connect()
		// Wait for connection (up to 3 seconds)
		for i := 0; i < 6; i++ {
			if client.IsConnected() && client.Store.ID != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	// 3. Try to logout (remote unpair)
	if client.IsConnected() && client.Store.ID != nil {
		err = client.Logout(m.Context)
		if err != nil {
			logger.Warn().Err(err).Str("company_id", companyID.String()).Msg("Failed to send remote logout signal")
		}
	}

	// 4. Forcefully delete local store session regardless of remote logout success
	if client.Store != nil {
		_ = client.Store.Delete(m.Context)
	}

	return m.DeactivateBot(companyID)
}

// PurgeBot forcefully deletes a bot's session from the database.
func (m *Manager) PurgeBot(companyID uuid.UUID) error {
	// 1. Try a clean logout if the bot is active
	_ = m.LogoutBot(companyID)

	// 2. Get company data to find the phone number
	company, err := m.ConfigUC.GetCompanyByID(m.Context, companyID)
	if err != nil {
		return err
	}

	// 3. If a phone exists, forcefully delete its device from the SQL store
	if company.WhatsappPhone.Valid && company.WhatsappPhone.String != "" {
		jid := types.NewJID(company.WhatsappPhone.String, "s.whatsapp.net")
		device, err := m.WAStore.GetDevice(context.Background(), jid)
		if err == nil && device != nil {
			err = device.Delete(m.Context)
			if err != nil {
				logger.Error().Err(err).Str("phone", company.WhatsappPhone.String).Msg("Failed to purge WhatsApp device from store")
			} else {
				logger.Info().Str("phone", company.WhatsappPhone.String).Msg("WhatsApp device purged from store")
			}
		}
	}

	return nil
}

// InitBotForCompany initializes a new bot instance based on company config.
func (m *Manager) InitBotForCompany(c db.Company) error {
	var device *store.Device
	var err error

	phone := ""
	if c.WhatsappPhone.Valid {
		phone = c.WhatsappPhone.String
	}

	if phone != "" {
		jid := types.NewJID(phone, "s.whatsapp.net")
		device, err = m.WAStore.GetDevice(m.Context, jid)
		if err != nil || device == nil {
			device = m.WAStore.NewDevice()
		}
	} else {
		device = m.WAStore.NewDevice()
	}

	companyName := strings.ToUpper(c.Name.String)
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	waClient := NewClientForDevice(device, companyName)
	prefix := "AWB"
	if c.TrackingPrefix.Valid && c.TrackingPrefix.String != "" {
		prefix = c.TrackingPrefix.String
	} else {
		prefix = utils.GenerateAbbreviation(c.Name.String)
	}

	sender := NewSender(waClient, c.Name.String)
	receipt.InitProcessor()

	bot := &BotInstance{
		CompanyID:   c.ID,
		CompanyName: c.Name.String,
		Prefix:      prefix,
		Tier:        c.SubscriptionStatus.String,
		WA:          waClient,
		Sender:      sender,
		AsynqClient: m.AsynqClient,
	}

	waClient.AddEventHandler(func(evt interface{}) {
		m.HandleWAEvent(bot, evt)
	})

	if waClient.Store.ID == nil {
		qrChan, _ := waClient.GetQRChannel(m.Context)
		go func() {
			for evt := range qrChan {
				bot.QRMu.Lock()
				if evt.Event == "code" {
					bot.CurrentQR = evt.Code
				} else {
					bot.CurrentQR = ""
				}
				bot.QRMu.Unlock()
			}
		}()
	}

	m.BotsMu.Lock()
	m.Bots[c.ID] = bot
	m.BotsMu.Unlock()

	if waClient.Store.ID != nil {
		if !waClient.IsConnected() {
			if err := waClient.Connect(); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
		}
	}

	return nil
}

// HandleWAEvent proxies events to the specific bot instance.
func (m *Manager) HandleWAEvent(bot *BotInstance, evt interface{}) {
	HandleEvent(bot, evt, m.Cfg, m.ConfigUC)

	switch evt.(type) {
	case *events.Connected:
		_ = m.ConfigUC.UpdateCompanyAuthStatus(m.Context, bot.CompanyID, "active")
		if bot.GetWAClient().Store != nil && bot.GetWAClient().Store.ID != nil {
			phone := utils.GetBarePhone(bot.GetWAClient().Store.ID.User)
			if phone != "" {
				_ = m.ConfigUC.UpdateCompanyWhatsAppPhone(m.Context, bot.CompanyID, phone)
			}
		}
		bot.ReconnectCount = 0
		bot.LastReconnect = time.Time{}

	case *events.PairSuccess:
		_ = m.ConfigUC.UpdateCompanyAuthStatus(m.Context, bot.CompanyID, "active")
		if bot.GetWAClient().Store != nil && bot.GetWAClient().Store.ID != nil {
			phone := utils.GetBarePhone(bot.GetWAClient().Store.ID.User)
			if phone != "" {
				_ = m.ConfigUC.UpdateCompanyWhatsAppPhone(m.Context, bot.CompanyID, phone)
			}

			// Priority 1: Send a Welcome Validation Message directly to the paired device
			ownJID := bot.GetWAClient().Store.ID.ToNonAD()
			welcomeMsg := "✅ *CargoHive Bot Activated*\n\nYour WhatsApp tracking bot is now securely linked and fully operational.\n\nAll tracking requests sent to this number will now be automatically processed. 🚀"
			bot.Sender.Send(ownJID, welcomeMsg)
		}
		bot.ReconnectCount = 0
		bot.LastReconnect = time.Time{}

	case *events.LoggedOut:
		// 1. Reset auth status
		_ = m.ConfigUC.UpdateCompanyAuthStatus(m.Context, bot.CompanyID, "pending")
		// 2. Clear the WhatsApp phone in the database
		_ = m.ConfigUC.UpdateCompanyWhatsAppPhone(m.Context, bot.CompanyID, "")
		
		// 3. Forcefully delete local store session cryptographic keys
		if bot.GetWAClient().Store != nil {
			_ = bot.GetWAClient().Store.Delete(m.Context)
		}
		
		// 4. Deactivate the bot from memory
		_ = m.DeactivateBot(bot.CompanyID)
	case *events.Disconnected:
		bot.ReconnectCount++

		// Update DB to reflect reconnecting state (only on first attempt to avoid write spam)
		if bot.ReconnectCount == 1 {
			_ = m.ConfigUC.UpdateCompanyAuthStatus(m.Context, bot.CompanyID, "reconnecting")
		}

		logger.Info().
			Str("company_id", bot.CompanyID.String()).
			Int("attempt", bot.ReconnectCount).
			Msg("Bot disconnected. whatsmeow internal auto-reconnect will handle this.")
	}
}

// getOrInitBot safely retrieves an existing bot or initializes a new one.
func (m *Manager) getOrInitBot(companyID uuid.UUID) (models.BotInstance, error) {
	mu := m.getPairLock(companyID)
	mu.Lock()
	defer mu.Unlock()

	bot, err := m.GetBot(companyID)
	if err == nil {
		return bot, nil
	}

	company, err := m.ConfigUC.GetCompanyByID(m.Context, companyID)
	if err != nil {
		return nil, err
	}
	if err := m.InitBotForCompany(company); err != nil {
		m.PairMu.Lock()
		delete(m.PairLocks, companyID)
		m.PairMu.Unlock()
		return nil, err
	}

	bot, err = m.GetBot(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bot after init: %w", err)
	}
	return bot, nil
}

// GeneratePairingCode generates a pairing code for a companion device.
func (m *Manager) GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error) {
	bot, err := m.getOrInitBot(companyID)
	if err != nil {
		return "", err
	}

	waClient := bot.GetWAClient()
	if waClient.Store.ID != nil {
		return "", fmt.Errorf("device is already paired")
	}

	// Do not forcefully disconnect if already connected.
	if !waClient.IsConnected() {
		if err := waClient.Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for pairing: %w", err)
		}
	}

	// Chain the incoming HTTP ctx with a fallback 60s timeout
	pairCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	displayName := strings.ToUpper(bot.GetCompanyName())
	if displayName == "" {
		displayName = "AIRWAYBILL"
	}

	return waClient.PairPhone(pairCtx, phone, true, whatsmeow.PairClientChrome, fmt.Sprintf("%s (Windows)", displayName))
}

// GetQR retrieves the current pairing QR code for the bot.
func (m *Manager) GetQR(ctx context.Context, companyID uuid.UUID) (string, error) {
	bot, err := m.getOrInitBot(companyID)
	if err != nil {
		return "", err
	}

	waClient := bot.GetWAClient()
	if waClient.Store.ID != nil {
		return "", fmt.Errorf("device is already paired")
	}

	// Ensure connection is active to receive QR events
	if !waClient.IsConnected() {
		if err := waClient.Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for QR: %w", err)
		}
	}

	// Wait for the QR code to be generated by the background goroutine (up to 10 seconds)
	for i := 0; i < 20; i++ {
		code := bot.GetCurrentQR()
		if code != "" {
			return code, nil
		}

		// Actually force a QR channel refresh if we hit the 5-second mark with no QR code.
		// This safely cycles the websocket so WhatsApp emits a fresh QR string.
		if i == 10 {
			logger.Info().Str("company", companyID.String()).Msg("Forcing QR channel refresh...")
			waClient.Disconnect()
			time.Sleep(200 * time.Millisecond)
			if err := waClient.Connect(); err != nil {
				logger.Warn().Err(err).Msg("Failed to reconnect during QR refresh")
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("qr code not available yet, please try again in a moment")
}



// LivenessCheck audits all active bots and corrects auth_status if
// a bot is tracked as "active" in the DB but is actually disconnected.
// Called by the cron scheduler.
func (m *Manager) LivenessCheck() {
	m.BotsMu.RLock()
	var snapshot []struct {
		ID  uuid.UUID
		Bot *BotInstance
	}
	for id, bot := range m.Bots {
		snapshot = append(snapshot, struct {
			ID  uuid.UUID
			Bot *BotInstance
		}{id, bot})
	}
	m.BotsMu.RUnlock()

	go func() {
		// Concurrency limit to prevent Thundering Herd on mass reconnect
		sem := make(chan struct{}, 5) // max 5 concurrent reconnects

		for _, entry := range snapshot {
			client := entry.Bot.GetWAClient()
			if client == nil {
				continue
			}

			if !client.IsConnected() && client.Store != nil && client.Store.ID != nil {
				sem <- struct{}{} // acquire
				
				go func(e struct {
					ID  uuid.UUID
					Bot *BotInstance
				}) {
					defer func() { <-sem }() // release

					// Don't step on other reconnect attempts
					mu := m.getPairLock(e.ID)
					if !mu.TryLock() {
						return
					}
					defer mu.Unlock()

					logger.Warn().Str("company_id", e.ID.String()).Msg("[LivenessCheck] Bot disconnected — attempting reconnect")
					e.Bot.ReconnectCount = 0
					if err := client.Connect(); err != nil {
						logger.Error().Err(err).Str("company_id", e.ID.String()).Msg("[LivenessCheck] Reconnect failed")
						_ = m.ConfigUC.UpdateCompanyAuthStatus(m.Context, e.ID, "disconnected")
					}
					
					// Jitter to prevent API throttling
					time.Sleep(200 * time.Millisecond)
				}(entry)
			}
		}
	}()
}
