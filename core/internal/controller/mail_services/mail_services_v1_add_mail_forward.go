package mail_services

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_services/v1"
)

func (c *ControllerV1) AddMailForward(ctx context.Context, req *v1.AddMailForwardReq) (res *v1.AddMailForwardRes, err error) {
	res = &v1.AddMailForwardRes{}

	count, err := g.DB().Model("alias").Where("address=?", req.Address).Count()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "check address failed: {}", err.Error())))
		return res, nil
	}

	if count > 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "forward address already exists")))
		return res, nil
	}

	domainParts := strings.Split(req.Address, "@")
	if len(domainParts) != 2 {
		res.SetError(gerror.New(public.LangCtx(ctx, "invalid email format: {}", req.Address)))
		return
	}
	domain := domainParts[1]
	if domain == "" {
		res.SetError(gerror.New(public.LangCtx(ctx, "domain part is empty in address: {}", req.Address)))
		return
	}

	count, err = g.DB().Model("domain").Where("domain=?", domain).Count()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "check domain failed: {}", err.Error())))
		return res, nil
	}

	if count == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "domain not found in mail server")))
		return res, nil
	}

	// process forward target addresses
	forwardUsers := processForwardUsers(req.Goto)
	if len(forwardUsers) == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "forward target address cannot be empty")))
		return res, nil
	}

	now := time.Now().Unix()

	_, err = g.DB().Model("alias").Insert(g.Map{
		"address":     req.Address,
		"goto":        strings.Join(forwardUsers, ","),
		"domain":      domain,
		"create_time": now,
		"update_time": now,
		"active":      req.Active,
	})

	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "add mail forward failed: {}", err.Error())))
		return res, nil
	}

	res.SetSuccess(public.LangCtx(ctx, "add mail forward success"))
	return res, nil
}
