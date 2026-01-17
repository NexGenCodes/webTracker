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

	// AI Fallback
	if len(m.MissingFields) > 0 {
		aiM, err := parser.ParseAI(job.Text, w.GeminiKey)
		if err == nil {
			if m.ReceiverName == "" {
				m.ReceiverName = aiM.ReceiverName
			}
			if m.ReceiverAddress == "" {
				m.ReceiverAddress = aiM.ReceiverAddress
			}
			if m.ReceiverPhone == "" {
				m.ReceiverPhone = aiM.ReceiverPhone
			}
			if m.ReceiverCountry == "" {
				m.ReceiverCountry = aiM.ReceiverCountry
			}
			if m.SenderName == "" {
				m.SenderName = aiM.SenderName
			}
			if m.SenderCountry == "" {
				m.SenderCountry = aiM.SenderCountry
			}

			// Recalculate missing
			m.MissingFields = []string{}
			if m.ReceiverName == "" {
				m.MissingFields = append(m.MissingFields, "Receiver Name")
			}
			if m.ReceiverAddress == "" {
				m.MissingFields = append(m.MissingFields, "Receiver Address")
			}
			if m.ReceiverPhone == "" {
				m.MissingFields = append(m.MissingFields, "Receiver Phone")
			}
			if m.ReceiverCountry == "" {
				m.MissingFields = append(m.MissingFields, "Receiver Country")
			}
			if m.SenderName == "" {
				m.MissingFields = append(m.MissingFields, "Sender Name")
			}
			if m.SenderCountry == "" {
				m.MissingFields = append(m.MissingFields, "Sender Country")
			}
		}
	}

	// 2. Threshold check (Min 2 fields)
	found := 0
	if m.ReceiverName != "" {
		found++
	}
	if m.ReceiverPhone != "" {
		found++
	}
	if m.ReceiverAddress != "" {
		found++
	}
	if m.ReceiverCountry != "" {
		found++
	}
	if m.SenderName != "" {
		found++
	}
	if m.SenderCountry != "" {
		found++
	}

	// 3. Validation
	if len(m.MissingFields) > 0 {
		logger.GlobalVitals.IncParseFailure()
		msg := "⚠️ *Manifest Incomplete*\nMissing:\n• " + strings.Join(m.MissingFields, "\n• ")
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
	w.sendReply(job.ChatJID, job.SenderJID, fmt.Sprintf("✅ *Manifest Created*\nID: *%s*", id), job.MessageID)
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

	_, _ = w.Client.SendMessage(context.Background(), chat, content)
}
