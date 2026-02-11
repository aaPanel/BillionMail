package frostbyte

import (
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	replyWebhookURL = "http://frostbyte-api:8001/webhooks/billionmail/reply"
	vmailBasePath   = "/opt/billionmail/vmail-data"
)

// ReplyPayload is the payload sent to FrostByte when a reply is detected.
type ReplyPayload struct {
	MessageID  string `json:"message_id"`
	FromEmail  string `json:"from_email"`
	ReplyBody  string `json:"reply_body"`
	InReplyTo  string `json:"in_reply_to"`
	ReceivedAt string `json:"received_at"`
}

// CheckForReplies scans Dovecot Maildir for new messages and matches them against
// sent message_ids in mailstat_message_ids, then forwards to FrostByte for classification.
func CheckForReplies() {
	ctx := context.Background()

	// Walk vmail directories looking for new messages
	err := filepath.Walk(vmailBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		// Only process files in /new/ directories (unread Maildir messages)
		dir := filepath.Dir(path)
		if filepath.Base(dir) != "new" || info.IsDir() {
			return nil
		}

		processNewMail(ctx, path)

		// Move to cur/ after processing (mark as read)
		curDir := filepath.Join(filepath.Dir(dir), "cur")
		if _, err := os.Stat(curDir); err == nil {
			newPath := filepath.Join(curDir, info.Name()+":2,S")
			_ = os.Rename(path, newPath)
		}

		return nil
	})

	if err != nil {
		g.Log().Debug(ctx, "frostbyte reply scan error:", err)
	}
}

func processNewMail(ctx context.Context, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		return
	}

	inReplyTo := strings.Trim(msg.Header.Get("In-Reply-To"), "<> ")
	references := msg.Header.Get("References")
	from := msg.Header.Get("From")

	// Parse from address
	if addr, err := mail.ParseAddress(from); err == nil {
		from = addr.Address
	}

	if inReplyTo == "" && references == "" {
		return // not a reply
	}

	// Try to match In-Reply-To against our sent message_ids
	messageID := ""
	candidates := []string{inReplyTo}
	for _, ref := range strings.Fields(references) {
		candidates = append(candidates, strings.Trim(ref, "<> "))
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		val, err := g.DB().Model("mailstat_message_ids").
			Where("message_id", candidate).
			Value("message_id")
		if err == nil && val.String() != "" {
			messageID = candidate
			break
		}
	}

	if messageID == "" {
		return // not a reply to one of our messages
	}

	// Read body (plain text portion)
	buf := make([]byte, 5000)
	n, _ := msg.Body.Read(buf)
	body := string(buf[:n])

	payload := ReplyPayload{
		MessageID:  messageID,
		FromEmail:  from,
		ReplyBody:  body,
		InReplyTo:  inReplyTo,
		ReceivedAt: nowISO(),
	}

	if err := dispatchReply(payload); err != nil {
		g.Log().Debug(ctx, "frostbyte reply dispatch failed:", err)
	} else {
		g.Log().Info(ctx, "frostbyte reply detected from:", from, "in-reply-to:", messageID)
	}
}

func dispatchReply(payload ReplyPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	client := public.GetHttpClient(10)
	resp, err := client.Post(replyWebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
