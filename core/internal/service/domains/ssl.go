package domains

import (
	"billionmail-core/api/domains/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/acme"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/public"
	"billionmail-core/internal/service/rbac"
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// ApplyLetsEncryptCertWithHttp applies for a Let's Encrypt certificate for the given domain.
func ApplyLetsEncryptCertWithHttp(ctx context.Context, domain string, accountInfo *model.Account) error {
	formattedDomain := public.FormatMX(domain)

	// Find the existing certificate for the domain (including expired)
	var crt *entity.Letsencrypt
	g.DB().Model("letsencrypts").
		Where("subject = ?", formattedDomain).
		Order("endtime desc").
		Limit(1).Scan(&crt)

	if crt != nil && crt.CertId != 0 {
		if crt.Status == 1 && crt.EndTime > time.Now().AddDate(0, 0, 30).Unix() {
			g.Log().Debug(ctx, "[ApplyLetsEncrypt] Found valid certificate for domain:", formattedDomain)
			return ApplyCertToService(domain, crt.Certificate, crt.PrivateKey)
		}

		// If last attempt failed recently (status=-1) and error_info indicates rate limit,
		// skip re-applying to avoid hitting LE rate limits repeatedly
		if crt.Status == -1 && crt.ErrorInfo != "" {
			// Check if the error contains rate limit info
			if strings.Contains(crt.ErrorInfo, "rateLimited") || strings.Contains(crt.ErrorInfo, "too many") {
				g.Log().Warning(ctx, "[ApplyLetsEncrypt] Skipping due to previous rate limit error for domain:", formattedDomain)
				return gerror.Newf("Rate limited, please wait and try later. Last error: %s", crt.ErrorInfo)
			}
		}
	}

	certificate, privateKey, err := acme.ApplySSLWithExistingServer(ctx, []string{formattedDomain}, accountInfo.Email, "http", "", nil, public.AbsPath(consts.SSL_PATH))

	if err != nil {
		// Save error info to database for debugging
		pdata := g.Map{
			"error_info": err.Error(),
			"status":     -1,
		}
		g.DB().Model("letsencrypts").
			Where("subject = ?", formattedDomain).
			Update(pdata)
		return err
	}

	dnsNames := "[]"
	notAfter := "0000-00-00"
	notBefore := "0000-00-00"
	endTime := 0
	status := -1
	subject := ""

	if certificate != "" {
		certInfo := acme.GetCertInfo(certificate)
		notAfter = certInfo.NotAfter
		notBefore = certInfo.NotBefore
		subject = certInfo.Subject
		endTime = certInfo.Endtime
		dnsNamesBytes, err := json.Marshal([]string{public.FormatMX(domain)})
		if err == nil {
			dnsNames = string(dnsNamesBytes)
		}
		status = 1
	}

	// get the progress of the certificate application
	progress := acme.GetAcmeLogBody(ctx)
	pdata := g.Map{
		"account_id":  accountInfo.AccountId,
		"certificate": certificate,
		"private_key": privateKey,
		"error_info":  "",
		"progress":    progress,
		"status":      status,
		"not_after":   notAfter,
		"not_before":  notBefore,
		"dns":         dnsNames,
		"endtime":     endTime,
		"subject":     subject,
	}

	// Update existing record for this domain if one exists (including expired),
	// otherwise insert a new one. This prevents accumulating stale expired
	// certificate records for the same domain.
	affected, err := g.DB().Model("letsencrypts").
		Where("subject = ?", formattedDomain).
		Update(pdata)

	if err != nil {
		return err
	}

	if rowsAffected, _ := affected.RowsAffected(); rowsAffected == 0 {
		// No existing record found, insert a new one
		_, err = g.DB().Model("letsencrypts").Insert(pdata)
		if err != nil {
			return err
		}
	}

	// Apply the certificate to the service, postfix, dovecot, etc.
	return ApplyCertToService(domain, certificate, privateKey)
}

// FindSSLByDomain retrieves the SSL certificate information for a given domain.
func FindSSLByDomain(domain string) (crt *entity.Letsencrypt, err error) {
	err = g.DB().Model("letsencrypts").
		Where("subject = ?", public.FormatMX(domain)).
		Where("status = 1").
		Where("endtime > ?", time.Now().Unix()).
		Order("endtime desc").
		Limit(1).Scan(&crt)
	return
}

// ApplyCertToService applies the SSL certificate to the service.
func ApplyCertToService(domain, crtPem, keyPem string) (err error) {
	formattedDomain := public.FormatMX(domain)
	crt := mail_service.NewCertificate()

	defer crt.Close()

	err = crt.SetSNI(formattedDomain, crtPem, keyPem)

	if err != nil {
		return err
	}

	UpdateBaseURL(context.Background())

	// Attempt apply the certificate to the console panel if domain is the console domain
	rawurl := GetBaseURL()

	u, err := url.Parse(rawurl)

	if err != nil {
		return err
	}

	g.Log().Debug(context.Background(), "HostName: ", u.Hostname())
	g.Log().Debug(context.Background(), "ssl domain: ", public.FormatMX(domain))

	if u.Hostname() == formattedDomain {
		_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.SSL_PATH, "cert.pem")), crtPem)

		if err != nil {
			return err
		}

		_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.SSL_PATH, "key.pem")), keyPem)

		if err != nil {
			return err
		}

		// Reload server ssl
		go func() {
			time.Sleep(time.Millisecond * 500)

			var dk *docker.DockerAPI
			dk, err = docker.NewDockerAPI()

			if err != nil {
				g.Log().Warning(context.Background(), "Get docker api instance failed")
				return
			}

			defer dk.Close()

			err = dk.RestartContainerByName(context.Background(), consts.SERVICES.Core)
		}()
	} else {

		csrPath := filepath.Join(consts.SSL_PATH, formattedDomain, "/fullchain.pem")
		ketPath := filepath.Join(consts.SSL_PATH, formattedDomain, "/privkey.pem")
		_, err = public.WriteFile(public.AbsPath(csrPath), crtPem)

		if err != nil {
			return err
		}

		_, err = public.WriteFile(public.AbsPath(ketPath), keyPem)

		if err != nil {
			return err
		}
		// Reload server ssl
		go func() {
			time.Sleep(time.Millisecond * 500)

			var dk *docker.DockerAPI
			dk, err = docker.NewDockerAPI()

			if err != nil {
				g.Log().Warning(context.Background(), "Get docker api instance failed")
				return
			}

			defer dk.Close()

			err = dk.RestartContainerByName(context.Background(), consts.SERVICES.Core)
		}()

	}

	return
}

