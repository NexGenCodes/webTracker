package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"webtracker-bot/internal/models"
)

type LangHandler struct{}

func (h *LangHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "🌐 *LANGUAGE MENU*\n\nUsage: `!lang [en|pt|es|de]`\n\nExample: `!lang pt` para Português"}
	}

	newLang := strings.ToLower(args[0])
	switch newLang {
	case "en", "pt", "es", "de":
		// Handled by dispatcher update
		return Result{
			Message:  fmt.Sprintf("🌐 *LANGUAGE UPDATED*\n\nYour language is now set to *%s*.", strings.ToUpper(newLang)),
			Language: newLang,
		}
	default:
		return Result{Message: "❌ *UNSUPPORTED LANGUAGE*\n\nAvailable: `en`, `pt`, `es`, `de`"}
	}
}
