package whatsapp

import (
	"context"
	"math/rand"
	"time"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type Sender struct {
	Client *whatsmeow.Client
}

func NewSender(client *whatsmeow.Client) *Sender {
	return &Sender{Client: client}
}

const BotFooter = "\n\n_🤖Bot_"

func (s *Sender) Reply(chat, sender types.JID, text string, quotedID string) {
	// Add Footer to all messages
	text += BotFooter

	content := &waProto.Message{}

	if quotedID != "" {
		content.ExtendedTextMessage = &waProto.ExtendedTextMessage{
			Text: models.StrPtr(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      models.StrPtr(quotedID),
				Participant:   models.StrPtr(sender.String()),
				QuotedMessage: &waProto.Message{Conversation: models.StrPtr("Original Message")},
			},
		}
	} else {
		content.Conversation = models.StrPtr(text)
	}

	// Anti-Spam Jitter: 500ms to 2500ms
	jitter := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(jitter)

	resp, err := s.Client.SendMessage(context.Background(), chat, content)
	if err != nil {
		logger.Error().Err(err).Str("chat", chat.String()).Msg("Failed to send WhatsApp message")
	} else {
		logger.Debug().Str("resp_id", resp.ID).Msg("Message sent successfully")
	}
}
