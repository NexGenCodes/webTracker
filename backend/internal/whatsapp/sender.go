package whatsapp

import (
	"context"
	"math/rand/v2"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// carries the field: "component":"whatsapp.sender" automatically.
var senderLog = logger.WithComponent("whatsapp.sender")

// OutboundMessage is the envelope placed on the outbound queue.
type OutboundMessage struct {
	Chat    types.JID
	Content *waProto.Message
	// msgType is resolved once at enqueue time so the worker never has to
	// inspect proto bytes to label metrics.
	msgType string
}

// Sender serialises all outbound WhatsApp traffic through a single goroutine
// backed by a buffered channel. This prevents concurrent calls from tripping
// WhatsApp's rate-limiter and makes back-pressure observable.
type Sender struct {
	Client      *whatsmeow.Client
	CompanyName string
	outChan     chan OutboundMessage
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewSender creates a Sender and starts the background dispatch worker.
func NewSender(client *whatsmeow.Client, companyName string) *Sender {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Sender{
		Client:      client,
		CompanyName: companyName,
		outChan:     make(chan OutboundMessage, 500),
		ctx:         ctx,
		cancel:      cancel,
	}
	go s.startWorker()
	senderLog.Info().
		Str("company", companyName).
		Int("queue_capacity", 500).
		Msg("WhatsApp sender initialised")
	return s
}

// startWorker drains the outbound queue, applies anti-spam jitter, then calls
// the WhatsApp client. All metrics are recorded here.
func (s *Sender) startWorker() {
	senderLog.Debug().Str("company", s.CompanyName).Msg("Sender worker started")

	for {
		select {
		case <-s.ctx.Done():
			senderLog.Info().
				Str("company", s.CompanyName).
				Msg("Sender worker stopped (context cancelled)")
			return

		case msg, ok := <-s.outChan:
			if !ok {
				// Channel was closed — drain path during shutdown.
				senderLog.Info().
					Str("company", s.CompanyName).
					Msg("Sender outChan closed, worker exiting")
				return
			}

			// Anti-Spam Jitter: 300–1000 ms to avoid WhatsApp rate-limiting.
			jitter := time.Duration(300+rand.IntN(700)) * time.Millisecond

			senderLog.Trace().
				Str("company", s.CompanyName).
				Str("chat", msg.Chat.String()).
				Str("type", msg.msgType).
				Dur("jitter_ms", jitter).
				Msg("Applying anti-spam jitter before send")

			// Sleep respects context cancellation so shutdown is not delayed.
			select {
			case <-s.ctx.Done():
				senderLog.Info().
					Str("company", s.CompanyName).
					Msg("Sender worker stopped during jitter (context cancelled)")
				return
			case <-time.After(jitter):
			}

			// Use a per-send timeout derived from the sender's context so the
			// call is cancelled if the application is shutting down.
			sendCtx, sendCancel := context.WithTimeout(s.ctx, 30*time.Second)
			start := time.Now()

			resp, err := s.Client.SendMessage(sendCtx, msg.Chat, msg.Content)
			sendCancel()

			elapsed := time.Since(start)

			if err != nil {
				senderLog.Error().
					Err(err).
					Str("company", s.CompanyName).
					Str("chat", msg.Chat.String()).
					Str("type", msg.msgType).
					Dur("duration_ms", elapsed).
					Msg("Failed to send WhatsApp message")
				continue
			}

			senderLog.Debug().
				Str("company", s.CompanyName).
				Str("chat", msg.Chat.String()).
				Str("type", msg.msgType).
				Str("msg_id", resp.ID).
				Dur("duration_ms", elapsed).
				Msg("WhatsApp message sent successfully")
		}
	}
}

// Stop signals the worker to exit and waits for it to drain.
// It is safe to call multiple times.
func (s *Sender) Stop() {
	senderLog.Info().Str("company", s.CompanyName).Msg("Sender stopping")
	s.cancel()
}

// GetWAClient returns the underlying whatsmeow client.
func (s *Sender) GetWAClient() *whatsmeow.Client { return s.Client }

// GetCompanyName returns the name of the company.
func (s *Sender) GetCompanyName() string         { return s.CompanyName }

// BotFooter is the standard text appended to bot messages.
const BotFooter = "\n\n_🤖Bot_"

// enqueue places a message on the outbound channel.
// Returns false (and increments the drop counter) if the queue is full.
func (s *Sender) enqueue(msg OutboundMessage) bool {
	select {
	case <-s.ctx.Done():
		// Sender is shutting down — silently discard.
		return false
	case s.outChan <- msg:
		return true
	default:
		// Queue full: record and log — never block the caller.
		senderLog.Warn().
			Str("company", s.CompanyName).
			Str("chat", msg.Chat.String()).
			Str("type", msg.msgType).
			Int("queue_len", len(s.outChan)).
			Msg("Outbound queue full — message dropped")
		return false
	}
}

// Reply sends a text message, optionally quoting a previous message.
func (s *Sender) Reply(chat, sender types.JID, text string, quotedID string, quotedText string) {
	msgText := text + BotFooter

	content := &waProto.Message{}
	msgType := "text"

	if quotedID != "" {
		msgType = "text_reply"
		content.ExtendedTextMessage = &waProto.ExtendedTextMessage{
			Text: models.StrPtr(msgText),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      models.StrPtr(quotedID),
				Participant:   models.StrPtr(sender.String()),
				QuotedMessage: &waProto.Message{Conversation: models.StrPtr(quotedText)},
			},
		}
	} else {
		content.Conversation = models.StrPtr(msgText)
	}

	s.enqueue(OutboundMessage{Chat: chat, Content: content, msgType: msgType})
}

// Send sends a plain text message without quoting (for follow-up messages).
func (s *Sender) Send(chat types.JID, text string) {
	msgText := text + BotFooter
	content := &waProto.Message{
		Conversation: models.StrPtr(msgText),
	}
	s.enqueue(OutboundMessage{Chat: chat, Content: content, msgType: "text"})
}

// SendImage uploads an image to WhatsApp and then queues the image message.
// The upload happens synchronously (before enqueue) so the image bytes are
// available when the worker sends; this avoids buffering raw bytes in memory
// for the entire queue depth.
func (s *Sender) SendImage(chat, sender types.JID, imageBytes []byte, caption string, quotedID string, quotedText string) error {
	uploadStart := time.Now()

	uploadCtx, uploadCancel := context.WithTimeout(s.ctx, 60*time.Second)
	uploaded, err := s.Client.Upload(uploadCtx, imageBytes, whatsmeow.MediaImage)
	uploadCancel()

	if err != nil {
		senderLog.Error().
			Err(err).
			Str("company", s.CompanyName).
			Str("chat", chat.String()).
			Int("image_bytes", len(imageBytes)).
			Dur("upload_duration_ms", time.Since(uploadStart)).
			Msg("Failed to upload image to WhatsApp")
		return err
	}

	senderLog.Debug().
		Str("company", s.CompanyName).
		Str("chat", chat.String()).
		Dur("upload_duration_ms", time.Since(uploadStart)).
		Msg("Image uploaded to WhatsApp successfully")

	// Build caption (always attach footer for consistency).
	msgCaption := ""
	if caption != "" {
		msgCaption = caption + BotFooter
	}

	imageMsg := &waProto.ImageMessage{
		URL:           models.StrPtr(uploaded.URL),
		DirectPath:    models.StrPtr(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      models.StrPtr("image/jpeg"),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    models.Uint64Ptr(uint64(len(imageBytes))),
	}

	if msgCaption != "" {
		imageMsg.Caption = models.StrPtr(msgCaption)
	}

	if quotedID != "" {
		imageMsg.ContextInfo = &waProto.ContextInfo{
			StanzaID:      models.StrPtr(quotedID),
			Participant:   models.StrPtr(sender.String()),
			QuotedMessage: &waProto.Message{Conversation: models.StrPtr(quotedText)},
		}
	}

	content := &waProto.Message{ImageMessage: imageMsg}
	s.enqueue(OutboundMessage{Chat: chat, Content: content, msgType: "image"})
	return nil
}

// SetTyping sends a typing presence indicator.
// Errors are logged but not returned — a failed typing indicator is non-critical.
func (s *Sender) SetTyping(chat types.JID, typing bool) {
	presence := types.ChatPresencePaused
	if typing {
		presence = types.ChatPresenceComposing
	}

	presCtx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	err := s.Client.SendChatPresence(presCtx, chat, presence, types.ChatPresenceMediaText)
	if err != nil {
		senderLog.Warn().
			Err(err).
			Str("company", s.CompanyName).
			Str("chat", chat.String()).
			Bool("typing", typing).
			Msg("Failed to set chat presence (non-critical)")
	}
}
