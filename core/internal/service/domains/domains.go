package domains

import (
	v2 "billionmail-core/api/dockerapi/v1"
	v1 "billionmail-core/api/domains/v1"
	mail_v1 "billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/mail_boxes"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/multi_ip_domain"
	"billionmail-core/internal/service/public"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gregex"
)

var (
	mutex sync.Mutex
)

// convertBlacklistDetails 将 model.BlacklistDetail 转换为 v1.BlacklistDetail
func convertBlacklistDetails(details []model.BlacklistDetail) []v1.BlacklistDetail {
	result := make([]v1.BlacklistDetail, len(details))
	for i, detail := range details {
		result[i] = v1.BlacklistDetail{
			Blacklist: detail.Blacklist,
			Response:  detail.Response,
			Time:      detail.Time,
		}
	}
	return result
}

func setCatchall(ctx context.Context, domainName, catchall string) error {
	address := fmt.Sprintf("@%s", domainName)
	if catchall != "" {
		var count int
		count, err := g.DB().Model("alias").
			Where("address", address).
			Where("domain", domainName).
			Count()
		if err != nil {
			return fmt.Errorf("failed to check alias existence: %v", err)
		}
		if count > 0 {
			_, err = g.DB().Model("alias").
				Where("address", address).
				Where("domain", domainName).
				Data(g.Map{"goto": catchall, "active": 1, "update_time": time.Now().Unix()}).
				Update()
			if err != nil {
				return fmt.Errorf("failed to update alias: %v", err)
			}
		} else {
			_, err = g.DB().Model("alias").Data(g.Map{
				"address":     address,
				"goto":        catchall,
				"domain":      domainName,
				"active":      1,
				"create_time": time.Now().Unix(),
				"update_time": time.Now().Unix(),
			}).Insert()
			if err != nil {
				return fmt.Errorf("failed to insert alias: %v", err)
			}
		}
	} else {
		// catchall is empty, disabled
		_, _ = g.DB().Model("alias").
			Where("address", address).
			Where("domain", domainName).
			Data(g.Map{"active": 0, "update_time": time.Now().Unix()}).
			Update()
	}
	return nil
}

func Add(ctx context.Context, domain *v1.Domain) error {
	domain.CreateTime = time.Now().Unix()
	domain.Active = 1

	_, err := g.DB().Model("domain").Ctx(ctx).Insert(domain)

	if err == nil {
		// Create mailbox abuse, postmaster, admin, noreply, and support for the new domain
		mailboxes := []string{"abuse", "postmaster", "admin", "noreply", "support"}

		for _, mailbox := range mailboxes {
			_ = mail_boxes.Add(ctx, &mail_v1.Mailbox{
				Username:  mailbox + "@" + domain.Domain,
				Password:  grand.S(16),
				FullName:  mailbox,
				IsAdmin:   0,
				Quota:     5242880,
				LocalPart: mailbox,
				Domain:    domain.Domain,
				Active:    1,
			})
		}

		// attempt update hostname in .env file
		hostname := public.MustGetDockerEnv("BILLIONMAIL_HOSTNAME", "")

		if hostname == "" || hostname == "mail.example.com" {
			err = public.SetDockerEnv("BILLIONMAIL_HOSTNAME", public.FormatMX(domain.Domain))

			if err != nil {
				return fmt.Errorf("failed to update BILLIONMAIL_HOSTNAME in .env file: %v", err)
			}

			// update postfix environment parameter
			_, err = public.DockerApiFromCtx(ctx).ExecCommandByName(ctx, consts.SERVICES.Postfix, []string{"bash", "-c", fmt.Sprintf("sed -i '/^BILLIONMAIL_HOSTNAME=/d' /postfix.sh && sed -i '/^#!\\/bin\\/bash/a BILLIONMAIL_HOSTNAME=%s' /postfix.sh", public.FormatMX(domain.Domain))}, "root")

			if err != nil {
				return fmt.Errorf("failed to update BILLIONMAIL_HOSTNAME in postfix container: %v", err)
			}

			// restart postfix service
			err = public.DockerApiFromCtx(ctx).RestartContainerByName(ctx, consts.SERVICES.Postfix)

			if err != nil {
				return fmt.Errorf("failed to restart postfix container: %v", err)
			}
		}

		// catchall
		if domain.Catchall != "" {
			err = setCatchall(ctx, domain.Domain, domain.Catchall)
		}
		return err
	}

	return err
}

