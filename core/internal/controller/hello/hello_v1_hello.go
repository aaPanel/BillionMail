package hello

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"

	"billionmail-core/api/hello/v1"
)

func (c *ControllerV1) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	g.RequestFromCtx(ctx).Response.Writeln("Hello World!")
	return
}