func ApplyCertToConsole(crtPem, keyPem string) (err error) {

	_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.SSL_PATH, "cert.pem")), crtPem)

	if err != nil {
		return err
	}

	_, err = public.WriteFile(public.AbsPath(filepath.Join(consts.SSL_PATH, "key.pem")), keyPem)

	if err != nil {
		return err
	}

	// Reload server ssl
	go func() {
		time.Sleep(time.Millisecond * 500)

		var dk *docker.DockerAPI
		dk, err = docker.NewDockerAPI()

		if err != nil {
			g.Log().Warning(context.Background(), "Get docker api instance failed")
			return
		}

		defer dk.Close()

		err = dk.RestartContainerByName(context.Background(), consts.SERVICES.Core)
	}()

	return
}

func GetConsoleSSLInfo() (v1.CertInfo, error) {
	certPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "cert.pem"))
	keyPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "key.pem"))
	certInfo := v1.CertInfo{}
	if !public.FileExists(certPath) || !public.FileExists(keyPath) {
		return certInfo, nil
	}
	crtPem, err := public.ReadFile(certPath)
	if err != nil {
		return certInfo, err
	}
	err = gconv.Struct(acme.GetCertInfo(crtPem), &certInfo)
	if err == nil {
		certInfo.CertPem = crtPem
		certInfo.KeyPem, _ = public.ReadFile(keyPath)
	}
	return certInfo, nil
}