func Update(ctx context.Context, updateData map[string]interface{}) error {

	domainName, _ := updateData["domain"].(string)

	_, err := g.DB().Model("domain").
		Ctx(ctx).
		Where("domain", domainName).
		Update(updateData)

	if err != nil {
		return err
	}

	// catchall
	if catchall, ok := updateData["catchall"].(string); ok && catchall != "" {
		err = setCatchall(ctx, domainName, catchall)
		if err != nil {
			return err
		}
	}

	return nil
}

func Delete(ctx context.Context, domainName string) error {
	_, err := g.DB().Model("domain").
		Ctx(ctx).
		Where("domain", domainName).
		Delete()

	if err == nil {
		// remove associated mailboxes
		_, err = g.DB().Model("mailbox").
			Ctx(ctx).
			Where("domain", domainName).
			Delete()

		// remove associated alias
		_, err = g.DB().Model("alias").
			Ctx(ctx).
			Where("domain", domainName).
			Delete()

	}

	return err
}

func Get(ctx context.Context, keyword string, page, pageSize int) ([]v1.Domain, int, error) {
	m := g.DB().Model("domain").Order("create_time", "desc")

	if keyword != "" {
		m = m.WhereLike("domain", fmt.Sprintf("%%%s%%", keyword))
	}

	count, err := m.Count()
	if err != nil {
		return nil, 0, err
	}

	var domains []v1.Domain
	err = m.Page(page, pageSize).Scan(&domains)

	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	var defaultDomain string
	val, err := g.DB().Model("bm_options").Where("name", "default_sender_domain").Value("value")
	if err == nil && val != nil {
		defaultDomain = val.String()
	}
	MultiIPDomainConfigs, err := multi_ip_domain.MultiIPDomainServiceInstance.GetConfigs(ctx)

	// 创建域名到专属IP配置的映射
	multiIPMap := make(map[string][]v1.MultiIPDomain)
	for _, config := range MultiIPDomainConfigs {
		multiIPMap[config.Domain] = append(multiIPMap[config.Domain], v1.MultiIPDomain{
			ID:             config.ID,
			Domain:         config.Domain,
			OutboundIP:     config.OutboundIP,
			NetworkName:    config.NetworkName,
			Subnet:         config.Subnet,
			PostfixIP:      config.PostfixIP,
			Aliases:        config.Aliases,
			SMTPServerName: config.SMTPServerName,
			Active:         config.Active,
			CreateTime:     config.CreateTime,
			UpdateTime:     config.UpdateTime,
			Status:         config.Status,
		})
	}

	crt := mail_service.NewCertificate()

	defer crt.Close()

	wg := sync.WaitGroup{}

	for i, domain := range domains {
		if domain.Domain == defaultDomain {
			domains[i].Default = 1
		} else {
			domains[i].Default = 0
		}
		// 补充专属ip信息
		if multiIPConfigs, exists := multiIPMap[domain.Domain]; exists && len(multiIPMap) > 0 {
			domains[i].MultiIPDomains = &multiIPConfigs[0]
		} else {
			domains[i].MultiIPDomains = nil
		}

		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()
			dedicatedIP := ""
			if domain.MultiIPDomains != nil {
				dedicatedIP = domain.MultiIPDomains.OutboundIP
			}
			domains[i].DNSRecords = GetRecordsInCache(strings.TrimPrefix(domain.Domain, "mail."), dedicatedIP)
		}(i, domain)

		// Retrieve Domains SSL certificate information
		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()
			domains[i].CertInfo, err = crt.GetSSLInfo(public.FormatMX(domain.Domain))

		}(i, domain)

		// catchall
		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()
			type Alias struct {
				Address    string `json:"address"     dc:"Alias address"`
				Goto       string `json:"goto"        dc:"Forwarding target"`
				Domain     string `json:"domain"      dc:"Domain name"`
				CreateTime int64  `json:"create_time" dc:"Creation time"`
				UpdateTime int64  `json:"update_time" dc:"Update time"`
				Active     int    `json:"active"      dc:"Status: 1-enabled, 0-disabled"`
			}

			address := fmt.Sprintf("@%s", domain.Domain)
			var alias Alias
			err := g.DB().Model("alias").
				Where("address = ? AND domain = ?", address, domain.Domain).
				Scan(&alias)

			if err != nil {
				domains[i].Catchall = ""

			} else {
				if alias.Active == 1 {
					domains[i].Catchall = alias.Goto
				} else {
					domains[i].Catchall = ""
				}
			}

		}(i, domain)

		// Blacklist check
		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()

			blacklistResult := GetBlacklistResult(domain.ARecord)
			if blacklistResult != nil {
				domains[i].BlackCheckResult = &v1.BlacklistCheckResult{
					Time:        blacklistResult.Time,
					Results:     blacklistResult.Domain,
					IP:          blacklistResult.IP,
					Tested:      blacklistResult.Tested,
					Passed:      blacklistResult.Passed,
					Invalid:     blacklistResult.Invalid,
					Blacklisted: blacklistResult.Blacklisted,
					BlackList:   convertBlacklistDetails(blacklistResult.BlackList),
				}
			} else {

				domains[i].BlackCheckResult = &v1.BlacklistCheckResult{
					Time:        0,
					Results:     "",
					IP:          "",
					Tested:      0,
					Passed:      0,
					Invalid:     0,
					Blacklisted: 0,
					BlackList:   []v1.BlacklistDetail{},
				}
			}

			domains[i].BlackCheckLog = GetBlacklistLogPath(domain.ARecord)
		}(i, domain)
	}

	wg.Wait()

	return domains, count, err
}

