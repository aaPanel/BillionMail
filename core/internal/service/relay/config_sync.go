package relay

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	domainsV1 "billionmail-core/api/domains/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model/entity"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"

	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

var (
	postfixConfigDir           = public.AbsPath("../conf/postfix")
	saslPasswdFile             = "/conf/sasl_passwd"
	mainCfFile                 = "main.cf"
	postfixContainerName       = consts.SERVICES.Postfix
	senderRelayMapsPrimaryFile = "/conf/sender_relay_maps_primary"
	saslPasswdPrimaryFile      = "/conf/sasl_passwd_primary"
)

func GetRelayEncryptionKey() (string, error) {
	// 1. Retrieve key from the database
	val, err := g.DB().Model("bm_options").
		Where("name", "relay_encryption_key").
		Value("value")

	if val != nil && val.String() != "" {
		if _, err := hex.DecodeString(val.String()); err == nil {
			return val.String(), nil
		}
	}

	// 2. Generate key
	newSecret := hex.EncodeToString(grand.B(16))

	// 3. Save the new key to the database
	_, err = g.DB().Model("bm_options").
		OnConflict("name").
		OnDuplicate("value").
		Data(g.Map{
			"name":  "relay_encryption_key",
			"value": newSecret,
		}).
		Save()
	if err != nil {
		// If insert fails, attempt to retrieve the key again
		val, err = g.DB().Model("bm_options").
			Where("name", "relay_encryption_key").
			Value("value")

		if val != nil && val.String() != "" {
			return val.String(), nil
		}
		return "", gerror.New("Failed to insert new key and retrieve key again")
	}

	return newSecret, nil
}

func EncryptPassword(ctx context.Context, plainText string) (string, error) {
	if plainText == "" {
		return "", gerror.New("Password cannot be empty")
	}
	relayEncryptionKey, err := GetRelayEncryptionKey()
	if err != nil {
		return "", gerror.Wrap(err, "Failed to retrieve encryption key")
	}

	keyBytes, err := hex.DecodeString(relayEncryptionKey)
	if err != nil {
		return "", gerror.Wrap(err, "Failed to parse encryption key")
	}

	if len(keyBytes) < 16 {
		return "", gerror.New("Encryption key length is insufficient")
	}
	keyBytes = keyBytes[:16]

	encrypted, err := gaes.Encrypt([]byte(plainText), keyBytes)
	if err != nil {
		//g.Log().Errorf(ctx, "Password encryption failed: %v", err)
		return "", gerror.Wrap(err, "Password encryption failed")
	}

	return hex.EncodeToString(encrypted), nil
}

func DecryptPassword(ctx context.Context, encryptedHex string) (string, error) {
	if encryptedHex == "" {
		return "", nil
	}

	encryptedBytes, err := hex.DecodeString(encryptedHex)
	if err != nil {
		//g.Log().Errorf(ctx, "Decryption failed, invalid hex format: %v", err)
		return "", gerror.Wrap(err, "Password format is incorrect")
	}

	relayEncryptionKey, err := GetRelayEncryptionKey()
	if err != nil {
		return "", gerror.Wrap(err, "Failed to retrieve encryption key")
	}

	keyBytes, err := hex.DecodeString(relayEncryptionKey)
	if err != nil {
		return "", gerror.Wrap(err, "Failed to parse encryption key")
	}

	if len(keyBytes) < 16 {
		return "", gerror.New("Encryption key length is insufficient")
	}
	keyBytes = keyBytes[:16]

	decrypted, err := gaes.Decrypt(encryptedBytes, keyBytes)
	if err != nil {
		//g.Log().Errorf(ctx, "Password decryption failed: %v", err)
		return "", gerror.Wrap(err, "Password decryption failed")
	}

	return string(decrypted), nil
}

func buildCacheKey(domain, recordType string) string {
	return fmt.Sprintf("DOMAIN_DNS_RECORDS_:%s:_%s", domain, recordType)
}

