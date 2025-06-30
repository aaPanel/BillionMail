package domains

import (
	v2 "billionmail-core/api/dockerapi/v1"
	v1 "billionmail-core/api/domains/v1"
	mail_v1 "billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/consts"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/mail_boxes"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/public"
	"context"
	"database/sql"
	"fmt"
	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gregex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	mutex sync.Mutex
)

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
			_, err = public.DockerApiFromCtx(ctx).ExecCommandByName(ctx, "billionmail-postfix-billionmail-1", []string{"bash", "-c", fmt.Sprintf("sed -i '/^BILLIONMAIL_HOSTNAME=/d' /postfix.sh && sed -i '/^#!\\/bin\\/bash/a BILLIONMAIL_HOSTNAME=%s' /postfix.sh", public.FormatMX(domain.Domain))}, "root")

			if err != nil {
				return fmt.Errorf("failed to update BILLIONMAIL_HOSTNAME in postfix container: %v", err)
			}

			// restart postfix service
			err = public.DockerApiFromCtx(ctx).RestartContainerByName(ctx, "billionmail-postfix-billionmail-1")

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

func Update(ctx context.Context, domain *v1.Domain) error {
	_, err := g.DB().Model("domain").
		Ctx(ctx).
		Where("domain", domain.Domain).
		Update(domain)

	// catchall
	err = setCatchall(ctx, domain.Domain, domain.Catchall)

	return err
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

	crt := mail_service.NewCertificate()

	defer crt.Close()

	wg := sync.WaitGroup{}

	for i, domain := range domains {
		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()
			domains[i].DNSRecords = GetRecordsInCache(strings.TrimPrefix(domain.Domain, "mail."))
		}(i, domain)

		// Retrieve Domains SSL certificate information
		wg.Add(1)
		go func(i int, domain v1.Domain) {
			defer wg.Done()
			domains[i].CertInfo, err = crt.GetSSLInfo(domain.Domain)

			if err != nil {
				err = nil
				domains[i].CertInfo, _ = crt.GetSSLInfo(public.FormatMX(domain.Domain))
			}

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
		wg.Add(6) // total of 6 record types

		// retrieve A record
		go func(d string) {
			defer wg.Done()
			dr, err := GetARecord(d, true)
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
	// Create DKIM directory
	dkimPath := public.AbsPath(filepath.Join(consts.RSPAMD_LIB_PATH, "dkim", domain))

	// Check if directory exists
	if !public.IsDir(dkimPath) {
		_ = os.MkdirAll(dkimPath, 0755)
	}

	// Check if DKIM private and public key files exist
	dkimPriPath := filepath.Join(dkimPath, "default.private")
	dkimPubPath := filepath.Join(dkimPath, "default.pub")

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
		res, err = dk.ExecCommandByName(context.Background(), "billionmail-rspamd-billionmail-1", []string{"rspamadm", "dkim_keygen", "-s", "'default'", "-b", "2048", "-d", domain, "-k", fmt.Sprintf("/var/lib/rspamd/dkim/%s/default.private", domain)}, "root")

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

		// build DKIM Sign config
		signConf := fmt.Sprintf(`#%s_DKIM_BEGIN
%s {
   selectors [
    {
      path: "/var/lib/rspamd/dkim/%s/default.private";
      selector: "default";
    }
  ]
}
#%s_DKIM_END
`, domain, domain, domain, domain)

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

		signContent = strings.Replace(signContent, "#BT_DOMAIN_DKIM_END", signConf+"\n#BT_DOMAIN_DKIM_END", 1)

		_, err = public.WriteFile(signConfPath, signContent)

		if err != nil {
			err = fmt.Errorf("Failed to write DKIM sign config: %v", err)
			return
		}

		// Restart rspamd service
		err = dk.RestartContainerByName(context.Background(), "billionmail-rspamd-billionmail-1")

		if err != nil {
			err = fmt.Errorf("Failed to restart rspamd container: %v", err)
			return
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
		Host:  "default._domainkey",
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
		Value: "mail." + domain,
	}

	if validateImmediate {
		// Validate the MX record
		record.Valid = ValidateMXRecord(record, domain)
	}

	return
}

// GetARecord retrieves the A record for a given domain.
func GetARecord(domain string, validateImmediate bool) (record v1.DNSRecord, err error) {
	serverIP, err := public.GetServerIP()

	if err != nil {
		err = fmt.Errorf("Failed to get server IP: %v", err)
		return
	}

	recordType := "A"

	if strings.Contains(serverIP, ":") {
		recordType = "AAAA"
	}

	record = v1.DNSRecord{
		Type:  recordType,
		Host:  "mail." + domain,
		Value: serverIP,
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
		Value: "mail." + domain,
	}

	if validateImmediate {
		// Validate the PTR record
		record.Valid = ValidatePTRRecord(record)
	}

	return
}

// GetRecordsInCache retrieves DNS records from the cache for a given domain.
func GetRecordsInCache(domain string) (records v1.DNSRecords) {
	// Get A record from cache
	aRecord := public.GetCache(buildCacheKey(domain, "A"))

	if aRecord != nil {
		if v, ok := aRecord.(v1.DNSRecord); ok {
			records.A = v
		}
	} else {
		records.A, _ = GetARecord(domain, false)
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

	// Read the DKIM signing configuration file
	signConfPath := public.AbsPath(filepath.Join(consts.RSPAMD_LOCAL_D_PATH, "dkim_signing.conf"))
	signContent, err := public.ReadFile(signConfPath)

	if err != nil {
		return fmt.Errorf("Failed to read DKIM signing config: %v", err)
	}

	// Check if the config contains the DKIM domain section
	if !strings.Contains(signContent, "#BT_DOMAIN_DKIM_BEGIN") || !strings.Contains(signContent, "#BT_DOMAIN_DKIM_END") {
		return fmt.Errorf("DKIM signing config is missing domain sections")
	}

	// fix the invalid selector private key path
	signContent = strings.ReplaceAll(signContent, public.AbsPath(consts.RSPAMD_LIB_PATH), "/var/lib/rspamd")

	// Write the repaired content back to the file
	_, err = public.WriteFile(signConfPath, signContent)

	if err != nil {
		return fmt.Errorf("Failed to write DKIM signing config: %v", err)
	}

	// update dkim files permission to 0644
	filepath.Walk(filepath.Join(public.AbsPath(consts.RSPAMD_LIB_PATH), "dkim"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// set directory permission to 0755
			err = os.Chmod(path, 0755)

			if err != nil {
				return fmt.Errorf("Failed to change DKIM directory permissions: %v", err)
			}
		}

		if strings.HasSuffix(path, ".private") || strings.HasSuffix(path, ".pub") {
			err = os.Chmod(path, 0644)

			if err != nil {
				return fmt.Errorf("Failed to change DKIM file permissions: %v", err)
			}
		}

		return nil
	})

	// Restart rspamd service
	dk, err := docker.NewDockerAPI()
	if err != nil {
		return fmt.Errorf("Failed to connect to Docker API: %v", err)
	}

	defer dk.Close()

	err = dk.RestartContainerByName(ctx, "billionmail-rspamd-billionmail-1")
	if err != nil {
		return fmt.Errorf("Failed to restart rspamd container: %v", err)
	}

	return nil
}