func All(ctx context.Context) ([]v1.Domain, error) {
	m := g.DB().Model("domain")

	var domains []v1.Domain
	err := m.Scan(&domains)

	if err != nil {
		return nil, err
	}

	return domains, err
}

// GetRelayDomains returns a set of domains that have active SMTP relay mappings.
// These domains should be excluded from local DKIM signing since the relay provider
// (e.g. SES, SendGrid) adds its own DKIM signature.
func GetRelayDomains(ctx context.Context) (map[string]bool, error) {
	type mapping struct {
		SenderDomain string `json:"sender_domain"`
	}
	var mappings []mapping
	err := g.DB().Model("bm_relay_domain_mapping rdm").
		LeftJoin("bm_relay_config rc", "rc.id = rdm.relay_id").
		Where("rc.active", 1).
		Fields("rdm.sender_domain").
		Scan(&mappings)
	if err != nil {
		return nil, fmt.Errorf("failed to query relay domain mappings: %v", err)
	}

	result := make(map[string]bool, len(mappings))
	for _, m := range mappings {
		domain := strings.TrimPrefix(m.SenderDomain, "@")
		if domain != "" {
			result[domain] = true
		}
	}
	return result, nil
}

func Exists(ctx context.Context, domainName string) (bool, error) {
	count, err := g.DB().Model("domain").
		Ctx(ctx).
		Where("domain", domainName).
		Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func buildCacheKey(domain, recordType string) string {
	return fmt.Sprintf("DOMAIN_DNS_RECORDS_:%s:_%s", domain, recordType)
}

func FreshRecords(ctx context.Context, domain ...string) {
	g.Log().Debug(ctx, "Fresh DNS records...")

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

	cacheSeconds := 600

	wg := sync.WaitGroup{}

	for _, d := range domains {
		// goroutine for each record type
		wg.Add(7) // total of 7 record types

		// retrieve A record
		go func(d string) {
			defer wg.Done()
			dr, err := GetARecord(d, true, "")
			if err != nil {
				g.Log().Error(ctx, "Failed to get A record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "A"), dr, cacheSeconds)
			}
		}(d)

		// retrieve MX record
		go func(d string) {
			defer wg.Done()
			dr, err := GetMXRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get MX record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "MX"), dr, cacheSeconds)
			}
		}(d)

		// retrieve SPF record
		go func(d string) {
			defer wg.Done()
			dr, err := GetSPFRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get SPF record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "SPF"), dr, cacheSeconds)
			}
		}(d)

		// retrieve DKIM record
		go func(d string) {
			defer wg.Done()
			dr, err := GetDKIMRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get DKIM record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "DKIM"), dr, cacheSeconds)
			}
		}(d)

		// retrieve DKIM short record
		go func(d string) {
			defer wg.Done()
			dr, err := GetDKIMShortRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get DKIM short record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "DKIM_SHORT"), dr, cacheSeconds)
			}
		}(d)

		// retrieve DMARC record
		go func(d string) {
			defer wg.Done()
			dr, err := GetDMARCRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get DMARC record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "DMARC"), dr, cacheSeconds)
			}
		}(d)

		//retrieve PTR record
		go func(d string) {
			defer wg.Done()
			dr, err := GetPTRRecord(d, true)
			if err != nil {
				g.Log().Error(ctx, "Failed to get PTR record for domain %s: %v", d, err)
			} else {
				public.SetCache(buildCacheKey(d, "PTR"), dr, cacheSeconds)
			}
		}(d)
	}

	wg.Wait()

	g.Log().Debug(ctx, "Fresh DNS records completed.")
}

