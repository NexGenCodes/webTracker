package whatsapp

import (
	"context"
	"strings"

	"webtracker-bot/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func NewClient(dbURL string) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "pgx", dbURL, dbLog)
	if err != nil {
		return nil, err
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, err
	}

	client := whatsmeow.NewClient(device, waLog.Stdout("whatsapp", "INFO", true))
	return client, nil
}

func HandleEvent(evt interface{}, queue chan<- models.Job, allowedGroups []string) {
	switch v := evt.(type) {
	case *events.Message:
		text := ""
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage().GetText() != "" {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		if text == "" {
			return
		}

		// Filter by Group Allowlist
		if len(allowedGroups) > 0 {
			isAllowed := false
			chatJID := v.Info.Chat.String()
			for _, g := range allowedGroups {
				if chatJID == g {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				return
			}
		}

		// Strictly check !INFO Prefix (Case Insensitive)
		upperText := strings.ToUpper(text)
		if !strings.HasPrefix(upperText, "!INFO") {
			return
		}

		// Clean up prefix (always 5 chars: "!INFO")
		cleanText := strings.TrimSpace(text[5:])
		if cleanText == "" {
			return
		}

		// Queue job
		queue <- models.Job{
			ChatJID:     v.Info.Chat,
			MessageID:   v.Info.ID,
			Text:        cleanText,
			SenderPhone: v.Info.Sender.User,
		}
	}
}
