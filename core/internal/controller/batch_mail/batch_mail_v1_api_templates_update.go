package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
	}

	if req.ResetKey {
		apiKey, err := generateApiKey()
		if err != nil {
			return nil, err
		}
		updateMap["api_key"] = apiKey
		updateMap["last_key_update_time"] = now
	}

	_, err = g.DB().Model("api_templates").
		Where("id", req.ID).
		Update(updateMap)

	if err != nil {
		return nil, err
	}

	res.SetSuccess(public.LangCtx(ctx, "Update API successfully"))
	return res, nil
}