// GetDKIMRecord retrieves the DKIM record for a given domain.
func GetDKIMRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	return getDKIMRecordWithKeySize(domain, "default", 2048, validateImmediate)
}

// GetDKIMShortRecord retrieves the short DKIM record for a given domain.
func GetDKIMShortRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	return getDKIMRecordWithKeySize(domain, "short", 1024, validateImmediate)
}

func getDKIMRecordWithKeySize(domain, selector string, keySize int, validateImmediate bool) (record v1.DNSRecord, err error) {
	// Create DKIM directory
	dkimPath := public.AbsPath(filepath.Join(consts.RSPAMD_LIB_PATH, "dkim", domain))

	// Check if directory exists
	if !public.IsDir(dkimPath) {
		_ = os.MkdirAll(dkimPath, 0755)
	}

	// Check if DKIM private and public key files exist
	dkimPriPath := filepath.Join(dkimPath, selector+".private")
	dkimPubPath := filepath.Join(dkimPath, selector+".pub")

	var dk *docker.DockerAPI

	dk, err = docker.NewDockerAPI()
	if err != nil {
		err = fmt.Errorf("Failed to connect to Docker API: %v", err)
		return
	}
	defer dk.Close()

	// Generate new keys if they don't exist
	if !public.FileExists(dkimPriPath) || !public.FileExists(dkimPubPath) {
		mutex.Lock()
		defer mutex.Unlock()

		var res *v2.ExecResult
		res, err = dk.ExecCommandByName(context.Background(), consts.SERVICES.Rspamd, []string{"rspamadm", "dkim_keygen", "-s", selector, "-b", fmt.Sprintf("%d", keySize), "-d", domain, "-k", fmt.Sprintf("/var/lib/rspamd/dkim/%s/%s.private", domain, selector)}, "root")
		if err != nil {
			err = fmt.Errorf("Failed to generate DKIM key pair: %v", err)
			return
		}

		if res != nil {
			_, err = public.WriteFile(dkimPubPath, res.Output)
			if err != nil {
				err = fmt.Errorf("Failed to write DKIM public key: %v", err)
				return
			}
		}

		// update dkim private key file permission to 0644
		err = os.Chmod(dkimPriPath, 0644)
		if err != nil {
			err = fmt.Errorf("Failed to change DKIM private key permissions: %v", err)
			return
		}

		// Skip DKIM signing config for relay-mapped domains — relay provider signs
		relayDomains, relayErr := GetRelayDomains(context.Background())
		if relayErr != nil {
			g.Log().Warning(context.Background(), "Failed to check relay domains for DKIM signing:", relayErr)
			relayDomains = make(map[string]bool)
		}

		if !relayDomains[domain] {
			// build DKIM Sign config
			signConf := fmt.Sprintf(`
#%s_DKIM_BEGIN
%s {
   selectors [
    {
      path: "/var/lib/rspamd/dkim/%s/default.private";
      selector: "default";
    },
    {
      path: "/var/lib/rspamd/dkim/%s/short.private";
      selector: "short";
    }
  ]
}
#%s_DKIM_END
`, domain, domain, domain, domain, domain)

			// Write DKIM sign config to file
			signConfPath := public.AbsPath(filepath.Join(consts.RSPAMD_LOCAL_D_PATH, "dkim_signing.conf"))
			signContent := `sign_headers = "from:sender:reply-to:subject:date:message-id:to:cc:mime-version:content-type:content-transfer-encoding:content-language:resent-to:resent-cc:resent-from:resent-sender:resent-message-id:in-reply-to:references:list-id:list-help:list-owner:list-unsubscribe:list-subscribe:list-post:list-unsubscribe-post:disposition-notification-to:disposition-notification-options:original-recipient:openpgp:autocrypt";

domain {
#BT_DOMAIN_DKIM_BEGIN
#BT_DOMAIN_DKIM_END
}`

			if public.FileExists(signConfPath) {
				signContent, err = public.ReadFile(signConfPath)
				if err != nil {
					err = fmt.Errorf("Failed to read DKIM sign config: %v", err)
					return
				}
			}

			// Remove old config block if it exists
			pattern := fmt.Sprintf(`(?s)#%s_DKIM_BEGIN.*?#%s_DKIM_END\s*`, domain, domain)
			signContent, err = gregex.ReplaceString(pattern, "", signContent)
			if err != nil {
				return
			}

			signContent = strings.Replace(signContent, "#BT_DOMAIN_DKIM_END", signConf+"\n#BT_DOMAIN_DKIM_END", 1)
			_, err = public.WriteFile(signConfPath, signContent)
			if err != nil {
				err = fmt.Errorf("Failed to write DKIM sign config: %v", err)
				return
			}

			// Restart rspamd service
			err = dk.RestartContainerByName(context.Background(), consts.SERVICES.Rspamd)
			if err != nil {
				err = fmt.Errorf("Failed to restart rspamd container: %v", err)
				return
			}
		}
	}

	// DKIM public key is typically stored in a specific location in the container or host
	// Assuming we use docker exec to read the DKIM public key from the rspamd container
	dkimPub, err := public.ReadFile(dkimPubPath)
	if err != nil {
		err = fmt.Errorf("Cannot read DKIM public key: %v", err)
		return
	}

	// Format DKIM record
	// Expected format is a pre-formatted TXT record value like "v=DKIM1; k=rsa; p=MIIBIjANBg..."
	dkimRecord := strings.TrimSpace(dkimPub)

	// If the raw public key is read, format it into DNS TXT record format
	if !strings.Contains(dkimRecord, "v=DKIM1") && !strings.Contains(dkimRecord, "k=rsa") && !strings.Contains(dkimRecord, "p=") {
		// Remove possible header/footer markers and newlines
		dkimRecord = strings.ReplaceAll(dkimRecord, "-----BEGIN PUBLIC KEY-----", "")
		dkimRecord = strings.ReplaceAll(dkimRecord, "-----END PUBLIC KEY-----", "")
		dkimRecord = strings.ReplaceAll(dkimRecord, "\n", "")
		dkimRecord = strings.TrimSpace(dkimRecord)
		dkimRecord = fmt.Sprintf("v=DKIM1; k=rsa; p=%s", dkimRecord)
	} else {
		var ms [][]string
		ms, err = gregex.MatchAllString(`"([^"\r\n]+)"`, dkimRecord)
		if err != nil {
			err = fmt.Errorf("Failed to parse DKIM record: %v", err)
			return
		}

		if len(ms) < 2 {
			err = fmt.Errorf("Invalid DKIM record format")
			return
		}

		s := ""
		for _, v := range ms {
			if len(v) < 2 {
				continue
			}
			s += v[1]
		}

		dkimRecord = s
	}

	record = v1.DNSRecord{
		Type:  "TXT",
		Host:  selector + "._domainkey",
		Value: dkimRecord,
	}

	if validateImmediate {
		// Validate the DKIM record
		record.Valid = ValidateTXTRecord(record, domain)
	}

	return
}

