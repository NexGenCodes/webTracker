package worker

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"webtracker-bot/internal/commands"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/supabase"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type Worker struct {
	ID          int
	Client      *whatsmeow.Client
	DB          *supabase.Client
	Jobs        <-chan models.Job
	WG          *sync.WaitGroup
	GeminiKey   string
	AwbCmd      string
	CompanyName string
	Cmd         *commands.Dispatcher
}

func (w *Worker) Start() {
	defer w.WG.Done()
	logger.Info().Int("worker_id", w.ID).Msg("Worker started")

	for job := range w.Jobs {
		w.process(job)
	}
}

func (w *Worker) process(job models.Job) {
	logger.GlobalVitals.IncJobs()

	// 1. Check for Commands (Explicit)
	if reply, ok := w.Cmd.Dispatch(context.Background(), job.Text); ok {
		w.sendReply(job.ChatJID, job.SenderJID, reply, job.MessageID)
		return
	}

	// 2. High-Performance Pre-filter
	// Instantly ignore messages that don't look like manifests/shipping info.
	if !w.isPotentialManifest(job.Text) {
		return
	}

	// 2. Normal Parsing (Regex first)
	m := parser.ParseRegex(job.Text)

	// AI Fallback if fields are missing
	if len(m.MissingFields) > 0 {
		if aiM, err := parser.ParseAI(job.Text, w.GeminiKey); err == nil {
			m.Merge(aiM)
			m.IsAI = true
			m.Validate()
		}
	}

	// 3. Validation
	if len(m.MissingFields) > 0 {
		logger.GlobalVitals.IncParseFailure()
		msg := "⚠️ *Information Incomplete*\nMissing:\n• " + strings.Join(m.MissingFields, "\n• ")
		w.sendReply(job.ChatJID, job.SenderJID, msg, job.MessageID)
		return
	}
	logger.GlobalVitals.IncParseSuccess()

	// 4. Duplicate Check
	exists, tracking, err := w.DB.CheckDuplicate(m.ReceiverPhone)
	if err == nil && exists {
		logger.GlobalVitals.IncDuplicate()
		msg := fmt.Sprintf("⚠️ *Duplicate Found*\nID: *%s*", tracking)
		w.sendReply(job.ChatJID, job.SenderJID, msg, job.MessageID)
		return
	}

	// 5. Insert
	id, err := w.DB.InsertShipment(m, job.SenderPhone)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		w.sendReply(job.ChatJID, job.SenderJID, "❌ System Error: Saving failed", job.MessageID)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 6. Success
	successMsg := fmt.Sprintf("✅ *Manifest Created*\nID: *%s*", id)
	if m.IsAI {
		successMsg += "\n_✨ (AI Enhanced)_"
	}
	w.sendReply(job.ChatJID, job.SenderJID, successMsg, job.MessageID)
}

func (w *Worker) isPotentialManifest(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"sender", "receiver", "destin", "phone", "name", "address", "country", "cargo", "ship"}
	count := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			count++
		}
	}
	return count >= 2
}

func (w *Worker) sendReply(chat, sender types.JID, text string, quotedID string) {
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

	resp, err := w.Client.SendMessage(context.Background(), chat, content)
	if err != nil {
		logger.Error().Err(err).Str("chat", chat.String()).Msg("Failed to send WhatsApp message")
	} else {
		logger.Debug().Str("resp_id", resp.ID).Msg("Message sent successfully")
	}
}
