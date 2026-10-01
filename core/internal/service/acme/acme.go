package acme

import (
	v1 "billionmail-core/api/domains/v1"
	"billionmail-core/internal/service/public"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	srcLog "log"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/azuredns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/cloudxns"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

/**
 * @description: Get user information (reuses persisted account key if available)
 * @param {string} email Email address
 * @return {*MyUser} User information
 */
func GetMyUser(ctx context.Context, email string) (*MyUser, error) {
	// Try to load existing account key from disk
	accountKeyPath := filepath.Join(public.ROOT_PATH, "core", "data", "acme", "account.key")
	var privateKey *ecdsa.PrivateKey
	var err error

	if public.FileExists(accountKeyPath) {
		keyBytes, readErr := public.ReadFile(accountKeyPath)
		if readErr == nil {
			block, _ := pem.Decode([]byte(keyBytes))
			if block != nil {
				parsedKey, parseErr := x509.ParseECPrivateKey(block.Bytes)
				if parseErr == nil {
					privateKey = parsedKey
				}
			}
		}
	}

	if privateKey == nil {
		// Generate new private key if none exists
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, errors.New(public.LangCtx(ctx, "Failed to generate user private key: {}", err.Error()))
		}

		// Persist the key for future use
		keyBytes, marshalErr := x509.MarshalECPrivateKey(privateKey)
		if marshalErr == nil {
			pemBlock := &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: keyBytes,
			}
			pemBytes := pem.EncodeToMemory(pemBlock)

			dir := filepath.Dir(accountKeyPath)
			if !public.FileExists(dir) {
				os.MkdirAll(dir, 0750)
			}
			public.WriteFile(accountKeyPath, string(pemBytes))
		}
	}

	// Initialize user
	myUser := MyUser{
		Email: email,
		key:   privateKey,
	}
	return &myUser, nil
}

/**
 * @description: Get client configuration
 * @param {*MyUser} myUser User information
 * @return {*lego.Config} Configuration information
 */
func GetConfig(myUser *MyUser) *lego.Config {
	// Initialize configuration
	config := lego.NewConfig(myUser)

	// Set CA directory URL
	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"

	// Set key type
	config.Certificate.KeyType = certcrypto.RSA2048

	return config
}

/**
 * @description: Configure DNS verification via Tencent Cloud
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsTencentcloud(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["SecretId"] == "" || keyConfig["SecretKey"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: SecretId or SecretKey is empty in TencentCloud configuration file"))
	}

	cfg := tencentcloud.NewDefaultConfig()
	cfg.SecretID = keyConfig["SecretId"]
	cfg.SecretKey = keyConfig["SecretKey"]

	p, err := tencentcloud.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}
	return nil
}

/**
 * @description: Configure DNS verification via Alibaba Cloud
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsAliyun(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["APIKey"] == "" || keyConfig["SecretKey"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: APIKey or SecretKey is empty in AliDNS configuration file"))
	}

	cfg := alidns.NewDefaultConfig()
	cfg.APIKey = keyConfig["APIKey"]
	cfg.SecretKey = keyConfig["SecretKey"]

	p, err := alidns.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}

	return nil
}

/**
 * @description: Configure DNS verification via Cloudflare
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsCloudflare(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["APIKey"] == "" || keyConfig["Email"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: APIKey or Email is empty in Cloudflare configuration file"))
	}

	cfg := cloudflare.NewDefaultConfig()
	cfg.AuthEmail = keyConfig["Email"]
	cfg.AuthKey = keyConfig["APIKey"]

	p, err := cloudflare.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}

	return nil
}

/**
 * @description: Configure DNS verification via CloudXNS
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsCloudxns(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["APIKey"] == "" || keyConfig["SecretKey"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: APIKey or SecretKey is empty in CloudXNS configuration file"))
	}

	cfg := cloudxns.NewDefaultConfig()
	cfg.APIKey = keyConfig["APIKey"]
	cfg.SecretKey = keyConfig["SecretKey"]

	p, err := cloudxns.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}

	return nil
}

/**
 * @description: Configure DNS verification via Azure DNS
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsAzuredns(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["ClientID"] == "" || keyConfig["ClientSecret"] == "" || keyConfig["TenantID"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: ClientID, ClientSecret or TenantID is empty in AzureDNS configuration file"))
	}

	cfg := azuredns.NewDefaultConfig()
	cfg.ClientID = keyConfig["ClientID"]
	cfg.ClientSecret = keyConfig["ClientSecret"]
	cfg.TenantID = keyConfig["TenantID"]

	p, err := azuredns.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}

	return nil
}

/**
 * @description: Configure DNS verification via GoDaddy
 * @param {*lego.Client} client Client
 * @param {map[string]string} keyConfig Configuration information
 * @return error Error information
 */
