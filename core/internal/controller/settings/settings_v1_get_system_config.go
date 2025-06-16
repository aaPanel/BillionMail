package settings

import (
	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetSystemConfig(ctx context.Context, req *v1.GetSystemConfigReq) (res *v1.GetSystemConfigRes, err error) {
	res = &v1.GetSystemConfigRes{}

	envMap, err := public.LoadEnvFile()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to read environment variable file: {}", err)))
		return res, nil
	}

	config := convertEnvToConfig(envMap)

	sslInfo, err := loadSSLInfo()
	if err == nil {
		config.SSL = *sslInfo
	}
	serverIP, err := public.GetServerIP()
	if err != nil {
		serverIP = "unknown"
	}
	config.ServerIP = serverIP

	// TODO: Whitelist ip

	res.Data = config
	res.SetSuccess(public.LangCtx(ctx, "Successfully retrieved system configuration"))
	return res, nil
}
