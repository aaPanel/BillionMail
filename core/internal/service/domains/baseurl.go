package domains

import (
	v1 "billionmail-core/api/domains/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
	"sync"
)

var (
	baseurl      = ""
	baseurlMap   = make(map[string]string)
	baseurlMapMu sync.RWMutex
)

// GetBaseURL get baseurl of console panel
func GetBaseURL() string {
	// Prioritize the verification of the reverse proxy domain name
	var reverseProxyDomain string
	err := public.OptionsMgrInstance.GetOption(context.Background(), "reverse_proxy_domain", &reverseProxyDomain)
	if err == nil && reverseProxyDomain != "" {
		return reverseProxyDomain
	}
	return baseurl
}

// GetBaseURLBySender get baseurl by sender email address
func GetBaseURLBySender(sender string) string {
	err := g.Validator().Data(sender).Rules("email").Run(context.Background())

	if err != nil {
		g.Log().Debug(context.Background(), "GetBaseURLBySender --> Invalid email address", sender, err)
		return GetBaseURL()
	}

	baseurlMapMu.RLock()
	s, ok := baseurlMap[strings.SplitN(sender, "@", 2)[1]]
	baseurlMapMu.RUnlock()
	if ok {
		return s
	}

	return GetBaseURL()
}

func UpdateBaseURL(ctx context.Context, domain ...string) {
	g.Log().Debug(context.Background(), "UpdateBaseURL --> Starting")
	defer func() {
		g.Log().Debug(context.Background(), "UpdateBaseURL --> Ending")
	}()

	var domains []string

	if len(domain) > 0 {
		domains = domain
	} else {
		ds, err := All(ctx)

		if err != nil {
			return
		}

		for _, d := range ds {
			domains = append(domains, d.Domain)
		}
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		baseurl = buildBaseURL("")
		g.Log().Debug(ctx, "UpdateBaseURL --> Base URL updated:", baseurl)
	}()

	for _, d := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			url := buildBaseURL(domain)
			baseurlMapMu.Lock()
			baseurlMap[domain] = url
			baseurlMapMu.Unlock()
			g.Log().Debug(ctx, "UpdateBaseURL --> Updating base URL for domain:", domain, " URL:", url)
		}(d)
	}

	wg.Wait()
}

func buildBaseURL(hostname string) (s string) {
	// If reverse proxy domain is configured, use it directly (no internal port)
	var reverseProxyDomain string
	if err := public.OptionsMgrInstance.GetOption(context.Background(), "reverse_proxy_domain", &reverseProxyDomain); err == nil && reverseProxyDomain != "" {
		return reverseProxyDomain
	}

	scheme := "https"
	serverPort := g.Server(consts.DEFAULT_SERVER_NAME).GetListenedHTTPSPort()

	if serverPort == -1 {
		scheme = "http"
		serverPort = g.Server(consts.DEFAULT_SERVER_NAME).GetListenedPort()
	}

	withPort := true

	if serverPort == -1 || serverPort == 80 || serverPort == 443 {
		withPort = false
	}

	serverIP, localIP, err := public.GetServerIPAndLocalIP()

	if err != nil {
		g.Log().Error(context.Background(), "failed to get server IP and local IP", err)
		return
	}

	if appEnv, err := public.DockerEnv("APP_ENV"); err == nil && appEnv != "" {
		switch appEnv {
		case "development":
			s = scheme + "://" + localIP
			if withPort {
				s += ":" + gconv.String(serverPort)
			}
			return
		}
	}

	if hostname == "" {
		hostname, err = public.DockerEnv("BILLIONMAIL_HOSTNAME")
		if hostname != "" && hostname != "mail.example.com" {
			s = scheme + "://" + hostname
		} else {
			s = scheme + "://" + serverIP
		}

		if withPort {
			s += ":" + gconv.String(serverPort)
		}

		return
	} else {
		hostname = public.FormatMX(hostname)
		v1DNSRecord := v1.DNSRecord{
			Type:  "A",
			Host:  hostname,
			Value: serverIP,
			Valid: true,
		}

		if ValidateARecord(v1DNSRecord) {
			s = scheme + "://" + hostname
			if withPort {
				s += ":" + gconv.String(serverPort)
			}
			return
		}
	}

	s = GetBaseURL()

	return
}
