package v1

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestEditMailForwardReq_GotoAcceptsNewlineSeparatedEmails(t *testing.T) {
	tests := []struct {
		name         string
		destinations string
	}{
		{"single email", "test1@example.com"},
		{"newline separated", "test1@example.com\ntest2@example.com"},
		{"escaped newline separated", "test1@example.com\\ntest2@example.com"},
		{"trailing newline", "test1@example.com\ntest2@example.com\n"},
		{"empty means status-only edit", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := EditMailForwardReq{
				Address: "src@example.com",
				Goto:    tt.destinations,
				Active:  1,
			}
			err := g.Validator().Data(req).Run(context.Background())
			if err != nil {
				t.Fatalf("EditMailForwardReq.Goto %q should be valid, got: %v", tt.destinations, err)
			}
		})
	}
}

func TestAddMailForwardReq_GotoAcceptsNewlineSeparatedEmails(t *testing.T) {
	req := AddMailForwardReq{
		Address: "src@example.com",
		Goto:    "test1@example.com\ntest2@example.com",
		Active:  1,
	}
	err := g.Validator().Data(req).Run(context.Background())
	if err != nil {
		t.Fatalf("AddMailForwardReq already accepts newline-separated Goto, got: %v", err)
	}
}