// ApplyConsoleCert Requesting a Console Certificate
func ApplyConsoleCert(ctx context.Context) error {
	// 取hostname
	envMap, err := public.LoadEnvFile()
	if err != nil {
		return err
	}
	hostname := envMap["BILLIONMAIL_HOSTNAME"]

	if hostname == "" {
		return gerror.New("BILLIONMAIL_HOSTNAME environment variable is not set")
	}
	//mailDomain := public.FormatMX(hostname)

	// Check for the existence of a certificate
	crt := &entity.Letsencrypt{}
	err = g.DB().Model("letsencrypts").
		Where("subject = ?", hostname).
		Where("status = 1").
		Where("endtime > ?", time.Now().Unix()).
		Order("endtime desc").
		Limit(1).Scan(crt)

	if err != nil && err != sql.ErrNoRows {
		g.Log().Warning(ctx, "letsencrypts query error:", err)
		return err
	}
	if crt.CertId == 0 || crt.Certificate == "" {
		g.Log().Debug(ctx, "No existing certificate found:", hostname)
	} else {
		if crt.Status == 1 && crt.EndTime > time.Now().AddDate(0, 0, 30).Unix() {
			err = ApplyCertToConsole(crt.Certificate, crt.PrivateKey)
			if err != nil {
				return gerror.Newf("Failed to apply existing certificate to console: %v", err)
			}
			return nil
		}

	}

	accountInfo, accErr := rbac.GetCurrentAccount(ctx)
	if accErr != nil {
		return accErr
	}

	// Check that the hostname a record exists
	Arecords, _ := GetARecord(hostname, false, "")

	if Arecords.Host == "" || Arecords.Value == "" {
		return gerror.Newf("A Record does not exist, please check DNS Settings: %s", hostname)
	}

	certificate, privateKey, applyErr := acme.ApplySSLWithExistingServer(ctx, []string{hostname}, accountInfo.Email, "http", "", nil, public.AbsPath(consts.SSL_PATH))
	if applyErr != nil {
		//g.Log().Error(ctx, "Failed to request console certificate: ", applyErr)
		return applyErr
	}

	if certificate == "" || privateKey == "" {
		return gerror.New("The content of the application certificate is empty")
	}
	dnsNames := "[]"
	notAfter := "0000-00-00"
	notBefore := "0000-00-00"
	endTime := 0
	status := -1
	subject := ""

	if certificate != "" {
		certInfo := acme.GetCertInfo(certificate)
		notAfter = certInfo.NotAfter
		notBefore = certInfo.NotBefore
		subject = certInfo.Subject
		endTime = certInfo.Endtime
		dnsNamesBytes, err := json.Marshal([]string{hostname})
		if err == nil {
			dnsNames = string(dnsNamesBytes)
		}
		status = 1
	}

	// get the progress of the certificate application
	progress := acme.GetAcmeLogBody(ctx)
	pdata := g.Map{
		"account_id":  accountInfo.AccountId,
		"certificate": certificate,
		"private_key": privateKey,
		"error_info":  "",
		"progress":    progress,
		"status":      status,
		"not_after":   notAfter,
		"not_before":  notBefore,
		"dns":         dnsNames,
		"endtime":     endTime,
		"subject":     subject,
	}

	// Update existing record for this domain if one exists (including expired),
	// otherwise insert a new one.
	affected, err := g.DB().Model("letsencrypts").
		Where("subject = ?", hostname).
		Update(pdata)

	if err != nil {
		return gerror.Newf("Failed to save certificate to database: %v", err)
	}

	if rowsAffected, _ := affected.RowsAffected(); rowsAffected == 0 {
		// No existing record found, insert a new one
		_, err = g.DB().Model("letsencrypts").Insert(pdata)
		if err != nil {
			return gerror.Newf("Failed to save certificate to database: %v", err)
		}
	}

	if err := ApplyCertToConsole(certificate, privateKey); err != nil {
		return err
	}
	return nil
}

