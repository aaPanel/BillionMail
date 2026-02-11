package domains

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
	"context"

	"billionmail-core/api/domains/v1"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) FreshDNSRecords(ctx context.Context, req *v1.FreshDNSRecordsReq) (res *v1.FreshDNSRecordsRes, err error) {
	res = &v1.FreshDNSRecordsRes{}

	domains.FreshRecords(ctx, req.Domain)

	// Look up dedicated IP for this domain
	dedicatedIP := ""
	val, _ := g.DB().Model("bm_multi_ip_domain").
		Where("domain = ? AND active = 1", req.Domain).
		Value("outbound_ip")
	if val != nil {
		dedicatedIP = val.String()
	}

	res.Data = domains.GetRecordsInCache(req.Domain, dedicatedIP)

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Domain,
		Log:  "Fresh DNS records for domain :" + req.Domain + " successfully",
		Data: res.Data,
	})

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}