// GenerateSPFRecord generates SPF record hints based on IP and host.
func GenerateSPFRecord(ip string, host string, domain string) string {
	// Retrieve the current SPF record for the domain
	var existingValue string
	var newSpfParts []string

	// Fetch SPF record from cache or domains service
	spfRecord := public.GetCache(buildCacheKey(domain, "SPF"))

	if spfRecord != nil {
		if v, ok := spfRecord.(domainsV1.DNSRecord); ok {
			existingValue = v.Value
		}
	}

	// If cache is empty, attempt to fetch from domains service
	if existingValue == "" {
		record, _ := domains.GetSPFRecord(domain, false)
		existingValue = record.Value
	}

	// Create a basic record if none exists
	if existingValue == "" {
		existingValue = "v=spf1 +a +mx"
	}

	// Analyze the existing record to avoid duplications
	parts := strings.Split(existingValue, " ")
	existingParts := make(map[string]bool)
	var allMechanism string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		existingParts[part] = true
		if strings.HasSuffix(part, "all") {
			allMechanism = part
			continue
		}

		if !strings.HasPrefix(part, "include:") && !strings.HasPrefix(part, "ip4:") && !strings.HasPrefix(part, "ip6:") {
			newSpfParts = append(newSpfParts, part)
		}
	}
	// Add IP if provided and not duplicated
	if ip != "" {
		var ipPart string
		if strings.Contains(ip, ":") {
			ipPart = fmt.Sprintf("+ip6:%s", ip)
		} else {
			ipPart = fmt.Sprintf("+ip4:%s", ip)
		}
		if !existingParts[ipPart] {
			newSpfParts = append(newSpfParts, ipPart)
		}
	}

	// Add host if provided and not duplicated
	if host != "" {
		hostPart := fmt.Sprintf("include:%s", host)
		if !existingParts[hostPart] {
			newSpfParts = append(newSpfParts, hostPart)
		}
	}

	if allMechanism != "" {
		newSpfParts = append(newSpfParts, allMechanism)
	} else {
		newSpfParts = append(newSpfParts, "~all")
	}

	return strings.Join(newSpfParts, " ")
}

func CheckRelayFirstSync(ctx context.Context) {
	// Change the marker file and ensure the latest synchronization
	markPath := public.AbsPath("../core/data/SMTP_relay_sync_config_mark2.pl")
	if gfile.Exists(markPath) {
		return
	}

	// Remove the old configuration
	beginMarker := "# BEGIN RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"
	endMarker := "# END RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"
	mainCfPath := path.Join(postfixConfigDir, mainCfFile)

	if gfile.Exists(mainCfPath) {
		content := gfile.GetContents(mainCfPath)

		// Comment out the old "sender_dependent_default_transport_maps" configuration
		content = commentOutOldTransportMaps(content)

		// Reply with the updated content
		if err := gfile.PutContents(mainCfPath, content); err != nil {
			g.Log().Errorf(ctx, "Failed to update main.cf with commented transport maps: %v", err)
		} else {
			g.Log().Info(ctx, "Successfully commented out old transport maps configuration")
		}

		beginIndex := strings.Index(content, beginMarker)
		endIndex := strings.Index(content, endMarker)

		// Check if the configuration block exists
		if beginIndex != -1 && endIndex != -1 && beginIndex < endIndex {
			// Find configuration block and delete
			blockStart := beginIndex
			blockEnd := endIndex + len(endMarker)

			if blockEnd < len(content) && content[blockEnd] == '\n' {
				blockEnd++
			}

			// Delete the configuration block
			newContent := content[:blockStart]
			if blockEnd < len(content) {
				newContent += content[blockEnd:]
			}

			if err := gfile.PutContents(mainCfPath, newContent); err != nil {
				g.Log().Errorf(ctx, "Failed to delete the old relay configuration block in main.cf: %v", err)
			} else {
				g.Log().Info(ctx, "Successfully deleted the old relay configuration block in the main.cf file")
			}
		} else {
			g.Log().Debug(ctx, "No need to delete.")
		}
	} else {
		g.Log().Warning(ctx, "The main.cf file does not exist")
	}

	err := SyncRelayConfigsToPostfix(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to sync relay configurations on first sync:  %v", err)
		return
	}

	// Check if there are beginMarker and endMarker after synchronization. If not, add them forcibly
	if gfile.Exists(mainCfPath) {
		content := gfile.GetContents(mainCfPath)
		beginIndex := strings.Index(content, beginMarker)
		endIndex := strings.Index(content, endMarker)
		hasConfigBlock := beginIndex != -1 && endIndex != -1 && beginIndex < endIndex

		if !hasConfigBlock {
			g.Log().Warning(ctx, "Relay configuration block not found after sync, adding it forcefully")

			configBlock := fmt.Sprintf(`%s
sender_dependent_relayhost_maps = hash:/etc/postfix/conf/sender_relay_maps_primary, pgsql:/etc/postfix/sql/pgsql_sender_relay_maps.cf
smtp_sasl_password_maps = hash:/etc/postfix/conf/sasl_passwd_primary, hash:/etc/postfix/conf/sasl_passwd
sender_dependent_default_transport_maps = pgsql:/etc/postfix/sql/pgsql_sender_transport_maps.cf
%s`, beginMarker, endMarker)

			// Add the configuration block
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n" + configBlock + "\n"

			if err := gfile.PutContents(mainCfPath, content); err != nil {
				g.Log().Errorf(ctx, "Failed to add relay configuration block to main.cf: %v", err)
			} else {
				g.Log().Info(ctx, "Successfully added relay configuration block to main.cf")
			}
		} else {
			g.Log().Debug(ctx, "Relay configuration block exists after sync")
		}

		// check hash
		senderRelayMapsPrimaryPath := path.Join(postfixConfigDir, senderRelayMapsPrimaryFile)
		saslPasswdPrimaryPath := path.Join(postfixConfigDir, saslPasswdPrimaryFile)
		needReload := false
		if !gfile.Exists(senderRelayMapsPrimaryPath) {
			needReload = true
			if err := gfile.PutContents(senderRelayMapsPrimaryPath, ""); err != nil {
				g.Log().Errorf(ctx, "Failed to create %s: %v", senderRelayMapsPrimaryPath, err)
			}
		}
		if !gfile.Exists(saslPasswdPrimaryPath) {
			needReload = true
			if err := gfile.PutContents(saslPasswdPrimaryPath, ""); err != nil {
				g.Log().Errorf(ctx, "Failed to create %s: %v", saslPasswdPrimaryPath, err)
			}
		}
		if needReload {
			err = reloadPostfixConfigs2(ctx)
		}
	}

	if err := gfile.PutContents(markPath, fmt.Sprintf("First sync completed at %s", time.Now().Format("2006-01-02 15:04:05"))); err != nil {
		g.Log().Warningf(ctx, "Failed to create the synchronization marker file: %v", err)
	}
	return
}

