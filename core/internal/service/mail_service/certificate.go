package mail_service

import (
	v1 "billionmail-core/api/domains/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/acme"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/util/gconv"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Certificate manages SSL certificates for mail services
type Certificate struct {
	mutex             sync.Mutex
	dk                *docker.DockerAPI
	PostfixMainConf   string
	PostfixMasterConf string
	PostfixConfPath   string
	PostfixSNIPath    string
	DovecotSslConf    string
	DovecotConfPath   string
}

// NewCertificate creates a new certificate management service instance
func NewCertificate() *Certificate {
	return &Certificate{
		PostfixMainConf:   public.AbsPath(consts.POSTFIX_MAIN_CONF),
		PostfixMasterConf: public.AbsPath(consts.POSTFIX_MASTER_CONF),
		DovecotSslConf:    public.AbsPath(filepath.Join(consts.DOVECOT_CONF_D_PATH, "10-ssl.conf")),
		DovecotConfPath:   public.AbsPath(consts.DOVECOT_CONF_D_PATH),
		PostfixConfPath:   public.AbsPath(consts.POSTFIX_CONF_PATH),
		PostfixSNIPath:    public.AbsPath(filepath.Join(consts.POSTFIX_CONF_PATH, "vmail_ssl.map")),
	}
}

func (c *Certificate) dockerApiClient() *docker.DockerAPI {
	if c.dk == nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		if c.dk != nil {
			return c.dk
		}

		var err error

		c.dk, err = docker.NewDockerAPI()

		if err != nil {
			panic(fmt.Sprintf("failed to create docker API: %v", err))
		}
	}

	return c.dk
}

// Close closes the certificate management service
func (c *Certificate) Close() error {
	if c.dk != nil {
		return c.dk.Close()
	}

	return nil
}

// SetSSL configures SSL certificate for mail services
func (c *Certificate) SetSSL(csrPem, keyPem string) error {
	// Validate certificate data
	if err := c.verifyCertificate(csrPem, keyPem); err != nil {
		return err
	}

	// Update Postfix configuration
	if err := c.updatePostfixConfig(csrPem, keyPem); err != nil {
		return err
	}

	// Update Dovecot configuration
	if err := c.updateDovecotConfig(csrPem, keyPem); err != nil {
		return err
	}

	// Restart Postfix and Dovecot services
	if err := c.restartPostfix(); err != nil {
		return err
	}

	if err := c.restartDovecot(); err != nil {
		return err
	}

	return nil
}

// SetSNI configures SSL certificate for SNI
func (c *Certificate) SetSNI(domain, csrPem, keyPem string) error {
	// Validate certificate data
	if err := c.verifyCertificate(csrPem, keyPem); err != nil {
		return err
	}

	// Update Postfix configuration
	if err := c.updatePostfixVMailConfig(domain, csrPem, keyPem); err != nil {
		return err
	}

	// Update Dovecot SNI configuration
	if err := c.updateDovecotSNIConfig(domain, csrPem, keyPem); err != nil {
		return err
	}

	// Restart Postfix and Dovecot services
	if err := c.restartPostfix(); err != nil {
		return err
	}

	if err := c.restartDovecot(); err != nil {
		return err
	}

	return nil
}

// SetPostfixSSL configures SSL certificate for Postfix
func (c *Certificate) SetPostfixSSL(csrPem, keyPem string) error {
	// Validate certificate data
	if err := c.verifyCertificate(csrPem, keyPem); err != nil {
		return err
	}

	// Update Postfix configuration
	if err := c.updatePostfixConfig(csrPem, keyPem); err != nil {
		return err
	}

	// Update Postfix Master configuration
	if err := c.SetPostfixMasterSSL(); err != nil {
		return err
	}

	// Restart Postfix service
	if err := c.restartPostfix(); err != nil {
		return err
	}

	return nil
}

// SetPostfixMasterSSL enables SSL for Postfix Master
func (c *Certificate) SetPostfixMasterSSL() error {
	// Read Postfix Master configuration
	content, err := public.ReadFile(consts.POSTFIX_MASTER_CONF)
	if err != nil {
		return fmt.Errorf("failed to read postfix master config: %v", err)
	}

	content, err = gregex.ReplaceString(`\n*#\s*-o\s+smtpd_tls_auth_only=yes`, "\n  -o smtpd_tls_auth_only=yes", content)

	if err != nil {
		return fmt.Errorf("failed to update postfix master config: %v", err)
	}

	content, err = gregex.ReplaceString(`\n*#\s*-o\s+smtpd_tls_wrappermode=yes`, "\\n  -o smtpd_tls_wrappermode=yes", content)

	if err != nil {
		return fmt.Errorf("failed to update postfix master config: %v", err)
	}

	// Update Postfix Master configuration
	if err := os.WriteFile(c.PostfixMasterConf, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write postfix master config: %v", err)
	}

	return nil
}

