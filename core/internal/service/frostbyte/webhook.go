package frostbyte

import (
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const webhookURL = "http://frostbyte-api:8001/webhooks/billionmail/tracking"

// TrackingEvent represents a tracking event to dispatch to FrostByte.
type TrackingEvent struct {
	Event     string `json:"event"`      // open, click, bounce, complaint
	MessageID string `json:"message_id"` // BillionMail message-id header value
	Recipient string `json:"recipient"`
	Timestamp string `json:"timestamp"`
}

// DispatchTrackingEvent sends a tracking event to FrostByte API asynchronously with retry.
func DispatchTrackingEvent(ctx context.Context, event TrackingEvent) {
	go func() {
		if err := doDispatchWithRetry(ctx, event); err != nil {
			g.Log().Warning(ctx, "frostbyte webhook dispatch failed after retries: %v", err)
		}
	}()
}

func doDispatchWithRetry(ctx context.Context, event TrackingEvent) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return fmt.Errorf("context canceled during retry: %w", lastErr)
			}
		}
		lastErr = doDispatch(event)
		if lastErr == nil {
			return nil
		}
		g.Log().Debug(ctx, "frostbyte webhook attempt %d failed: %v", attempt+1, lastErr)
	}
	return fmt.Errorf("all 3 attempts failed: %w", lastErr)
}

func doDispatch(event TrackingEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	client := public.GetHttpClient(10)
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

// nowISO returns current time in ISO 8601 format.
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// DispatchOpen dispatches an open event for the given message.
func DispatchOpen(ctx context.Context, messageID, recipient string) {
	DispatchTrackingEvent(ctx, TrackingEvent{
		Event:     "open",
		MessageID: messageID,
		Recipient: recipient,
		Timestamp: nowISO(),
	})
}

// DispatchClick dispatches a click event for the given message.
func DispatchClick(ctx context.Context, messageID, recipient string) {
	DispatchTrackingEvent(ctx, TrackingEvent{
		Event:     "click",
		MessageID: messageID,
		Recipient: recipient,
		Timestamp: nowISO(),
	})
}

// DispatchBounce dispatches a bounce event for the given message.
func DispatchBounce(ctx context.Context, messageID, recipient string) {
	DispatchTrackingEvent(ctx, TrackingEvent{
		Event:     "bounce",
		MessageID: messageID,
		Recipient: recipient,
		Timestamp: nowISO(),
	})
}