func EnsurePostfixConfExists(ctx context.Context) {

	senderRelayMapsPrimaryPath := path.Join(postfixConfigDir, senderRelayMapsPrimaryFile)
	saslPasswdPrimaryPath := path.Join(postfixConfigDir, saslPasswdPrimaryFile)
	needReload := false

	if !gfile.Exists(senderRelayMapsPrimaryPath) {
		needReload = true
		if err := gfile.PutContents(senderRelayMapsPrimaryPath, "# Created on "+time.Now().Format("2006-01-02 15:04:05")+"\n# This file maps sender domains to relay hosts\n"); err != nil {
			g.Log().Errorf(ctx, "Failed to create %s: %v", senderRelayMapsPrimaryPath, err)
		} else {
			g.Log().Info(ctx, "Created sender_relay_maps_primary file in SyncRelayConfigsToPostfix")
		}
	}

	if !gfile.Exists(saslPasswdPrimaryPath) {
		needReload = true
		if err := gfile.PutContents(saslPasswdPrimaryPath, "# Created on "+time.Now().Format("2006-01-02 15:04:05")+"\n# This file contains SASL authentication credentials for relay hosts\n"); err != nil {
			g.Log().Errorf(ctx, "Failed to create %s: %v", saslPasswdPrimaryPath, err)
		} else {
			g.Log().Info(ctx, "Created sasl_passwd_primary file in SyncRelayConfigsToPostfix")
		}
	}

	if needReload {
		if err := reloadPostfixConfigs2(ctx); err != nil {
			g.Log().Errorf(ctx, "Failed to reload Postfix configs after creating necessary files: %v", err)

		} else {
			g.Log().Info(ctx, "Successfully initialized and reloaded Postfix configs")
		}
	}

}

func reloadPostfixConfigs2(ctx context.Context) error {

	dk, err := docker.NewDockerAPI()
	if err != nil {
		g.Log().Error(ctx, "Failed to connect to Docker API:", err)
		return gerror.Wrap(err, "Failed to connect to Docker service")
	}
	defer dk.Close()
	// List of commands to run
	cmdsToRun := [][]string{
		{"postmap", "/etc/postfix/conf/sender_relay_maps_primary"},
		{"postmap", "/etc/postfix/conf/sasl_passwd_primary"},
		{"postfix", "reload"},
	}
	// Execute commands
	for _, cmd := range cmdsToRun {
		cmdStr := strings.Join(cmd, " ")

		result, err := dk.ExecCommandByName(ctx, postfixContainerName, cmd, "root")

		if err != nil {
			return gerror.Newf("Failed to execute command: %v, Command: %s", err, cmdStr)
		}

		if result == nil {
			return gerror.Newf("Command execution result is empty: %s", cmdStr)
		}

		if result.ExitCode != 0 {

			return gerror.Newf("Command execution failed, Exit code: %d, Output: %s, Command: %s",
				result.ExitCode, result.Output, cmdStr)
		}

	}

	return nil
}

