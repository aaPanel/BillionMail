package settings

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/os/gfile"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/settings/v1"
)

func (c *ControllerV1) SetSystemConfig(ctx context.Context, req *v1.SetSystemConfigReq) (res *v1.SetSystemConfigRes, err error) {
	res = &v1.SetSystemConfigRes{}

	envMap, err := loadEnvFile()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to read environment variable file: {}", err)))
		return res, nil
	}

	// 2. Update the configuration items that need to be modified
	modified := updateEnvMap(envMap, req)

	// 3. If there is a modification, write to the file
	if modified {

		if err := saveEnvFile(envMap); err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to update environment variable file: {}", err)))
			return res, nil
		}

		err = public.DockerApiFromCtx(ctx).RestartContainerByName(ctx, "billionmail-core-billionmail-1")

		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to restart container: {}", err)))
			return res, nil
		}
	}

	res.SetSuccess(public.LangCtx(ctx, "Configuration updated successfully"))
	return res, nil
}

// updateEnvMap Update the environment variable map
func updateEnvMap(envMap map[string]string, req *v1.SetSystemConfigReq) bool {
	modified := false

	// Basic configuration
	if req.AdminUsername != "" && req.AdminUsername != envMap["ADMIN_USERNAME"] {
		envMap["ADMIN_USERNAME"] = req.AdminUsername
		modified = true

	}
	if req.AdminPassword != "" && req.AdminPassword != envMap["ADMIN_PASSWORD"] {
		envMap["ADMIN_PASSWORD"] = req.AdminPassword
		modified = true

	}
	if req.SafePath != "" && req.SafePath != envMap["SafePath"] {
		envMap["SafePath"] = req.SafePath
		modified = true

	}
	if req.Hostname != "" && req.Hostname != envMap["BILLIONMAIL_HOSTNAME"] {
		envMap["BILLIONMAIL_HOSTNAME"] = req.Hostname
		modified = true

	}

	return modified
}

// saveEnvFile Save the environment variable to the file
func saveEnvFile(envMap map[string]string) error {
	// 1. Read the original file content
	content, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	// 2. Process by line
	var newLines []string
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Keep comment lines and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			newLines = append(newLines, line)
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		// If envMap has this key, update the value
		if val, ok := envMap[key]; ok {

			newLines = append(newLines, key+"="+val)
			delete(envMap, key) // Mark as processed
		} else {
			newLines = append(newLines, line)
		}
	}

	// 3. Add new key-value pairs in envMap
	for k, v := range envMap {
		newLines = append(newLines, k+"="+v)
	}

	// 4. Write to file
	return gfile.PutContents(envPath, strings.Join(newLines, "\n"))
}
