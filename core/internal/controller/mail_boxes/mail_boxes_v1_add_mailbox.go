package mail_boxes

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_boxes/v1"
)

func (c *ControllerV1) AddMailbox(ctx context.Context, req *v1.AddMailboxReq) (res *v1.AddMailboxRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
