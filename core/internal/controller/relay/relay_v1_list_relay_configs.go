package relay

import (
	"billionmail-core/api/relay/v1"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/public"
	relay_service "billionmail-core/internal/service/relay"
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) ListRelayConfigs(ctx context.Context, req *v1.ListRelayConfigsReq) (res *v1.ListRelayConfigsRes, err error) {
	res = &v1.ListRelayConfigsRes{}

	model := g.DB().Model("bm_relay").Safe()

	if req.Rtype != "" {
		model = model.Where("rtype", req.Rtype)
	}

	// Search functionality (supports searching by remarks, relay server address, sender domain)
	if req.Keyword != "" {
		searchKey := "%" + req.Keyword + "%"
		model = model.WhereOr(
			"remark LIKE ?", searchKey,
		).WhereOr(
			"relay_host LIKE ?", searchKey,
		).WhereOr(
			"sender_domain LIKE ?", searchKey,
		).WhereOr(
			"smtp_name LIKE ?", searchKey,
		).WhereOr(
			"auth_user LIKE ?", searchKey,
		).WhereOr(
			"rtype LIKE ?", searchKey,
		)
	}

	var list []*entity.BmRelay

	err = model.Order("create_time DESC").Scan(&list)
	if err != nil && err != sql.ErrNoRows {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to retrieve relay configuration list: {}", err.Error())))
		return res, nil
	}

	resultList := make([]*v1.BmRelayWithSPF, 0, len(list))
	for _, item := range list {
		// Generate SPF record suggestion
		spfRecord := GenerateSPFRecord(item.Ip, item.Host, item.SenderDomain)

		bmRelay := toPostfixRelayResponse(ctx, item)

		smtpStatus := relay_service.CheckSmtpConnection(ctx, item.RelayHost, item.RelayPort, item.AuthUser, "")

		relayWithSPF := &v1.BmRelayWithSPF{
			BmRelay: bmRelay,
			SPFRecord: v1.DNSRecord{
				Type:  "TXT",
				Host:  "@",
				Value: spfRecord,
			},
			SmtpStatus: v1.SmtpStatus{
				Status: smtpStatus.Status,
				Msg:    smtpStatus.Msg,
			},
		}

		resultList = append(resultList, relayWithSPF)
	}

	res.Data.Total = len(list)
	res.Data.List = resultList
	res.SetSuccess(public.LangCtx(ctx, "Relay configuration list retrieved successfully"))
	return res, nil
}

// toPostfixRelayResponse Converts database entity to API response structure
func toPostfixRelayResponse(ctx context.Context, relay *entity.BmRelay) *v1.BmRelay {
	password, _ := DecryptPassword(ctx, relay.AuthPassword)
	if password == "" {
		password = "********"
	}

	return &v1.BmRelay{
		Id:             relay.Id,
		Remark:         relay.Remark,
		Rtype:          relay.Rtype,
		SenderDomain:   relay.SenderDomain,
		RelayHost:      relay.RelayHost,
		RelayPort:      relay.RelayPort,
		AuthUser:       relay.AuthUser,
		AuthPassword:   password,
		Ip:             relay.Ip,
		Host:           relay.Host,
		Active:         relay.Active,
		CreateTime:     relay.CreateTime,
		UpdateTime:     relay.UpdateTime,
		AuthMethod:     relay.AuthMethod,
		TlsProtocol:    relay.TlsProtocol,
		SkipTlsVerify:  relay.SkipTlsVerify,
		HeloName:       relay.HeloName,
		SmtpName:       relay.SmtpName,
		HeaderJson:     relay.HeaderJson,
		MaxConcurrency: relay.MaxConcurrency,
		//MaxRetries:     relay.MaxRetries,
		//MaxIdleTime:    relay.MaxIdleTime,
		//MaxWaitTime:    relay.MaxWaitTime,
	}
}
