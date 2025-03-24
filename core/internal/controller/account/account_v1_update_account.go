package account

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/account/v1"
)

func (c *ControllerV1) UpdateAccount(ctx context.Context, req *v1.UpdateAccountReq) (res *v1.UpdateAccountRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
