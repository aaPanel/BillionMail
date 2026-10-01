package relay

import (
	v1 "billionmail-core/api/relay/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
	"billionmail-core/internal/service/relay"
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) UpdateRelayConfig(ctx context.Context, req *v1.UpdateRelayConfigReq) (res *v1.UpdateRelayConfigRes, err error) {
	res = &v1.UpdateRelayConfigRes{}

	relayInfo, err := g.DB().Model("bm_relay_config").Where("id", req.ID).One()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check relay configuration: {}", err.Error())))
		return res, nil
	}
	if relayInfo.IsEmpty() {
		res.SetError(gerror.New(public.LangCtx(ctx, "Relay configuration does not exist")))
		return res, nil
	}

	updateTime := int(time.Now().Unix())
	updateData := g.Map{
		"update_time": updateTime,
	}

	if req.Remark != "" {
		updateData["remark"] = req.Remark
	}
	if req.Rtype != "" {
		updateData["rtype"] = req.Rtype
	}
	if req.RelayHost != "" {
		updateData["relay_host"] = req.RelayHost
	}
	if req.RelayPort != "" {
		updateData["relay_port"] = req.RelayPort
	}

	// Handle authentication updates - support clearing auth info
	authUserProvided := false
	authPasswordProvided := false

	// Check if AuthUser is explicitly provided (even if empty)
	if req.AuthUser != "" {
		updateData["auth_user"] = req.AuthUser
		authUserProvided = true
	} else {
		// If AuthUser is explicitly set to empty string, clear authentication
		updateData["auth_user"] = ""
		updateData["auth_password"] = ""
		authUserProvided = true
	}

	if req.AuthPassword != "" {
		encryptedPwd, err := relay.EncryptPassword(ctx, req.AuthPassword)
		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to encrypt password: {}", err.Error())))
			return res, nil
		}
		updateData["auth_password"] = encryptedPwd
		authPasswordProvided = true
	}

	// Update AuthMethod based on authentication status
	if req.AuthMethod != "" {
		updateData["auth_method"] = req.AuthMethod
	} else if authUserProvided {
		// Auto-set AuthMethod based on whether authentication is being used
		if updateData["auth_user"].(string) == "" {
			updateData["auth_method"] = "NONE"
		} else if !authPasswordProvided {
			// If user is provided but password is not updated, keep existing auth method or default to LOGIN
			if req.AuthMethod == "" {
				updateData["auth_method"] = "LOGIN"
			}
		}
	}

	if req.IP != "" {
		updateData["ip"] = req.IP
	}
	if req.Host != "" {
		updateData["host"] = req.Host
	}
	updateData["active"] = req.Active
	if req.AuthMethod != "" {
		updateData["auth_method"] = req.AuthMethod
	}
	if req.SkipTlsVerify > 0 {
		updateData["skip_tls_verify"] = req.SkipTlsVerify
	}
	if req.HeloName != "" {
		updateData["helo_name"] = req.HeloName
	}
	if req.SmtpName != "" {
		updateData["smtp_name"] = req.SmtpName
	}

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to start transaction: {}", err.Error())))
		return res, nil
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if len(updateData) > 1 {
		_, err = tx.Model("bm_relay_config").Where("id", req.ID).Data(updateData).Update()
		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to update relay configuration: {}", err.Error())))
			return res, nil
		}
	}

	// Only update sender domain mappings when explicitly provided.
	if len(req.SenderDomains) > 0 {

		_, err = tx.Model("bm_relay_domain_mapping").
			Where("relay_id", req.ID).
			Delete()
		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to delete sender domains: {}", err.Error())))
			return res, nil
		}

		var domains_ []string
		for _, domain := range req.SenderDomains {

			domains_ = append(domains_, domain)
		}

		existingDomainsCount, err := tx.Model("bm_relay_domain_mapping").
			WhereIn("sender_domain", domains_).
			WhereNot("relay_id", req.ID).
			Count()
		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check domain mapping: {}", err.Error())))
			return res, nil
		}
		if existingDomainsCount > 0 {
			res.SetError(gerror.New(public.LangCtx(ctx, "One or more sender domains are already used by another relay configuration")))
			return res, nil
		}

		var currentDomains []string
		err = tx.Model("bm_relay_domain_mapping").
			Where("relay_id", req.ID).
			Fields("sender_domain").
			Scan(&currentDomains)
		if err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get current domain mappings: {}", err.Error())))
			return res, nil
		}

		currentDomainsMap := make(map[string]bool)
		for _, domain := range currentDomains {
			currentDomainsMap[domain] = true
		}

		newDomainsMap := make(map[string]bool)
		for _, domain := range domains_ {
			newDomainsMap[domain] = true
		}

		for _, oldDomain := range currentDomains {
			if !newDomainsMap[oldDomain] {
				_, err = tx.Model("bm_relay_domain_mapping").
					Where("relay_id", req.ID).
					Where("sender_domain", oldDomain).
					Delete()
				if err != nil {
					res.SetError(gerror.New(public.LangCtx(ctx, "Failed to delete obsolete domain mapping: {}", err.Error())))
					return res, nil
				}
			}
		}

		for _, newDomain := range domains_ {
			if !currentDomainsMap[newDomain] {
				_, err = tx.Model("bm_relay_domain_mapping").
					Data(g.Map{
						"relay_id":      req.ID,
						"sender_domain": newDomain,
						"create_time":   updateTime,
					}).
					Insert()
				if err != nil {
					res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create new domain mapping: {}", err.Error())))
					return res, nil
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to commit transaction: {}", err.Error())))
		return res, nil
	}

	if err = relay.SyncRelayConfigsToPostfix(ctx); err != nil {
		g.Log().Error(ctx, "Failed to sync relay configs to Postfix:", err)
	}

	var firstDomain string
	domainResult, err := g.DB().Model("bm_relay_domain_mapping").
		Where("relay_id", req.ID).
		Order("create_time ASC").
		Limit(1).
		Value("sender_domain")

	if err == nil && domainResult != nil && domainResult.String() != "" {
		firstDomain = domainResult.String()
		if strings.HasPrefix(firstDomain, "@") {
			firstDomain = firstDomain[1:]
		}

		ip := relayInfo["ip"].String()
		if req.IP != "" {
			ip = req.IP
		}
		host := relayInfo["host"].String()
		if req.Host != "" {
			host = req.Host
		}

		record, _ := domains.GetSPFRecord(firstDomain, false)
		if record.Value != "" {
			res.SPFRecord = v1.DNSRecord{
				Type:  record.Type,
				Host:  record.Host,
				Value: record.Value,
			}
		} else {

			spfValue := relay.GenerateSPFRecord(ip, host, firstDomain)
			res.SPFRecord = v1.DNSRecord{
				Type:  "TXT",
				Host:  firstDomain,
				Value: spfValue,
			}
		}
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.SMTP,
		Log:  "Updated relay configuration : " + req.RelayHost,
	})

	res.SetSuccess(public.LangCtx(ctx, "Relay configuration updated successfully"))
	return res, nil
}
