// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package settings

import (
	"context"

	"billionmail-core/api/settings/v1"
)

type ISettingsV1 interface {
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
	Restart(ctx context.Context, req *v1.RestartReq) (res *v1.RestartRes, err error)
}