// SyncRelayConfigsToPostfix
func SyncRelayConfigsToPostfix(ctx context.Context) error {

	activeConfigCount, err := g.DB().Model("bm_relay_config").Where("active", 1).Count()
	if err != nil {
		return gerror.Wrap(err, "Failed to query active relay configurations count")
	}

	// Only check domain mappings that belong to ACTIVE relay configs.
	// Using total mapping count would incorrectly include mappings from disabled relays.
	activeMappingCount, err := g.DB().Model("bm_relay_domain_mapping rdm").
		LeftJoin("bm_relay_config rc", "rdm.relay_id = rc.id").
		Where("rc.active", 1).
		Count()
	if err != nil {
		return gerror.Wrap(err, "Failed to query active relay domain mappings count")
	}

	// Relay is considered "enabled" when there are active relay configs.
	// Domain mappings are optional — a relay may exist without bound sender domains.
	isActiveRelaySystem := activeConfigCount > 0
	g.Log().Infof(ctx, "Relay system status: Active config count=%d, Active domain mapping count=%d, System enabled=%v",
		activeConfigCount, activeMappingCount, isActiveRelaySystem)

	var activeConfigs []*entity.BmRelayConfig
	if isActiveRelaySystem {
		err = g.DB().Model("bm_relay_config").Where("active", 1).Scan(&activeConfigs)
		if err != nil {

			return gerror.Wrap(err, "Failed to query active relay configurations")
		}

		// Update SMTP service name mapping
		if err := updateSmtpServiceMappings(ctx, activeConfigs); err != nil {

			return err
		}

	} else {

		activeConfigs = []*entity.BmRelayConfig{}
		g.Log().Info(ctx, "No active relay to domain mappings, relay functionality will be disabled")

		// Clear stale relay transport mappings when no active relay exists
		result, err := g.DB().Model("bm_domain_smtp_transport").Where("atype", "relay").Delete()
		if err != nil {
			g.Log().Warningf(ctx, "Failed to clear stale SMTP transport mappings: %v", err)
		} else {
			if affected, _ := result.RowsAffected(); affected > 0 {
				g.Log().Infof(ctx, "Cleared %d stale SMTP transport mappings", affected)
			}
		}
	}

	// 1. Generate configuration files-pwd
	if err := generateRelayPwdFiles(ctx, activeConfigs); err != nil {
		return err
	}

	// 2. update master.cf
	if err := updatePostfixMasterCf(ctx, activeConfigs); err != nil {
		return err
	}

	// 3. Check whether to enable relay functionality and update pgsql query file
	if err := ensurePostfixRelayConfig(path.Join(postfixConfigDir, mainCfFile), isActiveRelaySystem); err != nil {
		return err
	}

	// 4. Rebuild DKIM signing config to exclude relay-mapped domains
	if err := domains.RepairDKIMSigningConfig(ctx); err != nil {
		g.Log().Warningf(ctx, "Failed to repair DKIM signing config after relay sync: %v", err)
	}

	return reloadPostfixConfigs(ctx)
}

// updatePostfixMasterCf Updates the Postfix master configuration file
func updatePostfixMasterCf(ctx context.Context, configs []*entity.BmRelayConfig) error {
	masterCfPath := path.Join(postfixConfigDir, "master.cf")
	if !gfile.Exists(masterCfPath) {
		g.Log().Warning(ctx, "master.cf file not found, skipping configuration update")
		return nil
	}

	// Read existing content
	masterContent := gfile.GetContents(masterCfPath)

	// Define the smtps configuration to add
	masterContent, smtpsChanged := ensureSmtpsConfigInMasterCf(masterContent)
	if smtpsChanged {
		g.Log().Info(ctx, "smtps configuration block updated in master.cf")
	}

	beginMarker := "# BEGIN BILLIONMAIL RELAY CONFIG - DO NOT EDIT THIS MARKER"
	endMarker := "# END BILLIONMAIL RELAY CONFIG - DO NOT EDIT THIS MARKER"

	var customConfigBlock strings.Builder
	customConfigBlock.WriteString(beginMarker + "\n")

	//  If there are active configurations, create service lines for each configuration
	if len(configs) > 0 {
		for _, config := range configs {
			smtpName := generateSmtpServiceName(config)

			// Configure HELO name
			heloName := config.HeloName
			serviceLine := fmt.Sprintf("%s unix - - n - - smtp", smtpName)
			var params []string
			if heloName != "" {
				params = append(params, fmt.Sprintf("-o smtp_helo_name=%s", heloName))
			}

			// Add corresponding TLS settings for different ports
			if config.RelayPort == "465" {
				params = append(params, "-o smtp_tls_wrappermode=yes")
				params = append(params, "-o smtp_tls_security_level=encrypt")
			} else if config.RelayPort == "587" {
				params = append(params, "-o smtp_tls_security_level=encrypt")
				params = append(params, "-o smtp_tls_wrappermode=no")
			} else if config.RelayPort == "25" {
				params = append(params, "-o smtp_tls_wrappermode=no")
				params = append(params, "-o smtp_tls_security_level=none")
				//params = append(params, "-o smtp_tls_security_level=may")
			}

			// Add SASL authentication settings if auth user is provided
			if config.AuthUser != "" {
				params = append(params, "-o smtp_sasl_auth_enable=yes")
				params = append(params, "-o smtp_sasl_security_options=noanonymous")
				params = append(params, "-o smtp_sasl_mechanism_filter=login,plain,cram-md5")
			}

			if config.SkipTlsVerify == 1 {
				params = append(params, "-o smtp_tls_verify_cert_match=no")
			}

			// Authentication setting
			if config.AuthMethod != "" {
				switch config.AuthMethod {
				case "LOGIN":
					params = append(params, "-o smtp_sasl_mechanism_filter=login")
				case "PLAIN":
					params = append(params, "-o smtp_sasl_mechanism_filter=plain")
				case "CRAM-MD5":
					params = append(params, "-o smtp_sasl_mechanism_filter=cram-md5")
				case "NONE":
					params = append(params, "-o smtp_sasl_auth_enable=no")
				}
			}

			if len(params) > 0 {
				serviceLine += "\n  " + strings.Join(params, "\n  ")
			}

			customConfigBlock.WriteString(serviceLine + "\n")
		}
	} else {
		customConfigBlock.WriteString("# No active relay configurations\n")
	}

	customConfigBlock.WriteString(endMarker + "\n")

	beginIndex := strings.Index(masterContent, beginMarker)
	endIndex := strings.Index(masterContent, endMarker)
	hasConfigBlock := beginIndex != -1 && endIndex != -1 && beginIndex < endIndex

	var newContent string
	if hasConfigBlock {

		blockEnd := endIndex + len(endMarker)
		if blockEnd < len(masterContent) && (masterContent[blockEnd] == '\n' || masterContent[blockEnd] == '\r') {
			blockEnd++

			if blockEnd < len(masterContent) && masterContent[blockEnd] == '\n' {
				blockEnd++
			}
		}
		newContent = masterContent[:beginIndex] + customConfigBlock.String()
		if blockEnd < len(masterContent) {
			newContent += masterContent[blockEnd:]
		}
	} else {

		if !strings.HasSuffix(masterContent, "\n") {
			masterContent += "\n"
		}
		newContent = masterContent + customConfigBlock.String()
	}

	if err := gfile.PutContents(masterCfPath, newContent); err != nil {
		return gerror.Newf("Failed to write to master.cf file: %v", err)
	}

	return nil
}