// GetDMARCRecord retrieves the DMARC record for a given domain.
func GetDMARCRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	record = v1.DNSRecord{
		Type:  "TXT",
		Host:  "_dmarc",
		Value: fmt.Sprintf("v=DMARC1;p=quarantine;rua=mailto:admin@%s", domain),
	}

	if validateImmediate {
		// Validate the DMARC record
		record.Valid = ValidateTXTRecord(record, domain)
	}

	return
}

// GetSPFRecord retrieves the SPF record for a given domain.
func GetSPFRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	serverIP, err := public.GetServerIP()

	if err != nil {
		err = fmt.Errorf("Failed to get server IP: %v", err)
		return
	}

	ipType := "ip4"

	if strings.Contains(serverIP, ":") {
		ipType = "ip6"
	}

	// Format SPF record
	record = v1.DNSRecord{
		Type:  "TXT",
		Host:  "@",
		Value: fmt.Sprintf("v=spf1 +a +mx +%s:%s -all", ipType, serverIP),
	}

	if validateImmediate {
		// Validate the SPF record
		record.Valid = ValidateTXTRecord(record, domain)
	}

	return
}

// GetMXRecord retrieves the MX record for a given domain.
func GetMXRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	record = v1.DNSRecord{
		Type:  "MX",
		Host:  "@",
		Value: public.FormatMX(domain),
	}

	if validateImmediate {
		// Validate the MX record
		record.Valid = ValidateMXRecord(record, domain, public.FormatMX(domain))
	}

	return
}

