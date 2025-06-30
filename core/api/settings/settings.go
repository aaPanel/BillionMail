// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package settings

import (
	"context"

	"billionmail-core/api/settings/v1"
)

type ISettingsV1 interface {
	GetVersion(ctx context.Context, req *v1.GetVersionReq) (res *v1.GetVersionRes, err error)
	GetSystemConfig(ctx context.Context, req *v1.GetSystemConfigReq) (res *v1.GetSystemConfigRes, err error)
	SetSystemConfig(ctx context.Context, req *v1.SetSystemConfigReq) (res *v1.SetSystemConfigRes, err error)
	SetSystemConfigKey(ctx context.Context, req *v1.SetSystemConfigKeyReq) (res *v1.SetSystemConfigKeyRes, err error)
	SetSSLConfig(ctx context.Context, req *v1.SetSSLConfigReq) (res *v1.SetSSLConfigRes, err error)
	GetTimeZoneList(ctx context.Context, req *v1.GetTimeZoneListReq) (res *v1.GetTimeZoneListRes, err error)
	SetIPWhitelist(ctx context.Context, req *v1.SetIPWhitelistReq) (res *v1.SetIPWhitelistRes, err error)
}
