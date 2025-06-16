package mail_service

import (
	"billionmail-core/internal/consts"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"path/filepath"
	"strings"
)

func FixPostfixMainConfig(ctx context.Context) {
	g.Log().Debug(ctx, "Fix postfix main config")
	defer func() {
		g.Log().Debug(ctx, "Fix postfix main config done")
	}()

	var err error

	// initialize a header_checks file if it does not exist
	headerChecksFile := public.AbsPath(filepath.Join(consts.POSTFIX_CONF_PATH, "header_checks"))
	headerChecksContent := ""
	regularHeaderChecks := `if /^\s*Received:.*Authenticated sender.*\(PostBillionMail\)/
/^Received: from .*? \([\w\-.]* \[.*?\]\)(.*|\n.*)\(Authenticated sender: (.+)\)\s+by.+\(PostBillionMail\) with (.*)/
  REPLACE Received: from [127.0.0.1] (localhost [127.0.0.1]) by localhost (Mailerdaemon) with $3
endif

/^\s*X-Enigmail/        IGNORE
/^\s*X-Originating-IP/  IGNORE
/^\s*X-Forward/         IGNORE
`

	if !public.FileExists(headerChecksFile) {
		headerChecksContent = regularHeaderChecks
		_, err = public.WriteFile(headerChecksFile, headerChecksContent)

		if err != nil {
			g.Log().Warning(ctx, "Failed to create header_checks file: %v", err)
			return
		}
	} else {
		// Read the existing header_checks file content
		headerChecksContent, err = public.ReadFile(headerChecksFile)

		if err != nil {
			g.Log().Warning(ctx, "Failed to read header_checks file: %v", err)
			return
		}
	}

	// Ensure Received header replacement is in place
	if !strings.Contains(headerChecksContent, "REPLACE Received: from [127.0.0.1] (localhost [127.0.0.1]) by localhost (Mailerdaemon) with $3") {
		headerChecksContent += "\n" + regularHeaderChecks
		_, err = public.WriteFile(headerChecksFile, headerChecksContent)

		if err != nil {
			g.Log().Warning(ctx, "Failed to update header_checks file with Received header replacement: %v", err)
			return
		}
	}

	lines := make([]string, 0)
	skipNextEmptyLine := false

	// Read the main configuration file each rows
	err = public.ReadEach(consts.POSTFIX_MAIN_CONF, func(row string, cnt int) bool {
		if strings.HasPrefix(strings.TrimSpace(row), "smtp_header_checks") || strings.HasPrefix(strings.TrimSpace(row), "lmtp_header_checks") || strings.HasPrefix(strings.TrimSpace(row), "mail_name") {
			// Skip existing smtp_header_checks, lmtp_header_checks lines, and mail_name line
			skipNextEmptyLine = true
			return true
		}

		if skipNextEmptyLine {
			skipNextEmptyLine = false
			if strings.TrimSpace(row) == "" {
				return true
			}
		}

		if strings.HasPrefix(strings.TrimSpace(row), "compatibility_level") {
			lines = append(lines, "compatibility_level = 3.7\n")
			lines = append(lines, "mail_name = PostBillionMail\n")
			lines = append(lines, "\n")
			lines = append(lines, "smtp_header_checks = pcre:/etc/postfix/conf/header_checks\n")
			lines = append(lines, "lmtp_header_checks = pcre:/etc/postfix/conf/header_checks\n\n")
			return true
		}

		lines = append(lines, row)

		return true
	})

	if err != nil {
		g.Log().Warning(ctx, "Failed to read Postfix main configuration file: %v", err)
		return
	}

	// Write the updated lines back to the main configuration file
	_, err = public.WriteFile(consts.POSTFIX_MAIN_CONF, strings.Join(lines, ""))

	if err != nil {
		g.Log().Warning(ctx, "Failed to write updated Postfix main configuration file: %v", err)
		return
	}

	dk, err := docker.NewDockerAPI()

	if err != nil {
		g.Log().Warning(ctx, "Failed to create Docker API instance: %v", err)
		return
	}

	defer dk.Close()

	// Restart the Postfix container to apply the changes
	err = dk.RestartContainer(ctx, "billionmail-postfix-billionmail-1")

	if err != nil {
		g.Log().Warning(ctx, "Failed to restart Postfix container: %v", err)
		return
	}
}

