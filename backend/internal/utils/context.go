package utils

import (
	"context"

	"webtracker-bot/internal/logger"
)

type contextKey string

const (
	// JIDKey is the context key for the user's JID.
	JIDKey         contextKey = "jid"
	// SenderPhoneKey is the context key for the sender's phone number.
	SenderPhoneKey contextKey = "sender_phone"
	// IsAdminKey is the context key to indicate if the user is an admin.
	IsAdminKey     contextKey = "is_admin"
	// ChatJIDKey is the context key for the chat JID.
	ChatJIDKey     contextKey = "chat_jid"
	// MessageIDKey is the context key for the message ID.
	MessageIDKey   contextKey = "message_id"
	// TextKey is the context key for the original message text.
	TextKey        contextKey = "text"
)

// GetJID safely extracts the sender's JID from context
func GetJID(ctx context.Context) string {
	val := ctx.Value(JIDKey)
	if val == nil {
		logger.Warn().Msg("Context missing 'jid'")
		return ""
	}
	s, ok := val.(string)
	if !ok {
		logger.Error().Msgf("Context 'jid' is not a string: %T", val)
		return ""
	}
	return s
}

// GetChatJID safely extracts the chat JID from context
func GetChatJID(ctx context.Context) string {
	val := ctx.Value(ChatJIDKey)
	if val == nil {
		logger.Warn().Msg("Context missing 'chat_jid'")
		return ""
	}
	s, ok := val.(string)
	if !ok {
		logger.Error().Msgf("Context 'chat_jid' is not a string: %T", val)
		return ""
	}
	return s
}

// GetMessageID safely extracts the message ID from context
func GetMessageID(ctx context.Context) string {
	val := ctx.Value(MessageIDKey)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// GetText safely extracts the original message text from context
func GetText(ctx context.Context) string {
	val := ctx.Value(TextKey)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// GetSenderPhone safely extracts the sender phone from context
func GetSenderPhone(ctx context.Context) string {
	val := ctx.Value(SenderPhoneKey)
	if val == nil {
		logger.Warn().Msg("Context missing 'sender_phone'")
		return ""
	}
	s, ok := val.(string)
	if !ok {
		logger.Error().Msgf("Context 'sender_phone' is not a string: %T", val)
		return ""
	}
	return s
}

// IsAdmin safely checks if the sender is an admin from context
func IsAdmin(ctx context.Context) bool {
	val := ctx.Value(IsAdminKey)
	if val == nil {
		return false
	}
	b, ok := val.(bool)
	if !ok {
		logger.Error().Msgf("Context 'is_admin' is not a bool: %T", val)
		return false
	}
	return b
}

// WithValues enriches context with multiple values using typed keys
func WithValues(ctx context.Context, jid, phone string, isAdmin bool, chatJid, msgID, text string) context.Context {
	ctx = context.WithValue(ctx, JIDKey, jid)
	ctx = context.WithValue(ctx, SenderPhoneKey, phone)
	ctx = context.WithValue(ctx, IsAdminKey, isAdmin)
	ctx = context.WithValue(ctx, ChatJIDKey, chatJid)
	ctx = context.WithValue(ctx, MessageIDKey, msgID)
	ctx = context.WithValue(ctx, TextKey, text)
	return ctx
}
