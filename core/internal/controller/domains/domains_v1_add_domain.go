package domains

import (
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/rbac"
	"context"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/v2/frame/g"
	"time"

	"billionmail-core/api/domains/v1"
)

func (c *ControllerV1) AddDomain(ctx context.Context, req *v1.AddDomainReq) (res *v1.AddDomainRes, err error) {
	res = &v1.AddDomainRes{}

	domain := &v1.Domain{
		Domain:       req.Domain,
		ARecord:      "mail." + req.Domain,
		Mailboxes:    req.Mailboxes,
		MailboxQuota: int64(req.MailboxQuota),
		Quota:        int64(req.Quota),
		RateLimit:    req.RateLimit,
	}

	if err = domains.Add(ctx, domain); err != nil {
		return nil, err
	}

	// Automatic Certificate Application
	gtimer.AddOnce(500*time.Millisecond, func() {
		acc, accErr := rbac.GetCurrentAccount(ctx)
		if accErr == nil {
			certErr := domains.ApplyLetsEncryptCertWithHttp(ctx, req.Domain, acc)
			if certErr != nil {
				g.Log().Warningf(ctx, "Domain name [%s] added successfully, but auto-request certificate failed: %v", req.Domain, certErr)
			}
		} else {
			g.Log().Warningf(ctx, "The domain name [%s] was added successfully, but the current account failed to obtain the certificate automatically: %v", req.Domain, accErr)
		}

	})

	res.SetSuccess("Domain added successfully")
	return
}