func generateRelayPwdFiles(ctx context.Context, configs []*entity.BmRelayConfig) error {
	// Ensure the configuration directory exists
	saslPasswdPath := path.Join(postfixConfigDir, saslPasswdFile)

	if !gfile.Exists(postfixConfigDir) {
		return gerror.New("Postfix configuration directory does not exist : " + postfixConfigDir)
	}

	var saslPasswdContent strings.Builder

	if len(configs) == 0 {
		saslPasswdContent.WriteString("# No active relay configurations\n")
	} else {
		hasAuthConfigs := false
		for _, config := range configs {
			// Only generate SASL entries for configurations that require authentication
			if config.AuthUser == "" {
				g.Log().Debugf(ctx, "Skipping SASL entry for relay ID %d (no authentication required): [%s]:%s",
					config.Id, config.RelayHost, config.RelayPort)
				continue
			}

			// Decrypt password
			decryptedPass, err := DecryptPassword(ctx, config.AuthPassword)
			if err != nil {
				g.Log().Warningf(ctx, "Decryption of password for configuration ID %d failed, skipping this configuration: %v", config.Id, err)
				continue
			}

			saslPasswdContent.WriteString(fmt.Sprintf("[%s]:%s %s:%s\n",
				config.RelayHost, config.RelayPort, config.AuthUser, decryptedPass))
			hasAuthConfigs = true
		}

		if !hasAuthConfigs {
			saslPasswdContent.WriteString("# No relay configurations require authentication\n")
		}
	}

	if err := gfile.PutContents(saslPasswdPath, saslPasswdContent.String()); err != nil {
		return gerror.Newf("Failed to write sasl_passwd file: %v", err)
	}

	return nil
}

func reloadPostfixConfigs(ctx context.Context) error {
	// Connect to the Docker API
	dk, err := docker.NewDockerAPI()
	if err != nil {
		g.Log().Error(ctx, "Failed to connect to Docker API:", err)
		return gerror.Wrap(err, "Failed to connect to Docker service")
	}
	defer dk.Close()

	senderRelayMapsPrimaryPath := path.Join(postfixConfigDir, senderRelayMapsPrimaryFile)
	saslPasswdPrimaryPath := path.Join(postfixConfigDir, saslPasswdPrimaryFile)

	if !gfile.Exists(senderRelayMapsPrimaryPath) {
		if err := gfile.PutContents(senderRelayMapsPrimaryPath, ""); err != nil {
			g.Log().Errorf(ctx, "Failed to create %s: %v", senderRelayMapsPrimaryPath, err)
		} else {
			g.Log().Infof(ctx, "Created empty sender_relay_maps_primary file")
		}
	}

	if !gfile.Exists(saslPasswdPrimaryPath) {
		if err := gfile.PutContents(saslPasswdPrimaryPath, ""); err != nil {
			g.Log().Errorf(ctx, "Failed to create %s: %v", saslPasswdPrimaryPath, err)
		} else {
			g.Log().Infof(ctx, "Created empty sasl_passwd_primary file")
		}
	}

	// List of commands to run
	cmdsToRun := [][]string{
		{"postmap", "/etc/postfix/conf/sender_relay_maps_primary"},
		{"postmap", "/etc/postfix/conf/sasl_passwd_primary"},
		{"postmap", "/etc/postfix/conf/sasl_passwd"},
		{"postfix", "reload"},
	}
	// Execute commands
	for _, cmd := range cmdsToRun {
		cmdStr := strings.Join(cmd, " ")
		g.Log().Infof(ctx, "Executing command in container %s: %s", postfixContainerName, cmdStr)
		result, err := dk.ExecCommandByName(ctx, postfixContainerName, cmd, "root")

		if err != nil {
			//g.Log().Errorf(ctx, "Failed to execute command: %v, Command: %s", err, cmdStr)
			return gerror.Newf("Failed to execute command: %v, Command: %s", err, cmdStr)
		}

		if result == nil {
			//g.Log().Errorf(ctx, "Command execution result is empty: %s", cmdStr)
			return gerror.Newf("Command execution result is empty: %s", cmdStr)
		}

		if result.ExitCode != 0 {
			//g.Log().Errorf(ctx, "Command execution returned non-zero status: %d, Output: %s, Command: %s", result.ExitCode, result.Output, cmdStr)
			return gerror.Newf("Command execution failed, Exit code: %d, Output: %s, Command: %s",
				result.ExitCode, result.Output, cmdStr)
		}
		g.Log().Infof(ctx, "Command executed successfully: %s, Output: %s", cmdStr, result.Output)
	}

	return nil
}

