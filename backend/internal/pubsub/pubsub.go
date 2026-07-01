package pubsub

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"webtracker-bot/internal/logger"
)

const (
	ChannelCompanyActivated = "wt:company:activated"
	ChannelRPCGetQR         = "wt:rpc:qr:req"
	ChannelRPCPairBot       = "wt:rpc:pair:req"
)

// PublishCompanyActivated sends a signal to the Redis channel to activate a bot instance
func PublishCompanyActivated(ctx context.Context, rdb *redis.Client, companyID uuid.UUID) error {
	return rdb.Publish(ctx, ChannelCompanyActivated, companyID.String()).Err()
}

// Subscribe listens for activation signals and calls the handler
// It runs continuously in the background until ctx is cancelled
func Subscribe(ctx context.Context, rdb *redis.Client, channel string, handler func(payload string)) {
	pubsub := rdb.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Str("channel", channel).Msg("Pub/Sub subscriber shutting down due to context cancellation")
			return
		case msg, ok := <-ch:
			if !ok {
				logger.Warn().Str("channel", channel).Msg("Pub/Sub channel closed, subscriber exiting")
				return
			}
			// Run handler in a new goroutine so a slow handler doesn't block the subscriber
			go handler(msg.Payload)
		}
	}
}
