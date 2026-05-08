package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/proto"

	"webtracker-bot/internal/models"
)

const (
	TypeWhatsAppMessage = "whatsapp:message"
	TypeCronPulse       = "cron:pulse"
	TypeCronDailyStats  = "cron:daily_stats"
	TypeCronPruning     = "cron:pruning"
	TypeCronHealthCheck = "cron:health_check"
	TypeCronBotLiveness = "cron:bot_liveness"
	TypeOutboundAlert   = "whatsapp:outbound_alert"
	TypeCronCompanyPulse = "cron:company_pulse"
)

// CronCompanyPulsePayload defines the payload for processing a single company's status pulse.
type CronCompanyPulsePayload struct {
	CompanyID uuid.UUID `json:"company_id"`
}

// EnqueueCronCompanyPulse enqueues a status transition task for a specific company.
func EnqueueCronCompanyPulse(client *asynq.Client, companyID uuid.UUID) error {
	payload := CronCompanyPulsePayload{
		CompanyID: companyID,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal company pulse payload: %v", err)
	}

	// Max 1 retry, 3 minute timeout. Status processing should be fast but reliable.
	task := asynq.NewTask(TypeCronCompanyPulse, b, asynq.MaxRetry(1), asynq.Timeout(3*time.Minute))

	_, err = client.EnqueueContext(context.Background(), task)
	return err
}

// WhatsAppMessagePayload defines the payload for an incoming WhatsApp message.
type WhatsAppMessagePayload struct {
	CompanyID   uuid.UUID `json:"company_id"`
	ChatJID     string    `json:"chat_jid"`
	SenderJID   string    `json:"sender_jid"`
	MessageID   string    `json:"message_id"`
	Text        string    `json:"text"`
	SenderPhone string    `json:"sender_phone"`
	Language    string    `json:"language"`
	IsAdmin     bool      `json:"is_admin"`

	// Serialized protocol buffer message for things like Document retrieval
	RawMessageBytes []byte `json:"raw_message_bytes,omitempty"`
}

// EnqueueWhatsAppMessage enqueues a new WhatsApp message job
func EnqueueWhatsAppMessage(client *asynq.Client, job models.Job) error {
	payload := WhatsAppMessagePayload{
		CompanyID:   job.CompanyID,
		ChatJID:     job.ChatJID.String(),
		SenderJID:   job.SenderJID.String(),
		MessageID:   job.MessageID,
		Text:        job.Text,
		SenderPhone: job.SenderPhone,
		Language:    job.Language,
		IsAdmin:     job.IsAdmin,
	}

	// Safely serialize the raw protobuf message if it exists
	if job.RawMessage != nil && job.RawMessage.Message != nil {
		bytes, err := proto.Marshal(job.RawMessage.Message)
		if err == nil {
			payload.RawMessageBytes = bytes
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal whatsapp message payload: %v", err)
	}

	task := asynq.NewTask(TypeWhatsAppMessage, b, asynq.MaxRetry(3), asynq.Timeout(1*time.Minute))

	// Use MessageID as TaskID for idempotency — prevents processing same message twice
	_, err = client.EnqueueContext(context.Background(), task, asynq.TaskID(job.MessageID))
	if err != nil && err != asynq.ErrTaskIDConflict {
		return err
	}
	return nil
}

// OutboundAlertPayload defines the payload for an outbound WhatsApp status alert.
type OutboundAlertPayload struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	JIDStr      string    `json:"jid_str"`
	TrackingID  string    `json:"tracking_id"`
	Status      string    `json:"status"`
	Email       string    `json:"email"`
}

// EnqueueOutboundAlert enqueues a new outbound WhatsApp alert job
func EnqueueOutboundAlert(client *asynq.Client, companyID uuid.UUID, companyName, jidStr, tracking, status, email string) error {
	payload := OutboundAlertPayload{
		CompanyID:   companyID,
		CompanyName: companyName,
		JIDStr:      jidStr,
		TrackingID:  tracking,
		Status:      status,
		Email:       email,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbound alert payload: %v", err)
	}

	task := asynq.NewTask(TypeOutboundAlert, b, asynq.MaxRetry(5), asynq.Timeout(15*time.Second))

	_, err = client.EnqueueContext(context.Background(), task)
	return err
}
