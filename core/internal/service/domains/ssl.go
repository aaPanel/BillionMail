package domains

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/acme"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/public"
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/frame/g"
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
	err = g.DB().Model("letsencrypts").Where("dns::jsonb ? $1", public.FormatMX(domain)).Where("status = 1").Where("endtime > $1", time.Now().Unix()).Order("endtime desc").Limit(1).Scan(&crt)
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
