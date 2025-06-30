package mail_services

import (
	"billionmail-core/internal/service/public"
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_services/v1"
)

func (c *ControllerV1) GetMailForwardList(ctx context.Context, req *v1.GetMailForwardListReq) (res *v1.GetMailForwardListRes, err error) {
	res = &v1.GetMailForwardListRes{}

	model := g.DB().Model("alias").Where("1=1")

	if req.Domain != "" {
		model = model.Where("domain=?", req.Domain)
	}

	if req.SearchKey != "" {
		model = model.Where("(address LIKE ? OR goto LIKE ?)", "%"+req.SearchKey+"%", "%"+req.SearchKey+"%")
	}

	count, err := model.Count()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get email forwarding list: {}", err.Error())))
		return res, nil
	}
	res.Data.Total = count

	if count == 0 {
		res.Data.List = make([]*v1.Alias, 0)
		res.SetSuccess(public.LangCtx(ctx, "Get email forwarding list successfully"))
		return res, nil
	}

	err = model.Page(req.Page, req.PageSize).OrderDesc("create_time").Scan(&res.Data.List)
	if err != nil && err != sql.ErrNoRows {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get email forwarding list: {}", err.Error())))
		return res, nil
	}

	res.SetSuccess(public.LangCtx(ctx, "Get email forwarding list successfully"))
	return res, nil
}
