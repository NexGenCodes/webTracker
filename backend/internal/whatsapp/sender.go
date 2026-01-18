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

func (s *Sender) Reply(chat, sender types.JID, text string, quotedID string, quotedText string) {
	// Add Footer to all messages
	text += BotFooter

	content := &waProto.Message{}

	if quotedID != "" {
		content.ExtendedTextMessage = &waProto.ExtendedTextMessage{
			Text: models.StrPtr(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      models.StrPtr(quotedID),
				Participant:   models.StrPtr(sender.String()),
				QuotedMessage: &waProto.Message{Conversation: models.StrPtr(quotedText)},
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

// Send sends a message without quoting (for follow-up messages)
func (s *Sender) Send(chat types.JID, text string) {
	text += BotFooter
	content := &waProto.Message{
		Conversation: models.StrPtr(text),
	}

	jitter := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(jitter)

	resp, err := s.Client.SendMessage(context.Background(), chat, content)
	if err != nil {
		logger.Error().Err(err).Str("chat", chat.String()).Msg("Failed to send message")
	} else {
		logger.Debug().Str("resp_id", resp.ID).Msg("Message sent")
	}
}

// SendImage sends an image with optional caption as a quoted reply
func (s *Sender) SendImage(chat, sender types.JID, imageBytes []byte, caption string, quotedID string, quotedText string) error {
	// Upload image to WhatsApp servers
	uploaded, err := s.Client.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upload image to WhatsApp")
		return err
	}

	// Add footer to caption
	if caption != "" {
		caption += BotFooter
	}

	// Create image message
	imageMsg := &waProto.ImageMessage{
		URL:           models.StrPtr(uploaded.URL),
		DirectPath:    models.StrPtr(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      models.StrPtr("image/jpeg"),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    models.Uint64Ptr(uint64(len(imageBytes))),
	}

	if caption != "" {
		imageMsg.Caption = models.StrPtr(caption)
	}

	// Add quoted message context if provided
	if quotedID != "" {
		imageMsg.ContextInfo = &waProto.ContextInfo{
			StanzaID:      models.StrPtr(quotedID),
			Participant:   models.StrPtr(sender.String()),
			QuotedMessage: &waProto.Message{Conversation: models.StrPtr(quotedText)},
		}
	}

	content := &waProto.Message{
		ImageMessage: imageMsg,
	}

	// Anti-spam jitter
	jitter := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(jitter)

	resp, err := s.Client.SendMessage(context.Background(), chat, content)
	if err != nil {
		logger.Error().Err(err).Str("chat", chat.String()).Msg("Failed to send image")
		return err
	}

	logger.Info().Str("resp_id", resp.ID).Int("size_kb", len(imageBytes)/1024).Msg("Image sent successfully")
	return nil
}
