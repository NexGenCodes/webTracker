package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RPCRequestPayload defines the structure of a request sent over Redis
type RPCRequestPayload struct {
	ReplyChannel string      `json:"reply_channel"`
	Data         interface{} `json:"data"`
}

// RPCResponsePayload defines the structure of a response sent over Redis
type RPCResponsePayload struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// RPCRequest publishes a message to a requestChannel and waits for a response on replyChannel.
func RPCRequest(ctx context.Context, rdb *redis.Client, requestChannel string, data interface{}, timeout time.Duration) (*RPCResponsePayload, error) {
	// Create a unique reply channel
	replyChannel := fmt.Sprintf("%s:reply:%d", requestChannel, time.Now().UnixNano())

	req := RPCRequestPayload{
		ReplyChannel: replyChannel,
		Data:         data,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Subscribe to the reply channel before publishing to avoid race conditions
	pubsub := rdb.Subscribe(ctx, replyChannel)
	defer pubsub.Close()

	// Ensure subscription is active
	_, err = pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to reply channel: %w", err)
	}

	// Publish the request
	if err := rdb.Publish(ctx, requestChannel, string(reqBytes)).Err(); err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for response
	ch := pubsub.Channel()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("rpc timeout waiting for reply on %s", replyChannel)
	case msg := <-ch:
		var resp RPCResponsePayload
		if err := json.Unmarshal([]byte(msg.Payload), &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &resp, nil
	}
}

// RPCRespond publishes a response to a given replyChannel
func RPCRespond(ctx context.Context, rdb *redis.Client, replyChannel string, data interface{}, errStr string) error {
	resp := RPCResponsePayload{
		Data:  data,
		Error: errStr,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return rdb.Publish(ctx, replyChannel, string(respBytes)).Err()
}
