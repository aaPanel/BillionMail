package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/public"
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

func generateApiKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (c *ControllerV1) ApiTemplatesCreate(ctx context.Context, req *v1.ApiTemplatesCreateReq) (res *v1.ApiTemplatesCreateRes, err error) {
	res = &v1.ApiTemplatesCreateRes{}

	// check if template exists
	count, err := g.DB().Model("email_templates").Where("id", req.TemplateId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New(public.LangCtx(ctx, "Email template does not exist"))
	}

	// generate API key
	apiKey, err := generateApiKey()
	if err != nil {
		return nil, err
	}

	// create API template
	_, err = g.DB().Model("api_templates").Insert(g.Map{
		"api_key":              apiKey,
		"api_name":             req.ApiName,
		"template_id":          req.TemplateId,
		"subject":              req.Subject,
		"addresser":            req.Addresser,
		"full_name":            req.FullName,
		"unsubscribe":          req.Unsubscribe,
		"track_open":           1,
		"track_click":          1,
		"active":               req.Active,
		"expire_time":          req.ExpireTime, // 0 is a permanently valid unit of seconds
		"last_key_update_time": time.Now().Unix(),
	})

	if err != nil {
		return nil, err
	}

	res.SetSuccess(public.LangCtx(ctx, "Create API successfully"))
	return res, nil
}
