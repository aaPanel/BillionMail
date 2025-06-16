package public

import (
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Regular expression instances
	reMap = map[string]*regexp.Regexp{
		"dotAndNumber":           regexp.MustCompile(`^(?:\d+|\.)+$`),
		"host":                   regexp.MustCompile(`^[\w\-]+(?:\.[\w\-]+)+(?::\d+)?$`),
		"email":                  regexp.MustCompile(`^.+@\[?[\w\-.]+\.(?:[a-zA-Z]{2,}|\d{1,3})\]?$`),
		"email_new":              regexp.MustCompile(`^.+@\[?[\w\-.]+\.(?:[a-zA-Z]{2,}|\d{1,3})\]?$`),
		"admin_path":             regexp.MustCompile(`^[\w\/\-\.]+$`),
		"chinese":                regexp.MustCompile(`[\p{Han}]`),
		"double_byte":            regexp.MustCompile(`[^\x00-\xff]`),
		"base63":                 regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
		"safe_path":              regexp.MustCompile(`^[\w.\-/]+$`),
		"simple_chars":           regexp.MustCompile(`^[\w.\-]+$`),
		"alphanum_dash":          regexp.MustCompile(`^\w+$`),
		"general_version_format": regexp.MustCompile(`^\d+(?:\.\d+){1,2}$`),
		"md5_hash":               regexp.MustCompile(`^[a-fA-F0-9]{32}$`),
		"numeric":                regexp.MustCompile(`^\d+(?:\.\d+)?$`),
		"integer":                regexp.MustCompile(`^\d+$`),
		"password_1":             regexp.MustCompile(`^[\w!@#$%&;\-\+=\(\)]+$`),
		"password_2":             regexp.MustCompile(`^[\w!@#$%&+\-\(\)]+$`),
		"fileCheck":              regexp.MustCompile(`^[\/\w\.\-\+\(\)\*_=#@!]+$`),
		"database_access_format": regexp.MustCompile(`^[0-9a-fA-F%:.]+$`),
		"api_path":               regexp.MustCompile(`^(/\w+)+$`),
		"domain":                 regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`),
	}
)

// Check if it is an IP address
func IsIpAddr(ip string) bool {
	ipAddr := net.ParseIP(ip)
	return ipAddr != nil
}

// Check if it is an IPv4 address
func IsIpv4(ip string) bool {
	return IsIpAddr(ip) && strings.Contains(ip, ".")
}

// Check if it is an IPv6 address
func IsIpv6(ip string) bool {
	return IsIpAddr(ip) && strings.Contains(ip, ":")
}

// Check if the domain is valid
func IsDomain(s string) bool {
	return reMap["domain"].MatchString(s)
}

// IsHost checks if the string is a host address
func IsHost(s string) bool {
	if s == "localhost" {
		return true
	}

	if reMap["host"].MatchString(s) {
		return true
	}

	if ipAddr := net.ParseIP(s); ipAddr != nil {
		return true
	}

	return false
}

// IsPort checks if the string is a port number
func IsPort(s string) bool {
	port, err := strconv.Atoi(s)

	if err != nil {
		return false
	}

	if port < 1 || port > 65535 {
		return false
	}

	return true
}

// IsEmail checks if the string is an Email
func IsEmail(s string) bool {
	return reMap["email"].MatchString(s)
}

// IsEmailNew checks if the string is an Email
func IsEmailNew(s string) bool {
	return reMap["email_new"].MatchString(s)
}

// IsUrl checks if the string is a URL
func IsUrl(s string) bool {
	_, err := url.ParseRequestURI(s)

	if err != nil {
		return false
	}

	u, err := url.Parse(s)

	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// IsAdminPath checks secure entry format
func IsAdminPath(s string) bool {
	return reMap["admin_path"].MatchString(s)
}

// HasChinese matches Chinese characters (contains Chinese characters)
func HasChinese(s string) bool {
	return reMap["chinese"].MatchString(s)
}

// HasDouble matches double-byte characters (contains Chinese characters)
func HasDouble(s string) bool {
	return reMap["double_byte"].MatchString(s)
}

// IsBase63 matches numbers, letters, and underscores
func IsBase63(s string) bool {
	return reMap["base63"].MatchString(s)
}

// IsSafePath matches safe paths
func IsSafePath(s string) bool {
	return reMap["safe_path"].MatchString(s)
}

// IsSimpleChars matches simple characters
func IsSimpleChars(s string) bool {
	return reMap["simple_chars"].MatchString(s)
}

// IsGeneralVersionFormat matches general version format
func IsGeneralVersionFormat(s string) bool {
	return reMap["general_version_format"].MatchString(s)
}

// IsMd5Hash matches MD5 hash
func IsMd5Hash(s string) bool {
	return reMap["md5_hash"].MatchString(s)
}

// IsAlphanumDash matches alphanumeric characters and underscores
func IsAlphanumDash(s string) bool {
	return reMap["alphanum_dash"].MatchString(s)
}

// IsNumeric matches numbers
func IsNumeric(s string) bool {
	return reMap["numeric"].MatchString(s)
}

// IsInteger matches integers
func IsInteger(s string) bool {
	return reMap["integer"].MatchString(s)
}

// IsPassword1 matches password
func IsPassword1(s string) bool {
	return reMap["password_1"].MatchString(s)
}

// Compressed password matching rule
func IsPassword2(s string) bool {
	return reMap["password_2"].MatchString(s)
}

// IsDatabaseAccessFormat matches database access format
func IsDatabaseAccessFormat(s string) bool {
	return reMap["database_access_format"].MatchString(s)
}

// Compressed password matching rule
func FileCheck(s string) bool {
	return reMap["fileCheck"].MatchString(s)
}

// IsApiPath matches API paths
func IsApiPath(s string) bool {
	return reMap["api_path"].MatchString(s)
}

// IsValidUsername
// Rules: 4-32 bit characters, only allowed letters, numbers, underscores
func IsValidUsername(username string) bool {
	pattern := `^[a-zA-Z0-9_]{4,32}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// IsValidHostname
// Rules: DNS host name format, allowed letters, numbers, hyphens and dots, but cannot start or end with a hyphen
func IsValidHostname(hostname string) bool {
	// Check total length
	if len(hostname) > 255 {
		return false
	}

	// Split into labels for checking
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		// Each label rule:
		// 1. Length between 1-63
		// 2. Only letters, numbers and hyphens
		// 3. Cannot start or end with a hyphen
		if len(label) < 1 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		pattern := `^[a-zA-Z0-9-]+$`
		matched, _ := regexp.MatchString(pattern, label)
		if !matched {
			return false
		}
	}
	return true
}

// IsValidCIDR Check if it is a valid CIDR format
func IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IsValidTimezone Check if it is a valid timezone
func IsValidTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}

// ContainsDangerousChars Check if it contains dangerous characters
func ContainsDangerousChars(value string) bool {
	dangerousChars := []string{
		";", "|", "&", "$", "<", ">", "`", "\\",
		"(", ")", "{", "}", "[", "]", "!", "#",
		"\x00",
	}

	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return true
		}
	}

	// Check if it contains control characters (ASCII 0-31, except for common tab, newline, etc.)
	for _, r := range value {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}

	return false
}

// ParseInt Safely convert string to integer
func ParseInt(s string) int {
	// Remove whitespace
	s = strings.TrimSpace(s)

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
