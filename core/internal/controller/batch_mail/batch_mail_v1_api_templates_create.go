package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/public"
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"net"
	"strings"
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

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// create API template
	result, err := tx.Model("api_templates").Insert(g.Map{
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
		"expire_time":          0,
		"last_key_update_time": time.Now().Unix(),
		//"ip_whitelist_enabled": req.IpWhitelistEnabled,
	})

	if err != nil {
		return nil, err
	}

	apiId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if len(req.IpWhitelist) > 0 {
		now := time.Now().Unix()
		for _, ip := range req.IpWhitelist {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}

			if net.ParseIP(ip) == nil {
				continue
			}

			_, err = tx.Model("api_ip_whitelist").Insert(g.Map{
				"api_id":      apiId,
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

	res.SetSuccess(public.LangCtx(ctx, "Create API successfully"))
	return res, nil
}