func SetDnsGodaddy(ctx context.Context, client *lego.Client, keyConfig map[string]string) error {
	if keyConfig == nil || keyConfig["APIKey"] == "" || keyConfig["APISecret"] == "" {
		return errors.New(public.LangCtx(ctx, "DNS automated resolution failed: APIKey or APISecret is empty in Godaddy configuration file"))
	}

	cfg := godaddy.NewDefaultConfig()
	cfg.APIKey = keyConfig["APIKey"]
	cfg.APISecret = keyConfig["APISecret"]

	p, err := godaddy.NewDNSProviderConfig(cfg)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS provider initialization failed: {}", err.Error()))
	}

	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return errors.New(public.LangCtx(ctx, "DNS verification setup failed: {}", err.Error()))
	}

	return nil
}

/**
 * @description: Get ACME log file path
 * @param {context.Context} ctx Context
 * @return {string} Log file path
 */
func GetAcmeLogFile(ctx context.Context) string {
	logPath := public.AbsPath(fmt.Sprintf("logs/%s", public.GetUserName(ctx)))
	if !public.FileExists(logPath) {
		_ = os.Mkdir(logPath, 0750)
	}
	return fmt.Sprintf("%s/acme.log", logPath)
}

/**
 * @description: Get ACME log content
 * @param {context.Context} ctx Context
 * @return {string} Log content
 */
func GetAcmeLogBody(ctx context.Context) string {
	logFile := GetAcmeLogFile(ctx)
	body, err := public.ReadFile(logFile)
	if err != nil {
		return ""
	}
	return body
}

/**
 * @description: Get log file object
 * @param {context.Context} ctx Context
 * @return {*os.File} Log file
 */
func GetLogFile(ctx context.Context) *os.File {
	logFile := GetAcmeLogFile(ctx)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil
	}
	username := public.GetUserName(ctx)
	uid, gid := public.GetUidAndGid(username)
	file.Chown(uid, gid)
	file.Truncate(0)
	file.Seek(0, 0)
	return file
}

/**
 * @description: Configure logging
 * @param {*os.File} logFile Log file
 */
func SetLog(logFile *os.File) {
	if logFile == nil {
		return
	}
	log.Logger = srcLog.New(logFile, "", 0)
}

/**
 * @description: Close log file
 * @param {*os.File} logFile Log file
 */
func CloseLog(logFile *os.File) {
	if logFile == nil {
		return
	}
	log.Logger = srcLog.New(os.Stderr, "", srcLog.LstdFlags)
	logFile.Close()
}

/**
 * @brief Apply for SSL certificate with an existing HTTP server
 * @param domains Domain list
 * @param email Email address
 * @param vtype Verification type: http/dns
 * @param dnsProvider DNS provider for DNS verification
 * @param dnsProviderToken DNS provider credentials
 * @param savePath Path to save certificate files
 * @return certificate, privateKey, error
 */
