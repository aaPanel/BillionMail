package account

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/account/v1"
)

func (c *ControllerV1) GetAccounts(ctx context.Context, req *v1.GetAccountsReq) (res *v1.GetAccountsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