// ensurePostfixRelayConfig
func ensurePostfixRelayConfig(mainCfPath string, enableRelay bool) error {

	sqlConfigDir := path.Join(postfixConfigDir, "sql")
	if !gfile.Exists(sqlConfigDir) {
		if err := gfile.Mkdir(sqlConfigDir); err != nil {
			return gerror.Newf("Failed to create PostgreSQL configuration directory: %v", err)
		}
		g.Log().Info(context.Background(), "Created PostgreSQL configuration directory:", sqlConfigDir)
	}

	if err := createPostfixSqlConfigFile(); err != nil {
		return gerror.Newf("Failed to create SQL configuration files: %v", err)
	}

	// 3. update main.cf
	if err := writePostfixRelayConfig(mainCfPath, enableRelay); err != nil {
		return err
	}

	return nil
}

// createPostfixSqlConfigFile
func createPostfixSqlConfigFile() error {
	dbpass, _ := public.DockerEnv("DBPASS")
	dbname, _ := public.DockerEnv("DBNAME")
	dbuser, _ := public.DockerEnv("DBUSER")

	sqlConfigFiles := map[string]string{
		"pgsql_sender_relay_maps.cf": fmt.Sprintf(`user = %s
password = %s
hosts = pgsql
dbname = %s

query = SELECT CONCAT('[', rc.relay_host, ']:', rc.relay_port) FROM bm_relay_config rc JOIN bm_relay_domain_mapping rdm ON rc.id = rdm.relay_id WHERE rdm.sender_domain = REPLACE('%%s', '@', '') AND rc.active = 1 LIMIT 1`, dbuser, dbpass, dbname),

		"pgsql_sender_transport_maps.cf": fmt.Sprintf(`user = %s
password = %s
hosts = pgsql
dbname = %s

query = SELECT CONCAT(smtp_name, ':') FROM bm_domain_smtp_transport WHERE domain = '%%s' LIMIT 1`, dbuser, dbpass, dbname),
	}

	sqlDir := path.Join(postfixConfigDir, "sql")
	if !gfile.Exists(sqlDir) {
		if err := gfile.Mkdir(sqlDir); err != nil {
			return gerror.Newf("Failed to create SQL directory: %v", err)
		}
		g.Log().Info(context.Background(), "Created SQL directory:", sqlDir)
	}

	for fileName, fileContent := range sqlConfigFiles {
		filePath := path.Join(sqlDir, fileName)
		if !gfile.Exists(filePath) {
			if err := gfile.PutContents(filePath, fileContent); err != nil {
				return gerror.Newf("Failed to create SQL configuration file %s: %v", fileName, err)
			}
			g.Log().Infof(context.Background(), "Created SQL configuration file: %s", filePath)
		} else {
			g.Log().Debugf(context.Background(), "SQL configuration file already exists: %s", filePath)
		}
	}

	return nil
}

