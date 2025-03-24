package domains

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/domains/v1"
)

func (c *ControllerV1) SetSSL(ctx context.Context, req *v1.SetSSLReq) (res *v1.SetSSLRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