// GetARecord retrieves the A record for a given domain.
// If dedicatedIP is non-empty, it is used instead of the server's default IP.
func GetARecord(domain string, validateImmediate bool, dedicatedIP string) (record v1.DNSRecord, err error) {
	ip := dedicatedIP
	if ip == "" {
		ip, err = public.GetServerIP()
		if err != nil {
			err = fmt.Errorf("Failed to get server IP: %v", err)
			return
		}
	}

	recordType := "A"

	if strings.Contains(ip, ":") {
		recordType = "AAAA"
	}

	record = v1.DNSRecord{
		Type:  recordType,
		Host:  public.FormatMX(domain),
		Value: ip,
	}

	if validateImmediate {
		// Validate the A record
		record.Valid = ValidateARecord(record)
	}

	return
}

// GetPTRRecord retrieves the PTR record for a given domain.
func GetPTRRecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	serverIP, err := public.GetServerIP()

	if err != nil {
		err = fmt.Errorf("Failed to get server IP: %v", err)
		return
	}

	record = v1.DNSRecord{
		Type:  "PTR",
		Host:  serverIP,
		Value: public.FormatMX(domain),
	}

	if validateImmediate {
		// Validate the PTR record
		record.Valid = ValidatePTRRecord(record)
	}

	return
}

// GetRecordsInCache retrieves DNS records from the cache for a given domain.
// If dedicatedIP is non-empty, the A record uses that IP instead of the server default.
func GetRecordsInCache(domain string, dedicatedIP string) (records v1.DNSRecords) {
	// Get A record — bypass cache when a dedicated IP is set
	if dedicatedIP != "" {
		records.A, _ = GetARecord(domain, false, dedicatedIP)
	} else {
		aRecord := public.GetCache(buildCacheKey(domain, "A"))
		if aRecord != nil {
			if v, ok := aRecord.(v1.DNSRecord); ok {
				records.A = v
			}
		} else {
			records.A, _ = GetARecord(domain, false, "")
		}
	}

	// Get MX record from cache
	mxRecord := public.GetCache(buildCacheKey(domain, "MX"))

	if mxRecord != nil {
		if v, ok := mxRecord.(v1.DNSRecord); ok {
			records.MX = v
		}
	} else {
		records.MX, _ = GetMXRecord(domain, false)
	}

	// Get SPF record from cache
	spfRecord := public.GetCache(buildCacheKey(domain, "SPF"))

	if spfRecord != nil {
		if v, ok := spfRecord.(v1.DNSRecord); ok {
			records.SPF = v
		}
	} else {
		records.SPF, _ = GetSPFRecord(domain, false)
	}

	// Get DKIM record from cache
	dkimRecord := public.GetCache(buildCacheKey(domain, "DKIM"))

	if dkimRecord != nil {
		if v, ok := dkimRecord.(v1.DNSRecord); ok {
			records.DKIM = v
		}
	} else {
		records.DKIM, _ = GetDKIMRecord(domain, false)
	}

	// Get DKIM short record from cache
	dkimShortRecord := public.GetCache(buildCacheKey(domain, "DKIM_SHORT"))
	if dkimShortRecord != nil {
		if v, ok := dkimShortRecord.(v1.DNSRecord); ok {
			records.DKIMShort = v
		}
	} else {
		records.DKIMShort, _ = GetDKIMShortRecord(domain, false)
	}

	// Get DMARC record from cache
	dmarcRecord := public.GetCache(buildCacheKey(domain, "DMARC"))

	if dmarcRecord != nil {
		if v, ok := dmarcRecord.(v1.DNSRecord); ok {
			records.DMARC = v
		}
	} else {
		records.DMARC, _ = GetDMARCRecord(domain, false)
	}

	// Get PTR record from cache
	ptrRecord := public.GetCache(buildCacheKey(domain, "PTR"))

	if ptrRecord != nil {
		if v, ok := ptrRecord.(v1.DNSRecord); ok {
			records.PTR = v
		}
	} else {
		records.PTR, _ = GetPTRRecord(domain, false)
	}

	return
}