// SetDovecotSSL configures SSL certificate for Dovecot
func (c *Certificate) SetDovecotSSL(csrPem, keyPem string) error {
	// Validate certificate data
	if err := c.verifyCertificate(csrPem, keyPem); err != nil {
		return err
	}

	// Update Dovecot configuration
	if err := c.updateDovecotConfig(csrPem, keyPem); err != nil {
		return err
	}

	// Restart Dovecot service
	if err := c.restartDovecot(); err != nil {
		return err
	}

	return nil
}

// verifyCertificate validates certificate data
func (c *Certificate) verifyCertificate(csrPem, keyPem string) error {
	// Check if certificate data is empty
	if csrPem == "" {
		return fmt.Errorf("certificate data is empty")
	}
	if keyPem == "" {
		return fmt.Errorf("private key data is empty")
	}

	// Validate certificate
	cInfo := acme.GetCertInfo(csrPem)

	if cInfo.Endtime == 0 {
		return fmt.Errorf("certificate is invalid")
	}

	return nil
}

// updatePostfixConfig updates Postfix configuration with new certificate
func (c *Certificate) updatePostfixConfig(csrPem, keyPem string) error {
	mainCf := public.AbsPath(consts.POSTFIX_MAIN_CONF)
	content, err := os.ReadFile(mainCf)
	if err != nil {
		return fmt.Errorf("failed to read postfix config: %v", err)
	}

	// Write certificate and key to files
	certPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "postfix.crt"))
	keyPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "postfix.key"))

	if err := os.WriteFile(certPath, []byte(csrPem), 0755); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}

	if err := os.WriteFile(keyPath, []byte(keyPem), 0755); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	// Update SSL certificate configuration
	config := string(content)
	config = c.updateConfigLine(config, "smtpd_tls_key_file", keyPath)
	config = c.updateConfigLine(config, "smtpd_tls_cert_file", certPath)

	if err := os.WriteFile(mainCf, []byte(config), 0755); err != nil {
		return fmt.Errorf("failed to write postfix config: %v", err)
	}

	return nil
}

// SetPostfixVMailCert configures SSL certificate for Postfix virtual mail
func (c *Certificate) SetPostfixVMailCert(domain, csrPem, keyPem string) error {
	// Validate certificate data
	if err := c.verifyCertificate(csrPem, keyPem); err != nil {
		return err
	}

	// Update Postfix Master configuration
	if err := c.SetPostfixMasterSSL(); err != nil {
		return err
	}

	// Update Postfix virtual mail configuration
	if err := c.updatePostfixVMailConfig(domain, csrPem, keyPem); err != nil {
		return err
	}

	// Restart Postfix service
	if err := c.restartPostfix(); err != nil {
		return err
	}

	return nil
}

// updatePostfixVMailConfig updates Postfix virtual mail configuration
func (c *Certificate) updatePostfixVMailConfig(domain, csrPem, keyPem string) error {
	// Ensure domain directory exists
	domainDir := filepath.Join(consts.SSL_PATH, domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain directory: %v", err)
	}

	vmailCert := filepath.Join(domainDir, "fullchain.pem")
	vmailKey := filepath.Join(domainDir, "privkey.pem")

	// Write certificate and key to files
	if err := os.WriteFile(vmailCert, []byte(csrPem), 0755); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}

	if err := os.WriteFile(vmailKey, []byte(keyPem), 0755); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	// Create SNI mapping table
	if err := c.updatePostfixSNIMap("mail."+strings.TrimPrefix(domain, "mail."), vmailCert, vmailKey); err != nil {
		return fmt.Errorf("failed to update SNI map: %v", err)
	}

	return nil
}

// updatePostfixSNIMap updates Postfix SNI mapping table
func (c *Certificate) updatePostfixSNIMap(domain, certPath, keyPath string) error {
	// Read Postfix configuration file
	content, err := os.ReadFile(c.PostfixSNIPath)
	if err != nil {
		return fmt.Errorf("failed to read postfix config: %v", err)
	}

	// Update SNI mapping table
	lines := strings.Split(string(content), "\n")
	addNewLine := true

	for i, line := range lines {
		if strings.HasPrefix(line, domain) {
			lines[i] = fmt.Sprintf("%s %s %s", domain, keyPath, certPath)
			addNewLine = false
			break
		}
	}

	if addNewLine {
		lines = append(lines, fmt.Sprintf("%s %s %s", domain, keyPath, certPath))
	}

	if err := os.WriteFile(c.PostfixSNIPath, []byte(strings.Join(lines, "\n")), 0755); err != nil {
		return fmt.Errorf("failed to write postfix config: %v", err)
	}

	// Rehash configuration
	if _, err := c.dockerApiClient().ExecCommandByName(context.Background(), "billionmail-postfix-billionmail-1", []string{"postmap", "/etc/postfix/conf/vmail_ssl.map"}, "root"); err != nil {
		return fmt.Errorf("failed to hash postfix config: %v", err)
	}

	return nil
}

