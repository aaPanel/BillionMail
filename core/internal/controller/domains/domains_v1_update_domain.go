package domains

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/domains/v1"
	"billionmail-core/internal/service/domains"
)

func (c *ControllerV1) UpdateDomain(ctx context.Context, req *v1.UpdateDomainReq) (res *v1.UpdateDomainRes, err error) {
	res = &v1.UpdateDomainRes{}

	if req.Domain == "" {
		return nil, gerror.New("Domain cannot be empty")
	}

	domain := &v1.Domain{
		Domain:       req.Domain,
		ARecord:      "mail." + req.Domain,
		Mailboxes:    req.Mailboxes,
		MailboxQuota: int64(req.MailboxQuota),
		Quota:        int64(req.Quota),
		RateLimit:    req.RateLimit,
		Active:       req.Active,
		Catchall:     req.Catchall,
	}

	if err = domains.Update(ctx, domain); err != nil {
		return nil, err
	}

	res.SetSuccess("Domain updated successfully")
	return
}