func writePostfixRelayConfig(cfPath string, enableRelay bool) error {

	if !gfile.Exists(cfPath) {
		if err := gfile.PutContents(cfPath, ""); err != nil {
			return gerror.Newf("Failed to create configuration file %s: %v", cfPath, err)
		}
	}

	content := gfile.GetContents(cfPath)

	beginMarker := "# BEGIN RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"
	endMarker := "# END RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"

	var configBlock string

	configBlock = fmt.Sprintf(`%s
sender_dependent_relayhost_maps = hash:/etc/postfix/conf/sender_relay_maps_primary, pgsql:/etc/postfix/sql/pgsql_sender_relay_maps.cf
smtp_sasl_password_maps = hash:/etc/postfix/conf/sasl_passwd_primary, hash:/etc/postfix/conf/sasl_passwd
sender_dependent_default_transport_maps = pgsql:/etc/postfix/sql/pgsql_sender_transport_maps.cf
%s`, beginMarker, endMarker)

	// Find the existing configuration block
	beginIndex := strings.Index(content, beginMarker)
	endIndex := strings.Index(content, endMarker)

	// Check if the configuration block exists
	hasConfigBlock := beginIndex != -1 && endIndex != -1 && beginIndex < endIndex

	// Decide the operation based on whether relay is enabled and if the configuration block exists
	modified := false

	if enableRelay {
		if hasConfigBlock {
			g.Log().Debugf(nil, "Relay configuration block already exists in %s, no modification needed", cfPath)
		} else {
			// Need to add the configuration block
			if strings.HasSuffix(content, "\n") {
				content = content + configBlock + "\n"
			} else {
				content = content + "\n" + configBlock + "\n"
			}
			modified = true
		}
    
	} else if hasConfigBlock {
		// Remove the relay configuration block
		content = content[:beginIndex] + content[endIndex+len(endMarker):]
		// Clean up extra blank lines
		content = strings.TrimRight(content, "\n") + "\n"
		modified = true
	}
	// If there are modifications, write to the file
	if modified {
		if err := gfile.PutContents(cfPath, content); err != nil {
			return gerror.Newf("Failed to write to file %s: %v", cfPath, err)
		}
	}

	return nil
}

// updateSmtpServiceMappings updates the SMTP service name mapping table
func updateSmtpServiceMappings(ctx context.Context, configs []*entity.BmRelayConfig) error {
	// Query active configuration and domain mappings
	type RelayDomainMapping struct {
		RelayId      int64  `json:"relay_id"`
		SenderDomain string `json:"sender_domain"`
	}

	var mappings []RelayDomainMapping
	err := g.DB().Model("bm_relay_domain_mapping rdm").
		LeftJoin("bm_relay_config rc", "rdm.relay_id = rc.id").
		Where("rc.active", 1).
		Fields("rdm.relay_id, rdm.sender_domain").
		Scan(&mappings)

	if err != nil {
		return gerror.Wrap(err, "failed to query relay domain mappings")
	}

	if len(mappings) == 0 {
		g.Log().Info(ctx, "No domain mappings, skip updating SMTP service name mappings")
		return nil
	}

	// Build map from relay_id to config for lookup
	configMap := make(map[int64]*entity.BmRelayConfig)
	for _, config := range configs {
		configMap[int64(config.Id)] = config
	}

	// Clear existing mapping table (avoid stale data)
	_, err = g.DB().Model("bm_domain_smtp_transport").Where("atype", "relay").Delete()
	if err != nil {
		return gerror.Wrap(err, "failed to clear SMTP service name mapping table")
	}

	// Batch insert new mappings
	var transportMappings []g.Map

	for _, mapping := range mappings {
		config, exists := configMap[mapping.RelayId]
		if !exists || config.Active != 1 {
			continue // Skip inactive configs
		}

		// Generate SMTP service name
		smtpName := generateSmtpServiceName(config)

		// Ensure domain starts with @
		senderDomain := mapping.SenderDomain
		if !strings.HasPrefix(senderDomain, "@") {
			senderDomain = "@" + senderDomain
		}

		// Add to batch insert list
		transportMappings = append(transportMappings, g.Map{
			"domain":    senderDomain,
			"smtp_name": smtpName,
			"atype":     "relay",
		})
	}

	if len(transportMappings) == 0 {
		g.Log().Info(ctx, "No valid SMTP service name mappings to update")
		return nil
	}

	// Batch upsert new mappings, using ON CONFLICT to handle cases where
	// another atype (e.g. 'local') already owns the domain.
	// Log but don't fail — blocking this would prevent main.cf relay config
	// block from being written, breaking all relay functionality.
	_, err = g.DB().Model("bm_domain_smtp_transport").
		Data(transportMappings).
		OnConflict("domain").
		OnDuplicate("smtp_name", "atype").
		Batch(100).
		Insert()

	if err != nil {
		g.Log().Warningf(ctx, "Failed to batch insert SMTP service name mappings (non-fatal): %v", err)
		// Don't return — let the rest of the sync continue (main.cf update, postfix reload)
	} else {
		g.Log().Infof(ctx, "Updated %d SMTP service name mappings", len(transportMappings))
	}
	return nil
}

// maxSmtpServiceNameLength limits Postfix unix-domain socket path length.
const maxSmtpServiceNameLength = 64