// updateDovecotConfig updates Dovecot configuration with new certificate
func (c *Certificate) updateDovecotConfig(csrPem, keyPem string) error {
	dovecotConf := c.DovecotSslConf
	content, err := os.ReadFile(dovecotConf)
	if err != nil {
		return fmt.Errorf("failed to read dovecot config: %v", err)
	}

	// Write certificate and key to files
	certPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "dovecot.crt"))
	keyPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "dovecot.key"))

	if err := os.WriteFile(certPath, []byte(csrPem), 0755); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}

	if err := os.WriteFile(keyPath, []byte(keyPem), 0755); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	// Update SSL certificate configuration
	config := string(content)
	config = c.updateConfigLine(config, "ssl_cert", "<"+certPath)
	config = c.updateConfigLine(config, "ssl_key", "<"+keyPath)

	if err := os.WriteFile(dovecotConf, []byte(config), 0755); err != nil {
		return fmt.Errorf("failed to write dovecot config: %v", err)
	}

	return nil
}

// updateDovecotSNIConfig updates Dovecot SNI configuration
func (c *Certificate) updateDovecotSNIConfig(domain, certPem, keyPem string) error {
	// Ensure domain directory exists
	domainDir := filepath.Join(consts.SSL_PATH, domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain directory: %v", err)
	}

	sniCert := filepath.Join(domainDir, "fullchain.pem")
	sniKey := filepath.Join(domainDir, "privkey.pem")

	// Write certificate and key to files
	if err := os.WriteFile(sniCert, []byte(certPem), 0755); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}

	if err := os.WriteFile(sniKey, []byte(keyPem), 0755); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	// Read Dovecot configuration file
	content, err := public.ReadFile(c.DovecotSslConf)
	if err != nil {
		return fmt.Errorf("failed to read dovecot config: %v", err)
	}

	domain = public.FormatMX(domain)

	replaceStr := "\n#DOMAIN_SSL_BEGIN_" + domain +
		"\nlocal_name " + domain +
		" {\n    ssl_cert = < " + sniCert +
		"\n    ssl_key = < " + sniKey +
		"\n}\n#DOMAIN_SSL_END_" + domain + "\n"

	if strings.Contains(content, "#DOMAIN_SSL_BEGIN_"+domain) {
		// update
		content, err = gregex.ReplaceString(fmt.Sprintf(`#DOMAIN_SSL_BEGIN_%s[\s\S]+#DOMAIN_SSL_END_%s`, domain, domain), replaceStr, content)
	} else {
		// append
		content += replaceStr
	}

	// Update SNI configuration
	if err := os.WriteFile(c.DovecotSslConf, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write dovecot config: %v", err)
	}

	return nil
}

// updateConfigLine updates a configuration line in config file
func (c *Certificate) updateConfigLine(config, key, value string) string {
	lines := strings.Split(config, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key) && strings.Contains(line, "=") {
			lines[i] = key + " = " + value
			return strings.Join(lines, "\n")
		}
	}
	return config + "\n" + key + " = " + value
}

// restartPostfix restarts Postfix service
func (c *Certificate) restartPostfix() error {
	// Restart Postfix container
	if err := c.dockerApiClient().RestartContainerByName(context.Background(), "billionmail-postfix-billionmail-1"); err != nil {
		return fmt.Errorf("failed to restart postfix container: %v", err)
	}

	return nil
}

// restartDovecot restarts Dovecot service
func (c *Certificate) restartDovecot() error {
	// Restart Dovecot container
	if err := c.dockerApiClient().RestartContainerByName(context.Background(), "billionmail-dovecot-billionmail-1"); err != nil {
		return fmt.Errorf("failed to restart dovecot container: %v", err)
	}

	return nil
}

// GetSSLStatus checks SSL certificate status
func (c *Certificate) GetSSLStatus(domain string) (bool, error) {
	// Check Postfix certificate
	csrPath := filepath.Join(consts.SSL_PATH, domain, "/fullchain.pem")
	ketPath := filepath.Join(consts.SSL_PATH, domain, "/privkey.pem")

	return c.checkCertificateFiles(csrPath, ketPath), nil
}

// GetSSLInfo retrieves SSL certificate information
func (c *Certificate) GetSSLInfo(domain string) (certInfo v1.CertInfo, err error) {
	// Check Postfix certificate
	csrPath := filepath.Join(consts.SSL_PATH, domain, "/fullchain.pem")
	keyPath := filepath.Join(consts.SSL_PATH, domain, "/privkey.pem")

	if !c.checkCertificateFiles(csrPath, keyPath) {
		err = fmt.Errorf("certificate files do not exist")
		return
	}

	crtPem, err := public.ReadFile(csrPath)
	if err != nil {
		return
	}

	// Get certificate information
	err = gconv.Struct(acme.GetCertInfo(crtPem), &certInfo)

	if err == nil {
		certInfo.CertPem = crtPem
		certInfo.KeyPem, err = public.ReadFile(keyPath)
		if err != nil {
			return
		}
	}

	return
}

// checkCertificateFiles verifies certificate files exist
func (c *Certificate) checkCertificateFiles(certPath, keyPath string) bool {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false
	}

	return true
}
