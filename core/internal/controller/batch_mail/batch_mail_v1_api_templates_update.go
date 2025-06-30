package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"net"
	"strings"
	"time"
)

func (c *ControllerV1) ApiTemplatesUpdate(ctx context.Context, req *v1.ApiTemplatesUpdateReq) (res *v1.ApiTemplatesUpdateRes, err error) {
	res = &v1.ApiTemplatesUpdateRes{}

	// verify if API exists

	count, err := g.DB().Model("api_templates").Where("id", req.ID).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New(public.LangCtx(ctx, "API does not exist"))
	}

	// verify if template exists
	count, err = g.DB().Model("email_templates").Where("id", req.TemplateId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New(public.LangCtx(ctx, "Email template does not exist"))
	}

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// current time
	now := time.Now().Unix()

	updateMap := g.Map{
		"api_name":    req.ApiName,
		"template_id": req.TemplateId,
		"subject":     req.Subject,
		"addresser":   req.Addresser,
		"full_name":   req.FullName,
		"unsubscribe": req.Unsubscribe,
		"track_open":  req.TrackOpen,
		"track_click": req.TrackClick,
		"active":      req.Active,
		"expire_time": req.ExpireTime,
		"update_time": now,
		//"ip_whitelist_enabled": req.IpWhitelistEnabled,
	}

	_, err = tx.Model("api_templates").
		Where("id", req.ID).
		Update(updateMap)

	if err != nil {
		return nil, err
	}

	// ip
	_, err = tx.Model("api_ip_whitelist").
		Where("api_id", req.ID).
		Delete()
	if err != nil {
		g.Log().Errorf(ctx, "[API Update] Failed to delete IP whitelist: %v", err)
		return nil, err
	}

	if len(req.IpWhitelist) > 0 {
		for _, ip := range req.IpWhitelist {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				g.Log().Errorf(ctx, "[API Templates Update] Empty IP address")
				continue
			}

			if net.ParseIP(ip) == nil {
				g.Log().Errorf(ctx, "[API Templates Update] Invalid IP address: %s", ip)
				continue
			}
			_, err = tx.Model("api_ip_whitelist").Insert(g.Map{
				"api_id":      req.ID,
				"ip":          strings.TrimSpace(ip),
				"create_time": now,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	res.SetSuccess(public.LangCtx(ctx, "Update API successfully"))
	return res, nil
}
