package settings

import (
	"context"

	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
)

func (c *ControllerV1) SetSystemConfigKey(ctx context.Context, req *v1.SetSystemConfigKeyReq) (res *v1.SetSystemConfigKeyRes, err error) {
	res = &v1.SetSystemConfigKeyRes{}

	// 1. Read the current .env file content
	envMap, err := loadEnvFile()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to read environment variable file: {}", err)))
		return res, nil
	}

	// 2. Find the corresponding environment variable name based on the key
	_, modified := convertJsonKeyToEnvKey(req.Key, req.Value, envMap)
	if !modified {
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid configuration item or value not changed")))
		return res, nil
	}

	// 3. If there is a modification, write to the file
	if err := saveEnvFile(envMap); err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to update environment variable file: {}", err)))
		return res, nil
	}

	// 4. Restart the container
	g.Log().Info(ctx, public.LangCtx(ctx, "Configuration updated, preparing to restart container"))
	err = public.DockerApiFromCtx(ctx).RestartContainerByName(ctx, "billionmail-core-billionmail-1")
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to restart container: {}", err)))
		return res, nil
	}

	res.SetSuccess(public.LangCtx(ctx, "Configuration updated successfully"))
	return res, nil
}

// convertJsonKeyToEnvKey Convert JSON configuration key to environment variable key and update value
func convertJsonKeyToEnvKey(jsonKey string, value string, envMap map[string]string) (string, bool) {
	// JSON key to environment variable key mapping
	keyMapping := map[string]string{
		"admin_username":       "ADMIN_USERNAME",
		"admin_password":       "ADMIN_PASSWORD",
		"safe_path":            "SafePath",
		"billionmail_hostname": "BILLIONMAIL_HOSTNAME",
		"db_name":              "DBNAME",
		"db_user":              "DBUSER",
		"db_pass":              "DBPASS",
		"redis_pass":           "REDISPASS",
		"redis_port":           "REDIS_PORT",
		"smtp":                 "SMTP_PORT",
		"smtps":                "SMTPS_PORT",
		"submission":           "SUBMISSION_PORT",
		"imap":                 "IMAP_PORT",
		"imaps":                "IMAPS_PORT",
		"pop":                  "POP_PORT",
		"pops":                 "POPS_PORT",
		"http":                 "HTTP_PORT",
		"https":                "HTTPS_PORT",
		"timezone":             "TZ",
		"ipv4_network":         "IPV4_NETWORK",
		"fail2ban":             "FAIL2BAN_INIT",
	}

	// Get the environment variable key name
	envKey, exists := keyMapping[strings.ToLower(jsonKey)]
	if !exists {
		return "", false
	}

	// Special handling for fail2ban values
	if envKey == "FAIL2BAN_INIT" {
		if value == "1" {
			value = "y"
		} else if value == "0" {
			value = "n"
		}
	}

	// Check if the value has changed
	if currentValue, exists := envMap[envKey]; exists && currentValue == value {
		return envKey, false
	}

	// Update the environment variable map
	envMap[envKey] = value
	g.Log().Debug(nil, "Update configuration item %s to: %s", envKey, value)

	return envKey, true
}
