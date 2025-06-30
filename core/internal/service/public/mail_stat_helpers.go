package public

import (
	"billionmail-core/internal/consts"
	"strings"
)

// GetMailProviderGroup Gets the mail provider group based on the email or host.
func GetMailProviderGroup(emailOrHost string) string {
	hostname := strings.ToLower(strings.TrimSpace(emailOrHost))
	if hostname == "" {
		return consts.MailProviderGroupOther
	}

	if strings.Contains(hostname, "@") {
		hostname = strings.SplitN(hostname, "@", 2)[1]
	}

	switch {
	case strings.Contains(hostname, "gmail.com") || strings.Contains(hostname, "googlemail.com") || hostname == "google.com":
		return consts.MailProviderGroupGmail
	case strings.Contains(hostname, "outlook.com") || strings.Contains(hostname, "hotmail.com") ||
		strings.Contains(hostname, "live.com") || strings.Contains(hostname, "msn.com") ||
		strings.HasPrefix(hostname, "outlook."):
		return consts.MailProviderGroupOutlook
	case strings.Contains(hostname, "yahoo.com") || strings.Contains(hostname, "ymail.com") ||
		strings.Contains(hostname, "rocketmail.com") || strings.HasPrefix(hostname, "yahoo.") ||
		strings.Contains(hostname, "aol.com") || strings.Contains(hostname, "verizonmedia.com"):
		return consts.MailProviderGroupYahoo
	case strings.Contains(hostname, "icloud.com") || strings.Contains(hostname, "me.com") ||
		strings.Contains(hostname, "mac.com") || strings.Contains(hostname, "apple.com"):
		return consts.MailProviderGroupApple
	case strings.Contains(hostname, "protonmail.com") || strings.Contains(hostname, "proton.me") ||
		strings.Contains(hostname, "pm.me"):
		return consts.MailProviderGroupProton
	case strings.Contains(hostname, "zoho.com") || strings.Contains(hostname, "zohomail.com") ||
		strings.Contains(hostname, "zohocorp.com") || strings.Contains(hostname, "zmail.com") ||
		strings.HasPrefix(hostname, "zoho."):
		return consts.MailProviderGroupZoho
	case strings.Contains(hostname, "kindle.com") || strings.Contains(hostname, "amazon.com") ||
		strings.Contains(hostname, "awsapps.com"):
		return consts.MailProviderGroupAmazon
	default:
		return consts.MailProviderGroupOther
	}
}
