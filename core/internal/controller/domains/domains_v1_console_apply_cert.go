package domains

import (
	"billionmail-core/api/domains/v1"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) ConsoleApplyCert(ctx context.Context, req *v1.ConsoleApplyCertReq) (res *v1.ConsoleApplyCertRes, err error) {
	res = &v1.ConsoleApplyCertRes{}

	err = domains.ApplyConsoleCert(ctx)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed : {}", err.Error())))
		return
	}

	res.SetSuccess("Certificate applied successfully")
	return
}
