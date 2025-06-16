export interface NetworkSettings {
	domain: string
	sslEnabled: boolean
	sslCert: string
	sslKey: string
}

export interface SystemSettings {
	port: number
	timezone: string
	currentTime: string
}

export interface PasswordStrength {
	level: 'weak' | 'medium' | 'strong'
	score: number
}

export interface SystemInfo {
	admin_username: string
	billionmail_hostname: string
	safe_path: string
	db_name: string
	db_user: string
	db_pass: string
	redis_pass: string
	redis_port: string
	ipv4_network: string
	fail2ban: boolean
	manage_ports: {
		http: number
		https: number
		command_https: string
		command_http: string
	}
	manage_timezone: {
		timezone: string
		command: string
	}
	ssl: {
		cert_path: string
		key_path: string
		certPem: string
		privateKey: string
		notAfter: string
		notBefore: string
		issuer: string
		subject: string
		dns_names: any
		serial_number: string
		is_ca: boolean
		issuer_common_name: string
		subject_common_name: string
		version: number
		status: boolean
	}
}
