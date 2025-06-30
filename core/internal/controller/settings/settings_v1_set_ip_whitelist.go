package settings

import (
	"billionmail-core/internal/service/public"
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/frame/g"
	"net"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/settings/v1"
)

func (c *ControllerV1) SetIPWhitelist(ctx context.Context, req *v1.SetIPWhitelistReq) (res *v1.SetIPWhitelistRes, err error) {
	res = &v1.SetIPWhitelistRes{}

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 获取当前所有IP
	var currentIPs []string
	err = tx.Model("bm_console_ip_whitelist").Fields("ip").Scan(&currentIPs)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 找出需要删除的IP（在当前列表中但不在新列表中）
	ipsToDelete := make([]string, 0)
	for _, currentIP := range currentIPs {
		found := false
		for _, newIP := range req.IPList {
			if strings.TrimSpace(newIP) == currentIP {
				found = true
				break
			}
		}
		if !found {
			ipsToDelete = append(ipsToDelete, currentIP)
		}
	}

	_, err = tx.Model("bm_console_ip_whitelist").WhereIn("ip", ipsToDelete).Delete()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	for _, ip := range req.IPList {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if net.ParseIP(ip) == nil {
			return nil, gerror.Newf("Invalid IP address: %s", ip)
		}

		ipType := 1 // IPv4
		if net.ParseIP(ip).To4() == nil {
			ipType = 2 // IPv6
		}

		_, err = tx.Model("bm_console_ip_whitelist").Insert(g.Map{
			"ip":          ip,
			"ip_type":     ipType,
			"create_time": now,
		})
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	res.SetSuccess(public.LangCtx(ctx, "IP whitelist updated successfully"))

	return res, nil
}