// generateSmtpServiceName generates SMTP service name based on relay configuration.
// Names are truncated to avoid exceeding Postfix unix-domain socket path limits.
func generateSmtpServiceName(config *entity.BmRelayConfig) string {
	if config.SmtpName != "Custom SMTP Relay" {
		// Prefer user-defined SMTP name
		cleanedSmtpName := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(config.SmtpName, "_")
		name := fmt.Sprintf("smtp_Custom_%s", cleanedSmtpName)
		return truncateSmtpName(name)
	}

	// Build base from relay_host + auth_user
	relayDomain := config.RelayHost
	cleanedRelayDomain := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(relayDomain, "_")

	userName := ""
	if config.AuthUser != "" {
		if strings.Contains(config.AuthUser, "@") {
			parts := strings.Split(config.AuthUser, "@")
			userName = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(parts[0], "_")
		} else {
			authUserLen := len(config.AuthUser)
			if authUserLen > 5 {
				userName = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(config.AuthUser[:5], "_")
			} else {
				userName = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(config.AuthUser, "_")
			}
		}
	}

	var name string
	if userName != "" {
		name = fmt.Sprintf("smtp_Relay_%s_%s_%d", cleanedRelayDomain, userName, config.Id)
	} else {
		name = fmt.Sprintf("smtp_Relay_%s_%d", cleanedRelayDomain, config.Id)
	}

	return truncateSmtpName(name)
}

// truncateSmtpName ensures the SMTP service name fits within Postfix limits.
// If the name is too long, it replaces the middle portion with a short SHA-256
// hash while preserving the suffix (config ID) for uniqueness.
func truncateSmtpName(name string) string {
	if len(name) <= maxSmtpServiceNameLength {
		return name
	}

	// Extract the trailing _{id} part to preserve uniqueness
	idIdx := strings.LastIndex(name, "_")
	idSuffix := ""
	if idIdx > 0 {
		idSuffix = name[idIdx:]
	}

	// Use a hash of the full name to guarantee uniqueness
	hash := sha256.Sum256([]byte(name))
	hashStr := hex.EncodeToString(hash[:])[:12]

	prefix := "smtp"
	if strings.HasPrefix(name, "smtp_Custom_") {
		prefix = "smtp_Custom"
	} else if strings.HasPrefix(name, "smtp_Relay_") {
		prefix = "smtp_Relay"
	}

	truncated := fmt.Sprintf("%s_%s%s", prefix, hashStr, idSuffix)
	if len(truncated) > maxSmtpServiceNameLength {
		truncated = truncated[:maxSmtpServiceNameLength]
	}

	g.Log().Warningf(nil, "SMTP service name truncated from %d to %d chars: %s",
		len(name), len(truncated), truncated)
	return truncated
}

func ensureSmtpsConfigInMasterCf(content string) (string, bool) {
	const (
		markerBegin = "# BEGIN BILLIONMAIL SMTPS CONFIG - DO NOT EDIT THIS MARKER"
		markerEnd   = "# END BILLIONMAIL SMTPS CONFIG - DO NOT EDIT THIS MARKER"
	)
	smtpsService := `smtps     unix  -       -       n       -       -       smtp
    -o smtp_tls_wrappermode=yes
    -o smtp_tls_security_level=encrypt`

	var block strings.Builder
	block.WriteString(markerBegin + "\n")
	block.WriteString(smtpsService + "\n")
	block.WriteString(markerEnd + "\n")
	smtpsBlock := block.String()

	beginIndex := strings.Index(content, markerBegin)
	endIndex := strings.Index(content, markerEnd)
	hasBlock := beginIndex != -1 && endIndex != -1 && beginIndex < endIndex

	if hasBlock {
		blockEnd := endIndex + len(markerEnd)

		for blockEnd < len(content) && (content[blockEnd] == '\n' || content[blockEnd] == '\r') {
			blockEnd++
		}

		existingBlock := content[beginIndex:blockEnd]
		if strings.Contains(existingBlock, smtpsService) {

			return content, false
		}

		newContent := content[:beginIndex] + smtpsBlock
		if blockEnd < len(content) {
			newContent += content[blockEnd:]
		}
		return newContent, true
	}

	var updated string
	if !strings.HasSuffix(content, "\n") {
		updated = content + "\n"
	} else {
		updated = content
	}
	updated += "\n" + smtpsBlock
	return updated, true
}

// commentOutOldTransportMaps Comment out the previously manually added "sender_dependent_default_transport_maps" configuration
func commentOutOldTransportMaps(content string) string {

	beginMarker := "# BEGIN RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"
	endMarker := "# END RELAY SERVICE CONFIGURATION - DO NOT EDIT THIS MARKER"

	beginIndex := strings.Index(content, beginMarker)
	endIndex := strings.Index(content, endMarker)
	inAutomationBlock := func(pos int) bool {
		return beginIndex != -1 && endIndex != -1 && pos >= beginIndex && pos <= endIndex+len(endMarker)
	}

	transportRegex := regexp.MustCompile(`(?m)^\s*(sender_dependent_default_transport_maps\s*=\s*(hash|pgsql|tcp):[^\r\n]*)`)

	return transportRegex.ReplaceAllStringFunc(content, func(match string) string {
		matchIndex := strings.Index(content, match)
		if matchIndex == -1 {
			return match
		}

		// Within the automation block, no processing is performed
		if inAutomationBlock(matchIndex) {
			return match
		}

		// Commented, Not processed.
		if strings.HasPrefix(strings.TrimSpace(match), "#") {
			return match
		}

		return "# " + match + " # Migrated: replaced by automated config block"
	})
}