func FixRspamdDKIMSigningConfig(ctx context.Context) {
	g.Log().Debug(ctx, "Fix rspamd dkim signing config")
	defer func() {
		g.Log().Debug(ctx, "Fix rspamd dkim signing config done")
	}()

	lines := make([]string, 0)

	// Read the DKIM signing configuration file each rows
	err := public.ReadEach(public.AbsPath(filepath.Join(consts.RSPAMD_LOCAL_D_PATH, "dkim_signing.conf")), func(row string, cnt int) bool {
		if strings.HasPrefix(strings.TrimSpace(row), "sign_headers") {
			return true
		}

		lines = append(lines, row)

		return true
	})

	if err != nil {
		g.Log().Warning(ctx, "Failed to read DKIM signing configuration file: %v", err)
		return
	}

	lineLen := len(lines)
	lineBegin := -1

	for i := 0; i < lineLen; i++ {
		lineBegin = i
		if strings.TrimSpace(lines[i]) != "" {
			break
		}
	}

	if lineBegin > 0 {
		lines = lines[lineBegin:lineLen]
	}

	lines = append([]string{"sign_headers = \"from:sender:reply-to:subject:date:message-id:to:cc:mime-version:content-type:content-transfer-encoding:content-language:resent-to:resent-cc:resent-from:resent-sender:resent-message-id:in-reply-to:references:list-id:list-help:list-owner:list-unsubscribe:list-subscribe:list-post:list-unsubscribe-post:disposition-notification-to:disposition-notification-options:original-recipient:openpgp:autocrypt\";\n\n"}, lines...)

	// Write the updated lines back to the DKIM signing configuration file
	_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.RSPAMD_LOCAL_D_PATH, "dkim_signing.conf")), strings.Join(lines, ""))

	if err != nil {
		g.Log().Warning(ctx, "Failed to write updated DKIM signing configuration file: %v", err)
		return
	}

	dk, err := docker.NewDockerAPI()

	if err != nil {
		g.Log().Warning(ctx, "Failed to create Docker API instance: %v", err)
		return
	}

	defer dk.Close()

	// Restart the Rspamd container to apply the changes
	err = dk.RestartContainer(ctx, "billionmail-rspamd-billionmail-1")

	if err != nil {
		g.Log().Warning(ctx, "Failed to restart Rspamd container: %v", err)
		return
	}
}

func FixDovecotSSLConfig(ctx context.Context) {
	g.Log().Debug(ctx, "Fix Dovecot SSL config")
	defer func() {
		g.Log().Debug(ctx, "Fix Dovecot SSL config done")
	}()

	lines := make([]string, 0)
	appendLines := make([]string, 0, 1)
	appendLineFlag := false

	appendLines = append(appendLines, "\n")

	// Read the Dovecot SSL configuration file each rows
	err := public.ReadEach(public.AbsPath(filepath.Join(consts.DOVECOT_CONF_D_PATH, "10-ssl.conf")), func(row string, cnt int) bool {
		if strings.HasPrefix(strings.TrimSpace(row), "#DOMAIN_SSL_END_") {
			appendLineFlag = false
			appendLines = append(appendLines, row)
			return true
		}

		if appendLineFlag || strings.HasPrefix(strings.TrimSpace(row), "#DOMAIN_SSL_BEGIN_") {
			appendLineFlag = true
			appendLines = append(appendLines, row)
			return true
		}

		lines = append(lines, row)
		return true
	})

	if err != nil {
		g.Log().Warning(ctx, "Failed to read Dovecot SSL configuration file: %v", err)
		return
	}

	lineLen := len(lines)
	lineEnd := -1

	for i := lineLen - 1; i > -1; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			break
		}
		lineEnd = i
	}

	if lineEnd > 0 {
		lines = lines[:lineEnd]
	}

	lines = append(lines, appendLines...)

	// Write the updated lines back to the Dovecot SSL configuration file
	_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.DOVECOT_CONF_D_PATH, "10-ssl.conf")), strings.Join(lines, ""))

	if err != nil {
		g.Log().Warning(ctx, "Failed to write updated Dovecot SSL configuration file: %v", err)
		return
	}

	dk, err := docker.NewDockerAPI()

	if err != nil {
		g.Log().Warning(ctx, "Failed to create Docker API instance: %v", err)
		return
	}

	defer dk.Close()

	// Restart the Dovecot container to apply the changes
	err = dk.RestartContainer(ctx, "billionmail-dovecot-billionmail-1")

	if err != nil {
		g.Log().Warning(ctx, "Failed to restart Dovecot container: %v", err)
		return
	}
}
