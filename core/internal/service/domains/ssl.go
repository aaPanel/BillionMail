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
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"net/url"
	"path/filepath"
	"time"
)

// ApplyLetsEncryptCertWithHttp applies for a Let's Encrypt certificate for the given domain.
func ApplyLetsEncryptCertWithHttp(ctx context.Context, domain string, accountInfo *model.Account) error {
	// Find the existing certificate for the domain
	if crt, err := FindSSLByDomain(domain); err == nil && crt != nil {
		if crt.Status == 1 && crt.EndTime > time.Now().Unix() {
			g.Log().Debug(ctx, "Found existing certificate for domain:", domain)
			return ApplyCertToService(domain, crt.Certificate, crt.PrivateKey)
		}
	}

	certificate, privateKey, err := acme.ApplySSLWithExistingServer(ctx, []string{public.FormatMX(domain)}, accountInfo.Email, "http", "", nil, public.AbsPath(consts.SSL_PATH))

	if err != nil {
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

	// Save the certificate and key to the database
	_, err = g.DB().Model("letsencrypts").Insert(pdata)

	if err != nil {
		return err
	}

	// Apply the certificate to the service, postfix, dovecot, etc.
	return ApplyCertToService(domain, certificate, privateKey)
}

// FindSSLByDomain retrieves the SSL certificate information for a given domain.
func FindSSLByDomain(domain string) (crt *entity.Letsencrypt, err error) {
	err = g.DB().Model("letsencrypts").Where("dns::jsonb ? $1", public.FormatMX(domain)).Where("status = 1").Where("endtime > ?", time.Now().Unix()).Order("endtime desc").Limit(1).Scan(&crt)
	return
}

// ApplyCertToService applies the SSL certificate to the service.
func ApplyCertToService(domain, crtPem, keyPem string) (err error) {
	crt := mail_service.NewCertificate()

	defer crt.Close()

	err = crt.SetSNI(public.FormatMX(domain), crtPem, keyPem)

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

	if u.Hostname() == public.FormatMX(domain) {
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

			err = dk.RestartContainerByName(context.Background(), "billionmail-core-billionmail-1")
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

		err = dk.RestartContainerByName(context.Background(), "billionmail-core-billionmail-1")
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
		Where("dns::jsonb ? $1", hostname).
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
		if crt.Status == 1 && crt.EndTime > time.Now().Unix() {
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
	Arecords, _ := GetARecord(hostname, false)

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

	// Save the certificate and key to the database
	_, err = g.DB().Model("letsencrypts").Insert(pdata)

	if err != nil {
		return gerror.Newf("Failed to save certificate to database: %v", err)
	}

	if err := ApplyCertToConsole(certificate, privateKey); err != nil {
		return err
	}
	return nil
}

// Auto-renew SSL certificate
func AutoRenewSSL(ctx context.Context) {
	// Renew console SSL certificate
	certInfo, err := GetConsoleSSLInfo()
	if err == nil && certInfo.Endtime > 0 {
		remain := certInfo.Endtime - int(time.Now().Unix())
		if remain < 3*24*3600 {
			err = ApplyConsoleCert(ctx)
			if err != nil {
				g.Log().Warningf(ctx, "Console certificate is about to expire, auto-renewal failed: %v", err)
			} else {
				g.Log().Info(ctx, "Console certificate is about to expire, auto-renewal succeeded")
			}
		}
	}

	// admin account
	adminAccount := &model.Account{}
	adminUsername, err := public.DockerEnv("ADMIN_USERNAME")
	err = g.DB().Model("account").Where("username = ?", adminUsername).Scan(adminAccount)
	if err != nil {
		g.Log().Error(ctx, "Failed to get admin account for SSL renewal: ", err)
		return
	}

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

	for _, domain := range domainList {
		certInfo, err = mail_service.NewCertificate().GetSSLInfo(domain)
		if err == nil && certInfo.Endtime > 0 {
			remain := certInfo.Endtime - int(time.Now().Unix())
			//g.Log().Warningf(ctx, "Remaining time for domain certificate: %d seconds, domain: %s", remain, domain)
			if remain < 3*24*3600 {
				//g.Log().Info(ctx, "Domain certificate is about to expire, auto-renewing: ", domain)

				certErr := ApplyLetsEncryptCertWithHttp(ctx, domain, adminAccount)
				if certErr != nil {
					g.Log().Warningf(ctx, "Domain name [%s]  auto-request certificate failed: %v", domain, certErr)
				}

			}
		}
		// Skip if no certificate exists (endtime=0)
	}

}