// RepairDKIMSigningConfig repairs the DKIM signing configuration file.
func RepairDKIMSigningConfig(ctx context.Context) error {
	g.Log().Debug(ctx, "Repairing DKIM signing config...")
	defer func() {
		g.Log().Debug(ctx, "Repairing DKIM signing config completed.")
	}()

	// 1. Get all domains
	ds, err := All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all domains: %v", err)
	}

	// 1b. Get relay-mapped domains to exclude from DKIM signing
	relayDomains, err := GetRelayDomains(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to get relay domains, signing all: %v", err)
		relayDomains = make(map[string]bool)
	}

	// 2. Build the full DKIM config content
	var allSignConfBlocks strings.Builder
	for _, d := range ds {
		// Skip domains with active relay — relay provider signs DKIM
		if relayDomains[d.Domain] {
			g.Log().Debugf(ctx, "Skipping DKIM signing for relay-mapped domain: %s", d.Domain)
			continue
		}
		// Regenerate missing DKIM key files
		for _, sel := range []struct {
			name string
			size int
		}{{"default", 2048}, {"short", 1024}} {
			keyPath := public.AbsPath(filepath.Join(consts.RSPAMD_LIB_PATH, "dkim", d.Domain, sel.name+".private"))
			if !public.FileExists(keyPath) {
				g.Log().Warningf(ctx, "DKIM key missing for %s/%s, regenerating", d.Domain, sel.name)
				getDKIMRecordWithKeySize(d.Domain, sel.name, sel.size, false)
			}
		}
		// For each domain, generate the correct config block with both selectors
		signConf := fmt.Sprintf(`
#%s_DKIM_BEGIN
%s {
   selectors [
    {
      path: "/var/lib/rspamd/dkim/%s/default.private";
      selector: "default";
    },
    {
      path: "/var/lib/rspamd/dkim/%s/short.private";
      selector: "short";
    }
  ]
}
#%s_DKIM_END
`, d.Domain, d.Domain, d.Domain, d.Domain, d.Domain)
		allSignConfBlocks.WriteString(signConf)
	}

	// 3. Construct the final dkim_signing.conf content
	signConfPath := public.AbsPath(filepath.Join(consts.RSPAMD_LOCAL_D_PATH, "dkim_signing.conf"))
	finalSignContent := fmt.Sprintf(`sign_headers = "from:sender:reply-to:subject:date:message-id:to:cc:mime-version:content-type:content-transfer-encoding:content-language:resent-to:resent-cc:resent-from:resent-sender:resent-message-id:in-reply-to:references:list-id:list-help:list-owner:list-unsubscribe:list-subscribe:list-post:list-unsubscribe-post:disposition-notification-to:disposition-notification-options:original-recipient:openpgp:autocrypt";

domain {
#BT_DOMAIN_DKIM_BEGIN
%s
#BT_DOMAIN_DKIM_END
}`, allSignConfBlocks.String())

	// 4. Write the new content to the file, overwriting the old one
	_, err = public.WriteFile(signConfPath, finalSignContent)
	if err != nil {
		return fmt.Errorf("failed to write DKIM signing config: %v", err)
	}

	// 5. Ensure correct file permissions
	err = filepath.Walk(filepath.Join(public.AbsPath(consts.RSPAMD_LIB_PATH), "dkim"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, 0755)
		}
		if strings.HasSuffix(path, ".private") || strings.HasSuffix(path, ".pub") {
			return os.Chmod(path, 0644)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to change DKIM file permissions: %v", err)
	}

	// 6. Restart rspamd service
	dk, err := docker.NewDockerAPI()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker API: %v", err)
	}
	defer dk.Close()

	err = dk.RestartContainerByName(ctx, consts.SERVICES.Rspamd)
	if err != nil {
		return fmt.Errorf("failed to restart rspamd container: %v", err)
	}

	return nil
}
