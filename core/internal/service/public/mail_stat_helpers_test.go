package public

import (
	"billionmail-core/internal/consts"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMailProviderGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Gmail
		{name: "gmail.com email", input: "user@gmail.com", expected: consts.MailProviderGroupGmail},
		{name: "gmail.com host", input: "gmail.com", expected: consts.MailProviderGroupGmail},
		{name: "googlemail.com", input: "user@googlemail.com", expected: consts.MailProviderGroupGmail},
		{name: "google.com", input: "google.com", expected: consts.MailProviderGroupGmail},

		// Outlook
		{name: "outlook.com", input: "user@outlook.com", expected: consts.MailProviderGroupOutlook},
		{name: "hotmail.com", input: "user@hotmail.com", expected: consts.MailProviderGroupOutlook},
		{name: "live.com", input: "user@live.com", expected: consts.MailProviderGroupOutlook},
		{name: "msn.com", input: "user@msn.com", expected: consts.MailProviderGroupOutlook},
		{name: "outlook.jp", input: "outlook.jp", expected: consts.MailProviderGroupOutlook},

		// Yahoo
		{name: "yahoo.com", input: "user@yahoo.com", expected: consts.MailProviderGroupYahoo},
		{name: "ymail.com", input: "user@ymail.com", expected: consts.MailProviderGroupYahoo},
		{name: "rocketmail.com", input: "user@rocketmail.com", expected: consts.MailProviderGroupYahoo},
		{name: "aol.com", input: "user@aol.com", expected: consts.MailProviderGroupYahoo},
		{name: "yahoo.co.jp", input: "yahoo.co.jp", expected: consts.MailProviderGroupYahoo},

		// Apple
		{name: "icloud.com", input: "user@icloud.com", expected: consts.MailProviderGroupApple},
		{name: "me.com", input: "user@me.com", expected: consts.MailProviderGroupApple},
		{name: "mac.com", input: "user@mac.com", expected: consts.MailProviderGroupApple},
		{name: "apple.com", input: "user@apple.com", expected: consts.MailProviderGroupApple},

		// Proton
		{name: "protonmail.com", input: "user@protonmail.com", expected: consts.MailProviderGroupProton},
		{name: "proton.me", input: "user@proton.me", expected: consts.MailProviderGroupProton},
		{name: "pm.me", input: "user@pm.me", expected: consts.MailProviderGroupProton},

		// Zoho
		{name: "zoho.com", input: "user@zoho.com", expected: consts.MailProviderGroupZoho},
		{name: "zohomail.com", input: "user@zohomail.com", expected: consts.MailProviderGroupZoho},
		{name: "zoho.eu", input: "zoho.eu", expected: consts.MailProviderGroupZoho},

		// Amazon
		{name: "kindle.com", input: "user@kindle.com", expected: consts.MailProviderGroupAmazon},
		{name: "amazon.com", input: "user@amazon.com", expected: consts.MailProviderGroupAmazon},
		{name: "awsapps.com", input: "user@awsapps.com", expected: consts.MailProviderGroupAmazon},

		// Other
		{name: "custom domain", input: "user@example.com", expected: consts.MailProviderGroupOther},
		{name: "unknown host", input: "unknown.org", expected: consts.MailProviderGroupOther},
		{name: "empty string", input: "", expected: consts.MailProviderGroupOther},
		{name: "whitespace", input: "  ", expected: consts.MailProviderGroupOther},
		{name: "company domain", input: "user@company.io", expected: consts.MailProviderGroupOther},

		// Case insensitivity
		{name: "GMAIL uppercase", input: "user@GMAIL.COM", expected: consts.MailProviderGroupGmail},
		{name: "Gmail mixed case", input: "user@Gmail.Com", expected: consts.MailProviderGroupGmail},
		{name: "OUTLOOK uppercase", input: "user@OUTLOOK.COM", expected: consts.MailProviderGroupOutlook},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMailProviderGroup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMailProviderGroup_HostOnly(t *testing.T) {
	// Test with bare hostnames (no @ sign)
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{name: "gmail host", host: "gmail.com", expected: consts.MailProviderGroupGmail},
		{name: "outlook host", host: "outlook.com", expected: consts.MailProviderGroupOutlook},
		{name: "yahoo host", host: "yahoo.com", expected: consts.MailProviderGroupYahoo},
		{name: "random host", host: "mycompany.com", expected: consts.MailProviderGroupOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMailProviderGroup(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMailProviderGroup_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "at sign only", input: "@", expected: consts.MailProviderGroupOther},
		{name: "at with no domain", input: "user@", expected: consts.MailProviderGroupOther},
		{name: "multiple at signs", input: "user@name@gmail.com", expected: consts.MailProviderGroupGmail},
		{name: "subdomain of gmail", input: "user@sub.gmail.com", expected: consts.MailProviderGroupGmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMailProviderGroup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
