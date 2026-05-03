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
	"webtracker-bot/internal/worker"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Manager struct {
	Cfg        *config.Config
	ShipmentUC models.ShipmentUsecase
	ConfigUC   models.ConfigUsecase
	WAStore    *sqlstore.Container
	Bots       map[uuid.UUID]*BotInstance
	BotsMu     sync.RWMutex
	PairLocks  map[uuid.UUID]*sync.Mutex
	PairMu     sync.Mutex
	WG         *sync.WaitGroup
	Context    context.Context
}

func NewManager(ctx context.Context, cfg *config.Config, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, store *sqlstore.Container, wg *sync.WaitGroup) *Manager {
	return &Manager{
		Cfg:        cfg,
		ShipmentUC: shipUC,
		ConfigUC:   configUC,
		WAStore:    store,
		Bots:       make(map[uuid.UUID]*BotInstance),
		PairLocks:  make(map[uuid.UUID]*sync.Mutex),
		WG:         wg,
		Context:    ctx,
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

func (m *Manager) GetBot(companyID uuid.UUID) (models.BotInstance, error) {
	m.BotsMu.RLock()
	defer m.BotsMu.RUnlock()
	bot, ok := m.Bots[companyID]
	if !ok {
		return nil, fmt.Errorf("bot not found for company %s", companyID)
	}
	return bot, nil
}

func (m *Manager) GetAllBots() []models.BotInstance {
	m.BotsMu.RLock()
	defer m.BotsMu.RUnlock()
	var bots []models.BotInstance
	for _, b := range m.Bots {
		bots = append(bots, b)
	}
	return bots
}

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

func (m *Manager) DeactivateBot(companyID uuid.UUID) error {
	m.BotsMu.Lock()
	bot, exists := m.Bots[companyID]
	if !exists {
		m.BotsMu.Unlock()
		return fmt.Errorf("bot not found")
	}

	bot.GetWAClient().Disconnect()
	if bot.Jobs != nil {
		close(bot.Jobs)
	}
	delete(m.Bots, companyID)
	m.BotsMu.Unlock()

	logger.Info().Str("company_id", companyID.String()).Msg("Bot dynamically deactivated")
	return nil
}

func (m *Manager) LogoutBot(companyID uuid.UUID) error {
	// 1. Update Database immediately to "pending" to provide fast UI feedback
	err := m.ConfigUC.UpdateCompanyAuthStatus(context.Background(), companyID, "pending")
	if err != nil {
		logger.Error().Err(err).Str("company_id", companyID.String()).Msg("Failed to update auth status to pending")
		return fmt.Errorf("failed to update auth status: %w", err)
	}
	
	err = m.ConfigUC.UpdateCompanyWhatsAppPhone(context.Background(), companyID, "")
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
		err = client.Logout(context.Background())
		if err != nil {
			logger.Warn().Err(err).Str("company_id", companyID.String()).Msg("Failed to send remote logout signal")
		}
	}

	// 4. Forcefully delete local store session regardless of remote logout success
	if client.Store != nil {
		_ = client.Store.Delete(context.Background())
	}

	return m.DeactivateBot(companyID)
}

func (m *Manager) PurgeBot(companyID uuid.UUID) error {
	// 1. Try a clean logout if the bot is active
	_ = m.LogoutBot(companyID)

	// 2. Get company data to find the phone number
	company, err := m.ConfigUC.GetCompanyByID(context.Background(), companyID)
	if err != nil {
		return err
	}

	// 3. If a phone exists, forcefully delete its device from the SQL store
	if company.WhatsappPhone.Valid && company.WhatsappPhone.String != "" {
		jid := types.NewJID(company.WhatsappPhone.String, "s.whatsapp.net")
		device, err := m.WAStore.GetDevice(context.Background(), jid)
		if err == nil && device != nil {
			err = device.Delete(context.Background())
			if err != nil {
				logger.Error().Err(err).Str("phone", company.WhatsappPhone.String).Msg("Failed to purge WhatsApp device from store")
			} else {
				logger.Info().Str("phone", company.WhatsappPhone.String).Msg("WhatsApp device purged from store")
			}
		}
	}

	return nil
}

func (m *Manager) InitBotForCompany(c db.Company) error {
	var device *store.Device
	var err error

	phone := ""
	if c.WhatsappPhone.Valid {
		phone = c.WhatsappPhone.String
	}

	if phone != "" {
		jid := types.NewJID(phone, "s.whatsapp.net")
		device, err = m.WAStore.GetDevice(context.Background(), jid)
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
		Jobs:        make(chan models.Job, m.Cfg.BufferSize),
	}

	m.WG.Add(1)
	w := &worker.Worker{
		ID:              int(c.ID.ID()),
		Jobs:            bot.Jobs,
		WG:              m.WG,
		Cfg:             m.Cfg,
		ShipmentUC:      m.ShipmentUC,
		ConfigUC:        m.ConfigUC,
		FrontendURL:     m.Cfg.FrontendURL,
		ShipmentService: m.ShipmentUC.GetService(),
		Bots:            m,
		Context:         m.Context,
	}
	go w.Start()

	waClient.AddEventHandler(func(evt interface{}) {
		m.HandleWAEvent(bot, evt)
	})

	if waClient.Store.ID == nil {
		qrChan, _ := waClient.GetQRChannel(context.Background())
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

func (m *Manager) HandleWAEvent(bot *BotInstance, evt interface{}) {
	HandleEvent(bot, evt, bot.Jobs, m.Cfg, m.ConfigUC)

	switch evt.(type) {
	case *events.Connected, *events.PairSuccess:
		_ = m.ConfigUC.UpdateCompanyAuthStatus(context.Background(), bot.CompanyID, "active")
		if bot.GetWAClient().Store != nil && bot.GetWAClient().Store.ID != nil {
			phone := utils.GetBarePhone(bot.GetWAClient().Store.ID.User)
			if phone != "" {
				_ = m.ConfigUC.UpdateCompanyWhatsAppPhone(context.Background(), bot.CompanyID, phone)
			}
		}
		bot.ReconnectCount = 0
		bot.LastReconnect = time.Now()
	case *events.LoggedOut:
		_ = m.ConfigUC.UpdateCompanyAuthStatus(context.Background(), bot.CompanyID, "pending")
		_ = m.DeactivateBot(bot.CompanyID)
	case *events.Disconnected:
		if bot.ReconnectCount >= 5 {
			return
		}
		bot.ReconnectCount++
		delay := time.Duration(bot.ReconnectCount*30) * time.Second
		go func() {
			time.Sleep(delay)
			mu := m.getPairLock(bot.CompanyID)
			if !mu.TryLock() {
				return
			}
			mu.Unlock()

			m.BotsMu.RLock()
			_, stillActive := m.Bots[bot.CompanyID]
			m.BotsMu.RUnlock()
			if stillActive && bot.GetWAClient().Store != nil && bot.GetWAClient().Store.ID != nil {
				_ = bot.GetWAClient().Connect()
			}
		}()
	}
}

func (m *Manager) GeneratePairingCode(ctx context.Context, companyID uuid.UUID, phone string) (string, error) {
	mu := m.getPairLock(companyID)
	mu.Lock()
	bot, err := m.GetBot(companyID)
	if err != nil {
		company, err := m.ConfigUC.GetCompanyByID(context.Background(), companyID)
		if err != nil {
			mu.Unlock()
			return "", err
		}
		if err := m.InitBotForCompany(company); err != nil {
			mu.Unlock()
			return "", err
		}
		bot, _ = m.GetBot(companyID)
	}
	mu.Unlock()

	if bot.GetWAClient().IsConnected() {
		bot.GetWAClient().Disconnect()
		time.Sleep(500 * time.Millisecond)
	}
	if err := bot.GetWAClient().Connect(); err != nil {
		return "", err
	}

	pairCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	displayName := strings.ToUpper(bot.GetCompanyName())
	if displayName == "" {
		displayName = "AIRWAYBILL"
	}

	return bot.GetWAClient().PairPhone(pairCtx, phone, true, whatsmeow.PairClientChrome, fmt.Sprintf("%s (Windows)", displayName))
}

func (m *Manager) GetQR(ctx context.Context, companyID uuid.UUID) (string, error) {
	mu := m.getPairLock(companyID)
	mu.Lock()
	bot, err := m.GetBot(companyID)
	if err != nil {
		company, err := m.ConfigUC.GetCompanyByID(context.Background(), companyID)
		if err != nil {
			mu.Unlock()
			return "", err
		}
		if err := m.InitBotForCompany(company); err != nil {
			mu.Unlock()
			return "", err
		}
		bot, _ = m.GetBot(companyID)
	}
	mu.Unlock()

	// Ensure connection is active to receive QR events
	if !bot.GetWAClient().IsConnected() {
		if err := bot.GetWAClient().Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for QR: %w", err)
		}
	}

	// Wait for the QR code to be generated by the background goroutine (up to 10 seconds)
	for i := 0; i < 20; i++ {
		code := bot.GetCurrentQR()
		if code != "" {
			return code, nil
		}
		
		// If we've waited a bit and still no code, try refreshing the QR channel
		if i == 5 {
			logger.Info().Str("company", companyID.String()).Msg("Retrying QR generation...")
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("QR code not available yet. Please try again in a moment.")
}
