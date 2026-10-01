package relay

import (
	"billionmail-core/internal/model/entity"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSmtpServiceName(t *testing.T) {
	tests := []struct {
		name   string
		config *entity.BmRelayConfig
		want   string
	}{
		{
			"custom smtp name",
			&entity.BmRelayConfig{Id: 1, SmtpName: "My Relay", RelayHost: "smtp.example.com"},
			"smtp_Custom_My_Relay",
		},
		{
			"relay with auth user containing @",
			&entity.BmRelayConfig{Id: 5, SmtpName: "Custom SMTP Relay", RelayHost: "smtp.gmail.com", AuthUser: "user@gmail.com"},
			"smtp_Relay_smtp_gmail_com_user_5",
		},
		{
			"relay with auth user no @",
			&entity.BmRelayConfig{Id: 3, SmtpName: "Custom SMTP Relay", RelayHost: "relay.host.com", AuthUser: "apikey12345"},
			"smtp_Relay_relay_host_com_apike_3",
		},
		{
			"relay with short auth user no @",
			&entity.BmRelayConfig{Id: 3, SmtpName: "Custom SMTP Relay", RelayHost: "relay.host.com", AuthUser: "ab"},
			"smtp_Relay_relay_host_com_ab_3",
		},
		{
			"relay no auth",
			&entity.BmRelayConfig{Id: 7, SmtpName: "Custom SMTP Relay", RelayHost: "10.0.0.1", AuthUser: ""},
			"smtp_Relay_10_0_0_1_7",
		},
		{
			"special chars in name",
			&entity.BmRelayConfig{Id: 1, SmtpName: "My !@# Relay", RelayHost: "smtp.example.com"},
			"smtp_Custom_My_Relay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSmtpServiceName(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommentOutOldTransportMaps(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"comments out matching line",
			"sender_dependent_default_transport_maps = hash:/etc/postfix/transport\n",
			"# sender_dependent_default_transport_maps = hash:/etc/postfix/transport # Migrated: replaced by automated config block\n",
		},
		{
			"skips inside automation block",
			"# BEGIN RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER\nsender_dependent_default_transport_maps = hash:/etc/postfix/transport\n# END RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER\n",
			"# BEGIN RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER\nsender_dependent_default_transport_maps = hash:/etc/postfix/transport\n# END RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER\n",
		},
		{
			"already commented",
			"# sender_dependent_default_transport_maps = hash:/etc/postfix/transport\n",
			"# sender_dependent_default_transport_maps = hash:/etc/postfix/transport\n",
		},
		{
			"empty input",
			"",
			"",
		},
		{
			"no matching lines",
			"some_other_config = value\nanother_config = value\n",
			"some_other_config = value\nanother_config = value\n",
		},
		{
			"pgsql variant",
			"sender_dependent_default_transport_maps = pgsql:/etc/postfix/pgsql.cf\n",
			"# sender_dependent_default_transport_maps = pgsql:/etc/postfix/pgsql.cf # Migrated: replaced by automated config block\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commentOutOldTransportMaps(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureSmtpsConfigInMasterCf(t *testing.T) {
	markerBegin := "# BEGIN BILLIONMAIL SMTPS CONFIG - DO NOT EDIT THIS MARKER"
	markerEnd := "# END BILLIONMAIL SMTPS CONFIG - DO NOT EDIT THIS MARKER"
	smtpsService := `smtps     unix  -       -       n       -       -       smtp
    -o smtp_tls_wrappermode=yes
    -o smtp_tls_security_level=encrypt`

	t.Run("adds block when missing", func(t *testing.T) {
		content := "existing config\n"
		result, modified := ensureSmtpsConfigInMasterCf(content)
		assert.True(t, modified)
		assert.Contains(t, result, markerBegin)
		assert.Contains(t, result, smtpsService)
		assert.Contains(t, result, markerEnd)
		assert.True(t, strings.HasPrefix(result, "existing config\n"))
	})

	t.Run("no-op when block exists with correct content", func(t *testing.T) {
		content := "before\n" + markerBegin + "\n" + smtpsService + "\n" + markerEnd + "\nafter\n"
		result, modified := ensureSmtpsConfigInMasterCf(content)
		assert.False(t, modified)
		assert.Equal(t, content, result)
	})

	t.Run("replaces block when content differs", func(t *testing.T) {
		content := "before\n" + markerBegin + "\nold stuff\n" + markerEnd + "\nafter\n"
		result, modified := ensureSmtpsConfigInMasterCf(content)
		assert.True(t, modified)
		assert.Contains(t, result, smtpsService)
		assert.Contains(t, result, "after")
	})

	t.Run("adds newline to content without trailing newline", func(t *testing.T) {
		content := "no trailing newline"
		result, modified := ensureSmtpsConfigInMasterCf(content)
		assert.True(t, modified)
		assert.Contains(t, result, markerBegin)
	})
}