// Auto-renew SSL certificate
func AutoRenewSSL(ctx context.Context) {
	g.Log().Info(ctx, "[AutoRenewSSL] Starting SSL auto-renew check...")

	// Renew console SSL certificate
	certInfo, err := GetConsoleSSLInfo()
	if err == nil && certInfo.Endtime > 0 {
		remain := certInfo.Endtime - int(time.Now().Unix())
		g.Log().Infof(ctx, "[AutoRenewSSL] Console cert remaining: %d seconds (%.1f days)", remain, float64(remain)/86400)
		if remain < 3*24*3600 {
			g.Log().Info(ctx, "[AutoRenewSSL] Console certificate is about to expire, attempting renewal")
			err = ApplyConsoleCert(ctx)
			if err != nil {
				g.Log().Warningf(ctx, "[AutoRenewSSL] Console certificate auto-renewal failed: %v", err)
			} else {
				g.Log().Info(ctx, "[AutoRenewSSL] Console certificate auto-renewal succeeded")
			}
		}
	} else {
		g.Log().Infof(ctx, "[AutoRenewSSL] Console cert info: endtime=%d, err=%v", certInfo.Endtime, err)
	}

	// admin account
	adminAccount := &model.Account{}
	adminUsername, err := public.DockerEnv("ADMIN_USERNAME")
	if err != nil {
		g.Log().Error(ctx, "[AutoRenewSSL] Failed to get ADMIN_USERNAME: ", err)
		return
	}
	err = g.DB().Model("account").Where("username = ?", adminUsername).Scan(adminAccount)
	if err != nil {
		g.Log().Error(ctx, "[AutoRenewSSL] Failed to get admin account for SSL renewal: ", err)
		return
	}
	g.Log().Infof(ctx, "[AutoRenewSSL] Admin account found: %s", adminAccount.Email)

	var domainList []string
	rows, err := g.DB().Model("domain").Fields("domain").All()

	if err == nil {
		for _, row := range rows {
			domain := row["domain"].String()
			if domain != "" {
				domainList = append(domainList, domain)
			}
		}
	}
	g.Log().Infof(ctx, "[AutoRenewSSL] Found %d domain(s) to check: %v", len(domainList), domainList)

	for _, domain := range domainList {
		formattedDomain := public.FormatMX(domain)
		g.Log().Infof(ctx, "[AutoRenewSSL] Checking domain: %s (formatted: %s)", domain, formattedDomain)

		certInfo, err = mail_service.NewCertificate().GetSSLInfo(formattedDomain)
		if err != nil {
			g.Log().Warningf(ctx, "[AutoRenewSSL] GetSSLInfo for %s returned error: %v", formattedDomain, err)
		}

		if certInfo.Endtime <= 0 {
			// No certificate file found, check if there's a DB record indicating a recent failed attempt
			var dbCrt *entity.Letsencrypt
			g.DB().Model("letsencrypts").
				Where("subject = ?", formattedDomain).
				Order("cert_id desc").
				Limit(1).Scan(&dbCrt)

			if dbCrt != nil && dbCrt.CertId != 0 && dbCrt.Status == -1 && dbCrt.ErrorInfo != "" {
				// Recent failed attempt (e.g., rate limited), skip
				g.Log().Warningf(ctx, "[AutoRenewSSL] Domain %s has recent failed attempt, skipping. Last error: %s",
					formattedDomain, dbCrt.ErrorInfo)
				continue
			}

			// No cert file and no recent failure, attempt to apply for the first time
			g.Log().Infof(ctx, "[AutoRenewSSL] Domain %s has no certificate file, attempting to apply", formattedDomain)
			certErr := ApplyLetsEncryptCertWithHttp(ctx, domain, adminAccount)
			if certErr != nil {
				g.Log().Warningf(ctx, "[AutoRenewSSL] Domain [%s] auto-request certificate failed: %v", formattedDomain, certErr)
			} else {
				g.Log().Infof(ctx, "[AutoRenewSSL] Domain [%s] auto-request certificate succeeded", formattedDomain)
			}
			continue
		}

		remain := certInfo.Endtime - int(time.Now().Unix())
		g.Log().Infof(ctx, "[AutoRenewSSL] Domain %s cert remaining: %d seconds (%.1f days), issuer: %s",
			formattedDomain, remain, float64(remain)/86400, certInfo.Issuer)

		if remain < 3*24*3600 {
			g.Log().Infof(ctx, "[AutoRenewSSL] Domain %s cert is about to expire, starting renewal...", formattedDomain)
			certErr := ApplyLetsEncryptCertWithHttp(ctx, domain, adminAccount)
			if certErr != nil {
				g.Log().Warningf(ctx, "[AutoRenewSSL] Domain [%s] auto-request certificate failed: %v", formattedDomain, certErr)
			} else {
				g.Log().Infof(ctx, "[AutoRenewSSL] Domain [%s] auto-request certificate succeeded", formattedDomain)
			}
		} else {
			g.Log().Infof(ctx, "[AutoRenewSSL] Domain %s cert is still valid, skipping renewal", formattedDomain)
		}
	}

	g.Log().Info(ctx, "[AutoRenewSSL] SSL auto-renew check completed")
}
