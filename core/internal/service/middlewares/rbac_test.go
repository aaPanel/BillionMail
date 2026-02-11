package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathToRouteInfo(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		wantModule       string
		wantAction       string
		wantResource     string
	}{
		// Standard CRUD operations
		{"domains list", "/api/domains/list", "domains", "read", "domains"},
		{"domains detail", "/api/domains/detail/1", "domains", "read", "domains"},
		{"domains create", "/api/domains/create", "domains", "create", "domains"},
		{"domains update", "/api/domains/update", "domains", "update", "domains"},
		{"domains delete", "/api/domains/delete", "domains", "delete", "domains"},

		// All modules
		{"account", "/api/account/list", "account", "read", "account"},
		{"role", "/api/role/list", "role", "read", "role"},
		{"permission", "/api/permission/list", "permission", "read", "permission"},
		{"mail_boxes", "/api/mail_boxes/list", "mail_boxes", "read", "mail_boxes"},
		{"overview", "/api/overview/list", "overview", "read", "overview"},
		{"dockerapi", "/api/dockerapi/list", "dockerapi", "read", "dockerapi"},
		{"contact", "/api/contact/list", "contact", "read", "contact"},
		{"email_template", "/api/email_template/list", "email_template", "read", "email_template"},
		{"batch_mail", "/api/batch_mail/list", "batch_mail", "read", "batch_mail"},
		{"files", "/api/files/list", "files", "read", "files"},
		{"abnormal_recipient", "/api/abnormal_recipient/list", "abnormal_recipient", "read", "abnormal_recipient"},
		{"languages", "/api/languages/list", "languages", "read", "languages"},
		{"mail_services", "/api/mail_services/list", "mail_services", "read", "mail_services"},
		{"relay", "/api/relay/list", "relay", "read", "relay"},
		{"settings", "/api/settings/list", "settings", "read", "settings"},
		{"subscribe_list", "/api/subscribe_list/list", "subscribe_list", "read", "subscribe_list"},
		{"operation_log", "/api/operation_log/list", "operation_log", "read", "operation_log"},
		{"askai", "/api/askai/list", "askai", "read", "askai"},
		{"tags", "/api/tags/list", "tags", "read", "tags"},
		{"frostbyte", "/api/frostbyte/list", "frostbyte", "read", "frostbyte"},
		{"video_outreach", "/api/video_outreach/list", "video_outreach", "read", "video_outreach"},

		// Non-standard actions pass through
		{"custom action", "/api/domains/sync", "domains", "sync", "domains"},
		{"custom action 2", "/api/relay/test", "relay", "test", "relay"},

		// Edge cases
		{"empty path", "", "", "", ""},
		{"root only", "/", "", "", ""},
		{"no module match", "/api/unknown/list", "", "read", "unknown"},
		{"module as suffix", "/api/domains", "domains", "", ""},
		{"nested path", "/api/batch_mail/something/extra", "batch_mail", "something", "batch_mail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, action, resource := PathToRouteInfo(tt.path)
			assert.Equal(t, tt.wantModule, module, "module")
			assert.Equal(t, tt.wantAction, action, "action")
			assert.Equal(t, tt.wantResource, resource, "resource")
		})
	}
}
