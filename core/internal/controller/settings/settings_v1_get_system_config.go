package settings

import (
	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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

	ipWhitelist, err := g.DB().Model("bm_console_ip_whitelist").Fields("ip").All()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to read IP whitelist: {}", err)))
		return res, nil
	}
	if len(ipWhitelist) == 0 {
		config.IPWhitelist = []string{}
	} else {
		for _, ip := range ipWhitelist {
			config.IPWhitelist = append(config.IPWhitelist, ip["ip"].String())
		}
	}

	res.Data = config
	res.SetSuccess(public.LangCtx(ctx, "Successfully retrieved system configuration"))
	return res, nil
}
