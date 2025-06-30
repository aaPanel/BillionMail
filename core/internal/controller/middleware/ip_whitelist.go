package middleware

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"strings"
)

// Paths excluded from IP whitelist verification
var ExcludedPaths = map[string]struct{}{
	"/favicon.ico":                {},
	"/robots.txt":                 {},
	"/unsubscribe.html":           {},
	"/api/unsubscribe/user_group": {},
	"/api/unsubscribe":            {},
	"/api/batch_mail/api/send":    {},

	// other
}

func isExcludedPath(path string) bool {
	if _, ok := ExcludedPaths[path]; ok {
		g.Log().Debugf(context.Background(), "[IP Whitelist 1] Excluded path: %s allowed", path)
		return true
	}
	// Prefix match (static resources, acme challenge, etc.)
	if strings.HasPrefix(path, "/.well-known/acme-challenge") ||
		strings.HasPrefix(path, "/pmta/") ||
		strings.HasPrefix(path, "/roundcube") {
		g.Log().Warning(context.Background(), "[IP Whitelist 2] Excluded path: ", path, " allowed")
		return true
	}

	return false
}

// Allowed IP whitelist
func getAllowedIPs() []string {
	var ips []string
	result, err := g.DB().Model("bm_console_ip_whitelist").Fields("ip").All()
	if err != nil {
		g.Log().Error(context.Background(), "[IP Whitelist] Failed to get IP whitelist:", err)
		return []string{"127.0.0.1", "::1"}
	}

	for _, ip := range result {
		ips = append(ips, ip["ip"].String())
	}

	ips = append(ips, "127.0.0.1")
	ips = append(ips, "::1")
	// Add server IP
	ipServer, _ := public.GetServerIP()
	if ipServer != "" {
		ips = append(ips, ipServer)
	}

	return ips
}

func appendIPWhitelistEnvIfNotExists() {
	envFile := public.AbsPath("../.env")
	const comment = "# Console ip whitelist   Disabled:false  Enabled:true"
	const config = "IP_WHITELIST_ENABLE=false"

	content := gfile.GetContents(envFile)

	if !strings.Contains(content, "IP_WHITELIST_ENABLE") {
		appendStr := "\n" + comment + "\n" + config + "\n"
		if err := gfile.PutContentsAppend(envFile, appendStr); err != nil {
			g.Log().Errorf(context.Background(), "[IP Whitelist] Failed to append to .env file: %v", err)
		}
	}
}

// Whether to enable the IP whitelist
func isIPWhitelistEnabled(envMap map[string]string) bool {

	isIPWhite, exists := envMap["IP_WHITELIST_ENABLE"]

	if !exists {
		appendIPWhitelistEnvIfNotExists()
		return false
	}

	return isIPWhite == "1" || strings.ToLower(isIPWhite) == "true"
}

func IPWhitelist(r *ghttp.Request) {

	envMap, err := public.LoadEnvFile()
	if err != nil {
		g.Log().Errorf(r.Context(), "[IP Whitelist] Failed to load environment variables: %v", err)
		r.Middleware.Next()
		return
	}
	// 1. If the whitelist feature is not enabled, allow access directly
	if !isIPWhitelistEnabled(envMap) {
		r.Middleware.Next()
		// g.Log().Warningf(r.Context(), "[IP Whitelist] Not enabled, allowing access directly")
		return
	}
	// 2. Check if the path is in the exclusion list
	g.Log().Debugf(r.Context(), "[IP Whitelist] Checking path: %s", r.URL.Path)
	if isExcludedPath(r.URL.Path) {
		g.Log().Debugf(r.Context(), "[IP Whitelist] Excluded path allowed: %s", r.URL.Path)
		r.Middleware.Next()
		return
	}

	// 3. Check if the client IP is in the whitelist
	clientIP := r.GetClientIp()
	if isIPAllowed(clientIP) {
		g.Log().Warningf(r.Context(), "[IP Whitelist] Access allowed: %s", clientIP)
		r.Middleware.Next()
		return
	}

	g.Log().Warningf(r.Context(), "[IP Whitelist] Blocked non-whitelisted IP: %s, UA: %s, Path: %s", clientIP, r.Request.UserAgent(), r.Request.URL.Path)
	r.Response.WriteStatus(403, "IP not allowed")
	r.Exit()
}

func isIPAllowed(clientIP string) bool {
	allowedIPs := getAllowedIPs()
	for _, ip := range allowedIPs {
		if strings.TrimSpace(ip) == clientIP {
			return true
		}
	}
	return false
}
