package mail_boxes

import (
	domainsV1 "billionmail-core/api/domains/v1"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/mail_boxes"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"

	"billionmail-core/api/mail_boxes/v1"
)

func (c *ControllerV1) GetMailbox(ctx context.Context, req *v1.GetMailboxReq) (res *v1.GetMailboxRes, err error) {
	res = &v1.GetMailboxRes{}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	mailboxList, total, err := mail_boxes.Get(ctx, req.Domain, req.Keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	// transform mailbox list to include MX records
	mailboxListWithMx := make([]v1.MailboxWithMxRecord, len(mailboxList))
	for i, mailbox := range mailboxList {
		mailboxListWithMx[i] = v1.MailboxWithMxRecord{
			Mailbox:  mailbox,
			MxRecord: GenerateSPFRecord(mailbox.Domain), // 使用 GenerateSPFRecord 获取 MX 记录
		}
		// Handle sensitive information
		mailboxListWithMx[i].Password, _ = mail_boxes.PasswdDecode(ctx, mailboxListWithMx[i].PasswordEncode)
	}

	res.Data.Total = total
	res.Data.List = mailboxListWithMx
	res.SetSuccess("Success")

	return
}

func buildCacheKey(domain, recordType string) string {
	return fmt.Sprintf("DOMAIN_DNS_RECORDS_:%s:_%s", domain, recordType)
}

// GenerateSPFRecord generates SPF record hints based on IP and host.
func GenerateSPFRecord(domain string) string {
	// Retrieve the current SPF record for the domain
	var existingValue string

	// Fetch SPF record from cache or domains service
	spfRecord := public.GetCache(buildCacheKey(domain, "MX"))

	if spfRecord != nil {
		if v, ok := spfRecord.(domainsV1.DNSRecord); ok {
			existingValue = v.Value
		}
	}
	if existingValue == "" {
		record, _ := domains.GetMXRecord(domain, false)
		existingValue = record.Value
	}

	return existingValue
}
