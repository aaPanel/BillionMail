package settings

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/settings/v1"
)

func (c *ControllerV1) Restart(ctx context.Context, req *v1.RestartReq) (res *v1.RestartRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
