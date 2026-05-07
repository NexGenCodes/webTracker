package commands

import (
	"context"
	"fmt"
	"strings"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/utils"

	"github.com/google/uuid"
)

func i18nLang(s string) i18n.Language {
	return i18n.Language(strings.ToLower(s))
}

// Result represents the outcome of a command execution.
type Result struct {
	Message  string
	Language string
	EditID   string
	Image    []byte
	Error    error
}

type Handler interface {
	Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result
}

type Dispatcher struct {
	cfg           *config.Config
	shipUC        models.ShipmentUsecase
	configUC      models.ConfigUsecase
	sender        models.WhatsAppSender
	handlers      map[string]Handler
	AwbCmd        string
	CompanyName   string
	Tier          string
	BotPhone      string
	AdminTimezone string
}

func NewDispatcher(cfg *config.Config, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, sender models.WhatsAppSender, awbCmd string, companyName string, botPhone string, adminTimezone string, tier string) *Dispatcher {
	d := &Dispatcher{
		cfg:           cfg,
		shipUC:        shipUC,
		configUC:      configUC,
		sender:        sender,
		handlers:      make(map[string]Handler),
		AwbCmd:        awbCmd,
		CompanyName:   companyName,
		Tier:          tier,
		BotPhone:      botPhone,
		AdminTimezone: adminTimezone,
	}
	d.registerDefaults()
	return d
}

func (d *Dispatcher) registerDefaults() {
	d.handlers["stats"] = &StatsHandler{}
	d.handlers["info"] = &InfoHandler{}
	d.handlers["help"] = &HelpHandler{}
	d.handlers["lang"] = &LangHandler{}
	d.handlers["edit"] = &EditHandler{}
	d.handlers["delete"] = &DeleteHandler{}
	d.handlers["status"] = &StatusHandler{}
	d.handlers["receipt"] = &ReceiptHandler{}
}

func (d *Dispatcher) Dispatch(ctx context.Context, companyID uuid.UUID, text string) (*Result, bool) {
	if !presentsAsCommand(text) {
		return nil, false
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, false
	}
	rawCmd := strings.ToLower(parts[0][1:]) // Remove "!" prefix
	args := parts[1:]

	if handler, ok := d.handlers[rawCmd]; ok {
		jid := utils.GetJID(ctx)
		senderPhone := utils.GetSenderPhone(ctx)
		isAdmin := utils.IsAdmin(ctx)

		isOwner := senderPhone == d.BotPhone

		if !isOwner {
			allowed, retryIn := utils.Allow(senderPhone, d.Tier)
			if !allowed {
				return &Result{Message: fmt.Sprintf("⏳ *RATE LIMIT REACHED*\n\n_Please wait %d seconds before sending another command._", int(retryIn.Seconds()))}, true
			}
		}

		switch h := handler.(type) {
		case *StatsHandler:
			h.CompanyName = d.CompanyName
			h.AdminTimezone = d.AdminTimezone
		case *InfoHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *HelpHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
		case *EditHandler:
			h.CompanyName = d.CompanyName
			h.CompanyPrefix = d.AwbCmd
			h.AdminTimezone = d.AdminTimezone
			h.Sender = d.sender
			h.Cfg = d.cfg
		case *StatusHandler:
			h.BotPhone = d.BotPhone
		case *ReceiptHandler:
			h.Sender = d.sender
		}

		lang, _ := d.configUC.GetUserLanguage(ctx, companyID, jid)

		isOwnerOnlyCmd := rawCmd == "status"
		if isOwnerOnlyCmd {
			if !isOwner {
				logger.Warn().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Owner-only command blocked")
				return &Result{Message: i18n.T(i18nLang(lang), "ERR_OWNER_ONLY")}, true
			}
		}

		isPublicCmd := rawCmd == "info" || rawCmd == "help"
		if !isPublicCmd && !isOwnerOnlyCmd {
			if isAdmin {
				logger.Info().Str("cmd", rawCmd).Str("sender", senderPhone).Msg("Admin command authorized")
			} else {
				return &Result{Message: i18n.T(i18nLang(lang), "ERR_ACCESS_DENIED")}, true
			}
		}

		res := handler.Execute(ctx, d.shipUC, d.configUC, companyID, args, lang, isAdmin)
		if res.Language != "" {
			d.configUC.SetUserLanguage(ctx, companyID, jid, res.Language)
		}

		return &res, true
	}

	return nil, false
}

func presentsAsCommand(text string) bool {
	return len(text) > 1 && (text[0] == '!' || text[0] == '#')
}
