package mail_boxes

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_boxes/v1"
)

func (c *ControllerV1) GetMailbox(ctx context.Context, req *v1.GetMailboxReq) (res *v1.GetMailboxRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
