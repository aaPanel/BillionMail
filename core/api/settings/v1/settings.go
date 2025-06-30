package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

// SSLConfig
type SSLConfig struct {
	CertPath          string   `json:"cert_path" dc:"cert path"`
	KeyPath           string   `json:"key_path" dc:"key path"`
	CertData          string   `json:"certPem" dc:"cert data"`
	KeyData           string   `json:"privateKey" dc:"key data"`
	ValidUntil        string   `json:"notAfter" dc:"cert not after"`
	ValidFrom         string   `json:"notBefore" dc:"cert not before"`
	Issuer            string   `json:"issuer" dc:"cert issuer"`
	Subject           string   `json:"subject" dc:"cert subject"`
	DNSNames          []string `json:"dns_names" dc:"dns names"`
	SerialNumber      string   `json:"serial_number" dc:"serial number"`
	IsCA              bool     `json:"is_ca" dc:"is ca certificate"`
	IssuerCommonName  string   `json:"issuer_common_name" dc:"issuer common name"`
	SubjectCommonName string   `json:"subject_common_name" dc:"subject common name"`
	Version           int      `json:"version" dc:"version"`
	Status            bool     `json:"status" dc:"ssl certificate status"`
}

// SystemConfig
type SystemConfig struct {
	ServerIP string `json:"server_ip" dc:"server ip"`
	// Basic configuration
	AdminUsername string `json:"admin_username" dc:"admin username"`
	AdminPassword string `json:"admin_password" dc:"admin password"`
	SafePath      string `json:"safe_path" dc:"safe path"`

	// Domain name configuration
	Hostname string `json:"billionmail_hostname" dc:"billionmail hostname"`

	// Database configuration
	DBName string `json:"db_name" dc:"db name"`
	DBUser string `json:"db_user" dc:"db user"`
	DBPass string `json:"db_pass" dc:"db password"`

	// Redis configuration
	RedisPass string `json:"redis_pass" dc:"redis password"`
	RedisPort string `json:"redis_port" dc:"redis port"`

	// Port configuration
	MailPorts struct {
		SMTP       int `json:"smtp" dc:"smtp port"`
		SMTPS      int `json:"smtps" dc:"smtps port"`
		Submission int `json:"submission" dc:"submission port"`
		IMAP       int `json:"imap" dc:"imap port"`
		IMAPS      int `json:"imaps" dc:"imaps port"`
		POP        int `json:"pop" dc:"pop port"`
		POPS       int `json:"pops" dc:"pops port"`
	} `json:"mail_ports" dc:"mail ports configuration"`

	// Management port configuration
	ManagePorts struct {
		HTTP     int    `json:"http" dc:"http port"`
		HTTPS    int    `json:"https" dc:"https port"`
		Command1 string `json:"command_https" dc:"command https"`
		Command2 string `json:"command_http" dc:"command http"`
	} `json:"manage_ports" dc:"management ports configuration"`

	// Time zone configuration
	ManageTimeZone struct {
		TimeZone string `json:"timezone" dc:"time zone"`
		Command  string `json:"command" dc:"command"`
	} `json:"manage_timezone" dc:"time zone configuration"`

	// IPv4 network configuration
	IPv4Network string `json:"ipv4_network" dc:"ipv4 network"`
	Fail2ban    bool   `json:"fail2ban" dc:"fail2ban"`

	// SSL certificate configuration
	SSL SSLConfig `json:"ssl" dc:"ssl certificate configuration"`

	// IP whitelist configuration
	IPWhitelist        []string `json:"ip_whitelist" dc:"ip whitelist"`
	IPWhitelistEnabled bool     `json:"ip_whitelist_enable" dc:"ip whitelist enabled"`
}

type GetVersionReq struct {
	g.Meta `path:"/settings/get_version" tags:"Version" method:"get" summary:"Get current version information" in:"query"`
}

type GetVersionRes struct {
	api_v1.StandardRes
	Data struct {
		Version       string `json:"version" dc:"Version"`
		LatestVersion string `json:"latest_version" dc:"Latest version"`
		ReleaseDate   string `json:"release_date" dc:"Release date"`
		CanUpdate     bool   `json:"can_update" dc:"CanUpdate"`
	} `json:"data" dc:"Data"`
}

// Get system configuration request
type GetSystemConfigReq struct {
	g.Meta `path:"/settings/get_system_config" tags:"Settings" method:"get" summary:"Get system configuration"`
}

// Get system configuration response
type GetSystemConfigRes struct {
	api_v1.StandardRes
	Data *SystemConfig `json:"data" dc:"system configuration data"`
}

// Set system configuration request
type SetSystemConfigReq struct {
	g.Meta `path:"/settings/set_system_config" tags:"Settings" method:"post" summary:"Set system configuration"`
	SystemConfig
}

// Set system configuration response
type SetSystemConfigRes struct {
	api_v1.StandardRes
}

// Set system configuration request
type SetSystemConfigKeyReq struct {
	g.Meta `path:"/settings/set_system_config_key" tags:"Settings" method:"post" summary:"Set system configuration"`
	Key    string `json:"key" dc:"configuration key"`
	Value  string `json:"value" dc:"configuration value"`
}

type SetSystemConfigKeyRes struct {
	api_v1.StandardRes
}

// Set SSL certificate request
type SetSSLConfigReq struct {
	g.Meta   `path:"/settings/set_ssl_config" tags:"Settings" method:"post" summary:"Set SSL certificate configuration"`
	CertData string `json:"certPem" dc:"certificate data"`
	KeyData  string `json:"privateKey" dc:"private key data"`
}

type SetSSLConfigRes struct {
	api_v1.StandardRes
}

type GetTimeZoneListReq struct {
	g.Meta `path:"/settings/get_timezone_list" tags:"Settings" method:"get" summary:"Get available time zones"`
}
type GetTimeZoneListRes struct {
	api_v1.StandardRes
	Data map[string][]string `json:"data" dc:"available time zones"`
}

type SetIPWhitelistReq struct {
	g.Meta `path:"/settings/set_ip_whitelist" tags:"Settings" method:"post" summary:"Set IP whitelist"`
	IPList []string `json:"ip_list" dc:"IP list"`
}

type SetIPWhitelistRes struct {
	api_v1.StandardRes
}
