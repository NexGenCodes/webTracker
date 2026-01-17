package worker

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/supabase"
	"webtracker-bot/internal/utils"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type Worker struct {
	ID        int
	Client    *whatsmeow.Client
	DB        *supabase.Client
	Jobs      <-chan models.Job
	WG        *sync.WaitGroup
	GeminiKey string
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

	// 1. Check for Commands (!stats, !airwaybill)
	if strings.HasPrefix(job.Text, "!stats") {
		loc, _ := time.LoadLocation("Africa/Lagos") // Default Admin TZ
		pending, transit, err := w.DB.GetTodayStats(loc)
		if err != nil {
			w.sendReply(job.ChatJID, "❌ Failed to fetch stats", job.MessageID)
			return
		}
		msg := fmt.Sprintf("📊 *Today's Logistics*\n\n• PENDING: %d\n• IN_TRANSIT: %d\n\n_Total Created Today: %d_", pending, transit, pending+transit)
		w.sendReply(job.ChatJID, msg, job.MessageID)
		return
	}

	if strings.HasPrefix(job.Text, "!airwaybill") {
		parts := strings.Fields(job.Text)
		if len(parts) < 2 {
			w.sendReply(job.ChatJID, "💡 Usage: `!airwaybill AWB-XXXXX`", job.MessageID)
			return
		}
		id := parts[1]
		shipment, err := w.DB.GetShipment(id)
		if err != nil || shipment == nil {
			w.sendReply(job.ChatJID, "❌ Shipment not found", job.MessageID)
			return
		}
		wb := utils.GenerateWaybill(*shipment)
		w.sendReply(job.ChatJID, "```\n"+wb+"\n```", job.MessageID)
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

	if found < 2 {
		logger.Warn().Int("fields_found", found).Msg("Ignoring garbage message")
		return
	}

	// 3. Validation
	if len(m.MissingFields) > 0 {
		logger.GlobalVitals.IncParseFailure()
		msg := "⚠️ *Manifest Incomplete*\nMissing:\n• " + strings.Join(m.MissingFields, "\n• ")
		w.sendReply(job.ChatJID, msg, job.MessageID)
		return
	}
	logger.GlobalVitals.IncParseSuccess()

	// 4. Duplicate Check
	exists, tracking, err := w.DB.CheckDuplicate(m.ReceiverPhone)
	if err == nil && exists {
		logger.GlobalVitals.IncDuplicate()
		msg := fmt.Sprintf("⚠️ *Duplicate Found*\nID: *%s*", tracking)
		w.sendReply(job.ChatJID, msg, job.MessageID)
		return
	}

	// 5. Insert
	id, err := w.DB.InsertShipment(m, job.SenderPhone)
	if err != nil {
		logger.GlobalVitals.IncInsertFailure()
		w.sendReply(job.ChatJID, "❌ System Error: Saving failed", job.MessageID)
		return
	}
	logger.GlobalVitals.IncInsertSuccess()

	// 6. Success
	w.sendReply(job.ChatJID, fmt.Sprintf("✅ *Manifest Created*\nID: *%s*", id), job.MessageID)

	// Periodically log vitals (e.g., every 10 jobs)
	if logger.GlobalVitals.JobsProcessed%10 == 0 {
		logger.Info().Interface("vitals", logger.GlobalVitals.GetSnapshot()).Msg("Worker Vitals Snapshot")
	}
}

func (w *Worker) sendReply(jid types.JID, text string, quotedID string) {
	content := &waProto.Message{}

	if quotedID != "" {
		content.ExtendedTextMessage = &waProto.ExtendedTextMessage{
			Text: models.StrPtr(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      models.StrPtr(quotedID),
				Participant:   models.StrPtr(jid.String()),
				QuotedMessage: &waProto.Message{Conversation: models.StrPtr("Original Message")},
			},
		}
	} else {
		content.Conversation = models.StrPtr(text)
	}

	// Anti-Spam Jitter: 500ms to 2500ms
	jitter := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(jitter)

	_, _ = w.Client.SendMessage(context.Background(), jid, content)
}

// Note: I need to fix the context.Background() issue by passing it or importing it.
// I'll add the import and use context.Background() for now.
