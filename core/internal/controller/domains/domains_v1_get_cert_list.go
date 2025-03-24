package domains

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/domains/v1"
)

func (c *ControllerV1) GetCertList(ctx context.Context, req *v1.GetCertListReq) (res *v1.GetCertListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