func ApplySSLWithExistingServer(ctx context.Context, domains []string, email string, vtype string,
	dnsProvider string, dnsProviderToken map[string]string, savePath string) (string, string, error) {

	// Set up logging
	logFile := GetLogFile(ctx)
	SetLog(logFile)
	defer CloseLog(logFile)

	// Get user information
	myUser, err := GetMyUser(ctx, email)
	if err != nil {
		return "", "", err
	}

	// Get configuration
	config := GetConfig(myUser)
	if config == nil {
		return "", "", errors.New(public.LangCtx(ctx, "Failed to get configuration"))
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return "", "", errors.New(public.LangCtx(ctx, "Failed to create ACME client: {}", err.Error()))
	}

	// Set verification method
	if vtype == "http" {
		// Assume the HTTP server is already running and properly configured
		// to handle the challenge requests
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("127.0.0.1", "60880"))
		if err != nil {
			return "", "", errors.New(public.LangCtx(ctx, "Failed to set HTTP verification: {}", err.Error()))
		}
	} else if vtype == "dns" && dnsProvider != "" {
		// Set DNS verification - same as in the standard ApplySSL function
		switch dnsProvider {
		case "tencentcloud":
			err = SetDnsTencentcloud(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set Tencent Cloud DNS verification: {}", err.Error()))
			}
		case "alidns":
			err = SetDnsAliyun(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set Alibaba Cloud DNS verification: {}", err.Error()))
			}
		case "cloudxns":
			err = SetDnsCloudxns(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set CloudXNS DNS verification: {}", err.Error()))
			}
		case "azuredns":
			err = SetDnsAzuredns(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set AzureDNS verification: {}", err.Error()))
			}
		case "cloudflare":
			err = SetDnsCloudflare(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set Cloudflare DNS verification: {}", err.Error()))
			}
		case "godaddy":
			err = SetDnsGodaddy(ctx, client, dnsProviderToken)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to set Godaddy DNS verification: {}", err.Error()))
			}
		default:
			return "", "", errors.New(public.LangCtx(ctx, "Unsupported DNS provider: {}", dnsProvider))
		}
	}

	// Register or query existing user on ACME server
	var reg *registration.Resource
	// Try to query existing registration first (same key = same account)
	reg, err = client.Registration.QueryRegistration()
	if err != nil || reg == nil {
		// No existing registration, register new account
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return "", "", errors.New(public.LangCtx(ctx, "Failed to register user: {}", err.Error()))
		}
	}

	// Save user information
	myUser.Registration = reg

	// Submit application
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	// Get certificate
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return "", "", errors.New(public.LangCtx(ctx, "Failed to apply for SSL certificate: {}", err.Error()))
	}

	// Save certificate files if path is provided
	if savePath != "" {
		// Create directory if it doesn't exist
		if !public.FileExists(savePath) {
			err = os.MkdirAll(savePath, 0750)
			if err != nil {
				return "", "", errors.New(public.LangCtx(ctx, "Failed to create directory: {}", err.Error()))
			}
		}

		// Save certificate and private key files
		certificateFile := filepath.Join(savePath, "certificate.pem")
		privateKeyFile := filepath.Join(savePath, "private_key.pem")

		_, err = public.WriteFile(certificateFile, string(certificates.Certificate))
		if err != nil {
			return "", "", errors.New(public.LangCtx(ctx, "Failed to save certificate file: {}", err.Error()))
		}

		_, err = public.WriteFile(privateKeyFile, string(certificates.PrivateKey))
		if err != nil {
			return "", "", errors.New(public.LangCtx(ctx, "Failed to save private key file: {}", err.Error()))
		}
	}

	// Return certificate
	return string(certificates.Certificate), string(certificates.PrivateKey), nil
}

type CertInfo v1.CertInfo

/**
 * @description: Get certificate information
 * @param {string} certificateStr Certificate string
 * @return {CertInfo} Certificate information
 */
func GetCertInfo(certificateStr string) CertInfo {
	certInfo := CertInfo{}
	block, _ := pem.Decode([]byte(certificateStr))
	if block == nil {
		return certInfo
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		fmt.Println(err.Error())
		return certInfo
	}

	certInfo.Subject = cert.Subject.CommonName
	certInfo.Issuer = cert.Issuer.CommonName
	certInfo.NotBefore = cert.NotBefore.Format("2006-01-02")
	certInfo.NotAfter = cert.NotAfter.Format("2006-01-02")
	certInfo.DNSNames = cert.DNSNames
	certInfo.Endtime = int(cert.NotAfter.Unix())
	return certInfo
}
