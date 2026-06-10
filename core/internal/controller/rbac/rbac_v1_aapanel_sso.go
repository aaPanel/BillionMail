package rbac

import (
	"billionmail-core/api/rbac/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model"
	"billionmail-core/internal/service/public"
	service "billionmail-core/internal/service/rbac"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const aaPanelSSOSkewSeconds = 300

func aaPanelSSOSecret() string {
	for _, key := range []string{"AAPANEL_SSO_SECRET", "BT_PANEL_SSO_SECRET"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
		if value := strings.TrimSpace(public.MustGetDockerEnv(key, "")); value != "" {
			return value
		}
	}
	return ""
}

func aaPanelSSOExpectedSignature(secret, timestamp, nonce, username string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(nonce))
	mac.Write([]byte("\n"))
	mac.Write([]byte(username))
	return hex.EncodeToString(mac.Sum(nil))
}

func validateAaPanelSSORequest(ctx context.Context, username string) error {
	r := g.RequestFromCtx(ctx)
	secret := aaPanelSSOSecret()
	if secret == "" {
		return gerror.New("aaPanel SSO is disabled")
	}

	timestamp := strings.TrimSpace(r.Header.Get("X-Aapanel-Timestamp"))
	nonce := strings.TrimSpace(r.Header.Get("X-Aapanel-Nonce"))
	signature := strings.TrimSpace(r.Header.Get("X-Aapanel-Signature"))
	if timestamp == "" || nonce == "" || signature == "" {
		return gerror.New("missing aaPanel SSO signature headers")
	}
	if len(nonce) > 128 {
		return gerror.New("invalid aaPanel SSO nonce")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return gerror.New("invalid aaPanel SSO timestamp")
	}
	delta := time.Now().Unix() - ts
	if delta < 0 {
		delta = -delta
	}
	if delta > aaPanelSSOSkewSeconds {
		return gerror.New("aaPanel SSO request expired")
	}

	cacheKey := "AAPANEL_SSO_NONCE:" + nonce
	if _, ok := public.GetCache(cacheKey).(int); ok {
		return gerror.New("aaPanel SSO nonce already used")
	}

	signature = strings.TrimPrefix(strings.ToLower(signature), "sha256=")
	expected := aaPanelSSOExpectedSignature(secret, timestamp, nonce, username)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return gerror.New("invalid aaPanel SSO signature")
	}

	public.SetCache(cacheKey, 1, aaPanelSSOSkewSeconds)
	return nil
}

// AaPanelSSO signs in a trusted aaPanel request without exposing BillionMail credentials.
func (c *ControllerV1) AaPanelSSO(ctx context.Context, req *v1.AaPanelSSOReq) (res *v1.AaPanelSSORes, err error) {
	res = &v1.AaPanelSSORes{}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		username = strings.TrimSpace(public.MustGetDockerEnv("ADMIN_USERNAME", ""))
	}
	if username == "" {
		return nil, gerror.New("username cannot be empty")
	}

	if err = validateAaPanelSSORequest(ctx, username); err != nil {
		return
	}

	var account model.Account
	if err = g.DB().Model("account").Where("username = ?", username).Scan(&account); err != nil {
		return
	}
	if account.AccountId == 0 {
		return nil, gerror.New("account does not exist")
	}

	roles, err := service.Account().GetAccountRoles(ctx, account.AccountId)
	if err != nil {
		return nil, gerror.New("failed to get account roles")
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.RoleName)
	}

	token, _, err := service.JWT().GenerateToken(account.AccountId, account.Username, roleNames)
	if err != nil {
		return nil, gerror.New("failed to generate token")
	}
	refreshToken, err := service.JWT().GenerateRefreshToken(account.AccountId, account.Username)
	if err != nil {
		return nil, gerror.New("failed to generate refresh token")
	}

	_, _ = g.DB().Model("account").Data(g.Map{
		"last_login_time": time.Now().Unix(),
	}).Where("account_id = ?", account.AccountId).Update()

	r := g.RequestFromCtx(ctx)
	_ = r.Session.Set("SignedToken", token)
	_ = r.Session.Set("safe_path_pass", true)

	res.Success = true
	res.Code = 0
	res.Msg = "aaPanel SSO successful"
	res.Data.Token = token
	res.Data.RefreshToken = refreshToken
	res.Data.TTL = gconv.Int64(service.JWT().AccessExpiry.Seconds())
	res.Data.AccountInfo.Id = account.AccountId
	res.Data.AccountInfo.Username = account.Username
	res.Data.AccountInfo.Email = account.Email
	res.Data.AccountInfo.Status = account.Status
	res.Data.AccountInfo.Lang = account.Language

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Login,
		Log:  fmt.Sprintf("The user:%s login by aaPanel SSO was successful", account.Username),
		Data: res.Data,
	})
	return
}
