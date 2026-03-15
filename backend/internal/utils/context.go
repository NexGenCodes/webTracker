package utils

import (
	"context"

	"webtracker-bot/internal/logger"
)

// GetJID safely extracts the JID from context
func GetJID(ctx context.Context) string {
	val := ctx.Value("jid")
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

// GetSenderPhone safely extracts the sender phone from context
func GetSenderPhone(ctx context.Context) string {
	val := ctx.Value("sender_phone")
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
	val := ctx.Value("is_admin")
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
