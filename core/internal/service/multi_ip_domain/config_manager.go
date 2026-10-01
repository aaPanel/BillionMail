package multi_ip_domain

import (
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/util/gconv"
	"gopkg.in/yaml.v3"
)

var (
	BasePostfixDir      = public.AbsPath("../conf/postfix")
	PostfixMasterCfPath = BasePostfixDir + "/master.cf"
	PostfixMainCfPath   = BasePostfixDir + "/main.cf"
)

const (
	masterCfBlockMarkerStart = "# BEGIN BILLIONMAIL multi-ip services"
	masterCfBlockMarkerEnd   = "# END BILLIONMAIL multi-ip services"
)

// ConfigManager
type ConfigManager struct {
	ctx               context.Context
	dockerComposeData map[string]interface{}
	masterCfEntries   []string
	fileLocker        *gmutex.RWMutex
}

// NewConfigManager
func NewConfigManager(ctx context.Context) (*ConfigManager, error) {
	manager := &ConfigManager{
		ctx:             ctx,
		masterCfEntries: []string{},
		fileLocker:      gmutex.New(),
	}

	return manager, nil
}

// ApplyConfigsWithRollback Apply configurations and rollback on failure
func (m *ConfigManager) ApplyConfigsWithRollback(ctx context.Context, configs []map[string]interface{}) error {

	m.fileLocker.Lock()
	defer m.fileLocker.Unlock()

	backups, err := m.createBackups(ctx)
	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Apply configurations
	if err := m.applyConfigs(ctx, configs); err != nil {

		g.Log().Debugf(ctx, "Failed to apply configurations, rolling back... Error: %v", err)
		if rollbackErr := m.rollback(ctx, backups); rollbackErr != nil {
			return fmt.Errorf("failed to apply configurations: %v, rollback also failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("failed to apply configurations, but rollback succeeded: %v", err)
	}

	m.cleanupBackups(ctx, backups)
	return nil
}

// createBackups Create backups for Postfix's main.cf and master.cf
func (m *ConfigManager) createBackups(ctx context.Context) (map[string]string, error) {
	backups := make(map[string]string)

	files := []string{PostfixMainCfPath, PostfixMasterCfPath}

	for _, file := range files {
		if gfile.Exists(file) {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %v", file, err)
			}

			backupPath := fmt.Sprintf("%s.backup.%d", file, time.Now().UnixNano())
			if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
				return nil, fmt.Errorf("failed to create backup for %s: %v", file, err)
			}
			g.Log().Debugf(ctx, "Created backup for %s at %s", file, backupPath)
			backups[file] = backupPath
		}
	}

	return backups, nil
}

// rollback Roll back configurations
func (m *ConfigManager) rollback(ctx context.Context, backups map[string]string) error {
	var allErrors []string
	for file, backup := range backups {
		if gfile.Exists(backup) {
			g.Log().Infof(ctx, "Rolling back %s from %s", file, backup)
			content, err := ioutil.ReadFile(backup)
			if err != nil {
				err = fmt.Errorf("failed to read backup file %s: %v", backup, err)
				allErrors = append(allErrors, err.Error())
				continue
			}

			if err := ioutil.WriteFile(file, content, 0644); err != nil {
				err = fmt.Errorf("failed to rollback file %s: %v", file, err)
				allErrors = append(allErrors, err.Error())
			}
		}
	}
	if len(allErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(allErrors, "; "))
	}
	return nil
}

// cleanupBackups Clean up backup files
func (m *ConfigManager) cleanupBackups(ctx context.Context, backups map[string]string) {
	for _, backup := range backups {
		if gfile.Exists(backup) {
			g.Log().Debugf(ctx, "Cleaning up backup file: %s", backup)
			_ = os.Remove(backup)
		}
	}
}

// applyConfigs Apply configurations
func (m *ConfigManager) applyConfigs(ctx context.Context, configs []map[string]interface{}) error {
	// 1. Update docker-compose.yml
	if err := m.updateDockerCompose(ctx, configs); err != nil {

		return fmt.Errorf("failed to update docker-compose.yml: %v", err)
	}

	// 2. Update Postfix configurations
	if err := m.updatePostfixConfigs(ctx, configs); err != nil {
		return fmt.Errorf("failed to update Postfix configurations: %v", err)
	}

	return nil
}

// updateDockerCompose Generate docker-compose_addnetwork.yml
func (m *ConfigManager) updateDockerCompose(ctx context.Context, configs []map[string]interface{}) error {
	originalPath := filepath.Join(public.HostWorkDir, "docker-compose.yml")
	outputPath := filepath.Join(public.HostWorkDir, "docker-compose_addnetwork.yml")

	// 设置临时文件路径
	// 容器内路径：/opt/billionmail/core/data (通过 public.AbsPath 获取)
	// 宿主机路径：./core-data (相对于 docker-compose.yml 所在目录)
	containerDataPath := public.AbsPath("../core/data/")
	hostDataPath := filepath.Join(public.HostWorkDir, "core-data") // 宿主机的实际映射路径
	tempDockerComposePath := filepath.Join(containerDataPath, "temp_docker-compose.yml")
	hostTempDockerComposePath := filepath.Join(hostDataPath, "temp_docker-compose.yml")

	// Ensure container directory exists
	if !gfile.Exists(containerDataPath) {
		if err := gfile.Mkdir(containerDataPath); err != nil {
			return fmt.Errorf("failed to create temporary directory %s: %v", containerDataPath, err)
		}
	}

	dk, err := docker.NewDockerAPI()
	if err != nil {
		return gerror.New(public.LangCtx(ctx, "failed to create Docker API instance: %v", err.Error()))
	}
	defer dk.Close()

	// Step 1: 复制宿主机的 docker-compose.yml 到宿主机的 core-data 目录
	// 注意：这里目标路径使用宿主机的路径，因为 chroot /host_root 看到的是宿主机文件系
	copyCmd := []string{
		"/bin/sh", "-c",
		fmt.Sprintf(`chroot /host_root cp "%s" "%s"`, originalPath, hostTempDockerComposePath),
	}
	result, err := dk.ExecHostCommand(ctx, copyCmd)
	if err != nil || result.ExitCode != 0 {
		return gerror.New("failed to copy original configuration file, please check path and permissions")
	}

	// Step 2: 从容器内路径读取复制的文件
	// 由于文件映射，容器内可以通过映射路径读到刚才复制的文件
	content := gfile.GetContents(tempDockerComposePath)
	if content == "" {
		return fmt.Errorf("failed to read temporary docker-compose.yml file: file is empty or does not exist")
	}

	// Cleanup temporary file
	defer func() {
		if gfile.Exists(tempDockerComposePath) {
			os.Remove(tempDockerComposePath)
			g.Log().Debugf(ctx, "Cleaned up temporary file: %s", tempDockerComposePath)
		}
	}()

	g.Log().Debugf(ctx, "Successfully read docker-compose.yml file, size: %d bytes", len(content))
	newContent, err := m.modifyDockerComposeText(ctx, content, configs)
	if err != nil {
		return fmt.Errorf("failed to modify docker-compose content: %v", err)
	}

	// Step 7: Write new file docker-compose_addnetwork.yml
	tempOutputPath := filepath.Join(containerDataPath, "temp_docker-compose_addnetwork.yml")
	hostTempOutputPath := filepath.Join(hostDataPath, "temp_docker-compose_addnetwork.yml")

	// First write to temporary container file
	if err := gfile.PutContents(tempOutputPath, newContent); err != nil {
		return fmt.Errorf("failed to write temporary output file: %v", err)
	}

	// 然后从宿主机临时位置复制到最终目标位置
	copyOutputCmd := []string{
		"/bin/sh", "-c",
		fmt.Sprintf(`chroot /host_root cp "%s" "%s"`, hostTempOutputPath, outputPath),
	}

	result, err = dk.ExecHostCommand(ctx, copyOutputCmd)
	if err != nil || result.ExitCode != 0 {
		if gfile.Exists(tempOutputPath) {
			os.Remove(tempOutputPath)
		}
		return gerror.New("failed to copy output file to destination")
	}

	// Cleanup temporary output file
	if gfile.Exists(tempOutputPath) {
		os.Remove(tempOutputPath)
		g.Log().Debugf(ctx, "Cleaned up temporary output file: %s", tempOutputPath)
	}

	return nil
}

// modifyDockerComposeText
func (m *ConfigManager) modifyDockerComposeText(ctx context.Context, originalContent string, configs []map[string]interface{}) (string, error) {
	lines := strings.Split(originalContent, "\n")

	// === 1. Modify networks section ===
	modifiedContent, err := m.addCustomNetworksText(lines, configs)
	if err != nil {
		return "", err
	}

	// === 2. Modify networks configuration of the postfix-billionmail service ===
	finalContent, err := m.modifyPostfixNetworksText(modifiedContent, configs)
	if err != nil {
		return "", err
	}

	return strings.Join(finalContent, "\n"), nil
}

// addCustomNetworksText
func (m *ConfigManager) addCustomNetworksText(lines []string, configs []map[string]interface{}) ([]string, error) {
	result := make([]string, 0, len(lines)+50) // Reserve space

	// Find the position of the networks section
	networksStartIdx := -1
	networksEndIdx := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "networks:") {
			networksStartIdx = i
		} else if networksStartIdx != -1 && strings.HasPrefix(line, "services:") {
			networksEndIdx = i
			break
		}
	}

	// Initialize indentation variables (within appropriate scope)
	baseIndent := "  "     // Default 2 spaces (network name indentation)
	configIndent := "    " // Default 4 spaces (configuration item indentation)

	if networksStartIdx == -1 {
		// If there is no networks section, add it at the end of the file
		result = append(result, lines...)
		result = append(result, "", "networks:")
		networksStartIdx = len(result) - 1
		networksEndIdx = len(result)
	} else {
		// Copy content before the networks section
		result = append(result, lines[:networksStartIdx+1]...)

		// Copy existing billionmail-network configuration and detect indentation style
		billionmailNetworkLines := []string{}
		inBillionmailNetwork := false

		for i := networksStartIdx + 1; i < len(lines) && (networksEndIdx == -1 || i < networksEndIdx); i++ {
			line := lines[i]
			if strings.Contains(line, "billionmail-network:") {
				inBillionmailNetwork = true
				// Detect indentation level of network name
				baseIndent = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				billionmailNetworkLines = append(billionmailNetworkLines, line)
			} else if inBillionmailNetwork {
				if strings.TrimSpace(line) == "" {
					billionmailNetworkLines = append(billionmailNetworkLines, line)
				} else {
					// Detect indentation of configuration items (e.g., driver: bridge)
					lineIndent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
					if len(lineIndent) > len(baseIndent) {
						if configIndent == "    " { // Only update if using default value
							configIndent = lineIndent
						}
						billionmailNetworkLines = append(billionmailNetworkLines, line)
					} else if len(lineIndent) == len(baseIndent) && strings.TrimSpace(line) != "" {
						// Encountered another network definition at the same level, stop
						break
					} else if len(lineIndent) > len(baseIndent) {
						// Continue collecting billionmail-network configurations
						billionmailNetworkLines = append(billionmailNetworkLines, line)
					} else {
						// Encountered lower-level configuration, stop
						break
					}
				}
			}
		}

		result = append(result, billionmailNetworkLines...)

		// Ensure correct indentation style (use default if not detected)
		if baseIndent == "" {
			baseIndent = "  " // Default 2 spaces
		}
		if configIndent == "" || configIndent == "    " {
			configIndent = baseIndent + "  " // Add 2 more spaces to base indentation
		}
	}

	// Add custom networks using detected indentation style
	for _, cfg := range configs {
		if !gconv.Bool(cfg["active"]) {
			continue
		}

		networkName := gconv.String(cfg["network_name"])
		subnet := gconv.String(cfg["subnet"])

		if networkName == "" || subnet == "" {
			continue
		}

		// Calculate gateway
		gateway := ""
		if subnet != "" {
			if _, ipNet, err := net.ParseCIDR(subnet); err == nil {
				gatewayIP := make(net.IP, len(ipNet.IP))
				copy(gatewayIP, ipNet.IP)
				gatewayIP[len(gatewayIP)-1] += 1
				gateway = gatewayIP.String()
			}
		}

		// Use detected indentation style to add custom network configuration
		result = append(result, fmt.Sprintf("%s%s:", baseIndent, networkName))
		result = append(result, fmt.Sprintf("%sdriver: bridge", configIndent))
		result = append(result, fmt.Sprintf("%sipam:", configIndent))
		result = append(result, fmt.Sprintf("%s%sconfig:", configIndent, "  "))
		if gateway != "" {
			result = append(result, fmt.Sprintf("%s%s- gateway: %s", configIndent, "    ", gateway))
			result = append(result, fmt.Sprintf("%s%s  subnet: %s", configIndent, "    ", subnet))
		} else {
			result = append(result, fmt.Sprintf("%s%s- subnet: %s", configIndent, "    ", subnet))
		}
		result = append(result, fmt.Sprintf("%s%sdriver: default", configIndent, "  "))

		g.Log().Debugf(context.Background(), "Added custom network: %s with subnet: %s", networkName, subnet)
	}

	// Add content after the networks section
	if networksEndIdx != -1 {
		result = append(result, lines[networksEndIdx:]...)
	}

	return result, nil
}

// modifyPostfixNetworksText 修改postfix-billionmail服务的networks配置
func (m *ConfigManager) modifyPostfixNetworksText(lines []string, configs []map[string]interface{}) ([]string, error) {
	result := make([]string, 0, len(lines)+20)

	// Find the postfix-billionmail service section
	postfixStartIdx := -1
	postfixNetworksStartIdx := -1
	postfixNetworksEndIdx := -1

	for i, line := range lines {
		if strings.Contains(line, "postfix-billionmail:") {
			postfixStartIdx = i
		} else if postfixStartIdx != -1 && strings.Contains(line, "networks:") {
			postfixNetworksStartIdx = i
		} else if postfixNetworksStartIdx != -1 && strings.TrimSpace(line) != "" &&
			!strings.HasPrefix(line, "        ") && !strings.HasPrefix(line, "\t\t") {
			postfixNetworksEndIdx = i
			break
		}
	}

	if postfixStartIdx == -1 {
		return lines, fmt.Errorf("postfix-billionmail service not found")
	}

	// Copy content up to the postfix networks section
	if postfixNetworksStartIdx != -1 {
		result = append(result, lines[:postfixNetworksStartIdx+1]...)
	} else {
		// If there is no networks configuration, it needs to be added
		return lines, fmt.Errorf("postfix-billionmail networks section not found")
	}

	// Add default billionmail-network configuration with fixed IP
	result = append(result, "        billionmail-network:")
	result = append(result, "          aliases:")
	result = append(result, "            - postfix")
	result = append(result, "          ipv4_address: 172.66.1.100")

	// Add custom network configurations
	for _, cfg := range configs {
		if !gconv.Bool(cfg["active"]) {
			continue
		}

		networkName := gconv.String(cfg["network_name"])
		postfixIP := gconv.String(cfg["postfix_ip"])
		aliases := gconv.String(cfg["aliases"])

		if networkName == "" || postfixIP == "" {
			continue
		}

		result = append(result, fmt.Sprintf("        %s:", networkName))
		result = append(result, "          aliases:")
		result = append(result, fmt.Sprintf("            - %s", aliases))
		result = append(result, fmt.Sprintf("          ipv4_address: %s", postfixIP))

		g.Log().Debugf(context.Background(), "Added postfix network: %s -> %s (%s)", networkName, postfixIP, aliases)
	}

	// Add content after the networks section
	if postfixNetworksEndIdx != -1 {
		result = append(result, lines[postfixNetworksEndIdx:]...)
	}

	return result, nil
}

// deepCopy performs a deep copy of any value using YAML marshal/unmarshal
func deepCopy(src interface{}) interface{} {
	data, _ := yaml.Marshal(src)
	var copy interface{}
	yaml.Unmarshal(data, &copy)
	return copy
}

// rebuildPostfixServiceNetworks Rebuild Postfix service network configuration
func (m *ConfigManager) rebuildPostfixServiceNetworks(config map[string]interface{}, configs []map[string]interface{}) error {
	services, ok := config["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'services' in config")
	}

	postfix, ok := services["postfix-billionmail"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("postfix-billionmail service not found")
	}

	// Get current networks (preserve original)
	networks, ok := postfix["networks"].(map[string]interface{})
	if !ok {
		networks = make(map[string]interface{})
	}

	// === 1. Remove all custom networks (except default billionmail-network) ===
	for name := range networks {
		if name != "billionmail-network" {
			delete(networks, name)
			g.Log().Debugf(context.Background(), "Removed old custom network: %s", name)
		}
	}

	// === 2. Add networks enabled in configs ===
	for _, cfg := range configs {
		if !gconv.Bool(cfg["active"]) {
			g.Log().Debugf(context.Background(), "Skipping inactive configuration: %+v", cfg)
			continue
		}

		networkName := gconv.String(cfg["network_name"])
		postfixIP := gconv.String(cfg["postfix_ip"])
		aliases := gconv.String(cfg["aliases"])

		if networkName == "" || postfixIP == "" {
			g.Log().Debugf(context.Background(), "Skipping invalid configuration: network_name=%s, postfix_ip=%s", networkName, postfixIP)
			continue
		}

		// Build network configuration, including ipv4_address and aliases
		netConfig := map[string]interface{}{
			"ipv4_address": postfixIP,
		}

		if aliases != "" {
			netConfig["aliases"] = []string{aliases}
		}

		networks[networkName] = netConfig
		g.Log().Infof(context.Background(), "Added Postfix network configuration: %s -> %+v", networkName, netConfig)
	}

	postfix["networks"] = networks

	return nil
}

// rebuildTopLevelNetworks Rebuild top-level networks configuration
func (m *ConfigManager) rebuildTopLevelNetworks(config map[string]interface{}, configs []map[string]interface{}) error {
	networks, ok := config["networks"].(map[string]interface{})
	if !ok {
		networks = make(map[string]interface{})
		config["networks"] = networks
	}

	// Keep default network
	defaultNet, hasDefault := networks["billionmail-network"]

	// Clear custom networks
	for name := range networks {
		if name != "billionmail-network" {
			delete(networks, name)
		}
	}

	// Re-add default network
	if hasDefault {
		networks["billionmail-network"] = defaultNet
	}

	// Add networks defined in configs
	for _, cfg := range configs {
		if !gconv.Bool(cfg["active"]) {
			continue
		}

		networkName := gconv.String(cfg["network_name"])
		subnet := gconv.String(cfg["subnet"])

		if networkName == "" || subnet == "" {
			continue
		}

		// Calculate gateway from subnet
		gateway := ""
		if subnet != "" {

			if _, ipNet, err := net.ParseCIDR(subnet); err == nil {
				gatewayIP := make(net.IP, len(ipNet.IP))
				copy(gatewayIP, ipNet.IP)
				gatewayIP[len(gatewayIP)-1] += 1
				gateway = gatewayIP.String()
			}
		}

		// Build network definition
		customNet := map[string]interface{}{
			"driver": "bridge",
		}

		// Add ipam configuration
		ipamConfig := map[string]interface{}{
			"driver": "default",
			"config": []interface{}{
				map[string]interface{}{
					"subnet": subnet,
				},
			},
		}

		// If gateway exists, add to configuration
		if gateway != "" {
			ipamConfig["config"] = []interface{}{
				map[string]interface{}{
					"subnet":  subnet,
					"gateway": gateway,
				},
			}
		}

		customNet["ipam"] = ipamConfig

		networks[networkName] = customNet
	}

	return nil
}

// rebuildPostfixServiceNetworks Rebuild network configuration for Postfix service
func (m *ConfigManager) rebuildPostfixServiceNetworks1(dockerConfig map[string]interface{}, configs []map[string]interface{}) error {
	services, ok := dockerConfig["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'services' section in docker-compose.yml")
	}
	postfixService, ok := services["postfix-billionmail"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Postfix service 'postfix-billionmail' configuration does not exist")
	}

	// Create a new network mapping, starting with the default network
	newNetworks := map[string]interface{}{
		"billionmail-network": map[string]interface{}{
			"aliases": []string{"postfix"},
		},
	}

	// Add all active custom networks
	for _, config := range configs {
		networkName := gconv.String(config["network_name"])
		postfixIP := gconv.String(config["postfix_ip"])
		alias := gconv.String(config["aliases"])

		if networkName == "" || postfixIP == "" || alias == "" {
			continue
		}

		if net.ParseIP(postfixIP) == nil {
			continue
		}

		newNetworks[networkName] = map[string]interface{}{
			"ipv4_address": postfixIP,
			"aliases":      []string{alias},
		}
	}

	postfixService["networks"] = newNetworks
	return nil
}

// updatePostfixConfigs Update all Postfix-related configuration files
func (m *ConfigManager) updatePostfixConfigs(ctx context.Context, configs []map[string]interface{}) error {

	// Check if main.cf has sender_dependent_relayhost_maps = pgsql:/etc/postfix/sql/pgsql_sender_relay_maps.cf
	if err := m.updateMainCf(ctx, configs); err != nil {
		return err
	}
	// Update corresponding master.cf configuration
	if err := m.updateMasterCf(ctx, configs); err != nil {
		return err
	}

	return nil
}

// updateMainCf
func (m *ConfigManager) updateMainCf(ctx context.Context, configs []map[string]interface{}) error {
	content, err := ioutil.ReadFile(PostfixMainCfPath)
	if err != nil {
		return gerror.Wrapf(err, "failed to read main.cf file: %s", PostfixMainCfPath)
	}
	contentStr := string(content)

	// check smtp_bind_address=172.66.1.100
	defaultBindIP := "172.66.1.100"
	smtpBindRegex := regexp.MustCompile(`(?m)^[\s]*smtp_bind_address[\s]*=[\s]*(.*)$`)
	smtpBindMatches := smtpBindRegex.FindStringSubmatch(contentStr)

	if smtpBindMatches == nil {
		// smtp_bind_address  configuration does not exist, add it
		smtpBindLine := "\nsmtp_bind_address = " + defaultBindIP + "\n"
		contentStr += smtpBindLine

		if err := ioutil.WriteFile(PostfixMainCfPath, []byte(contentStr), 0644); err != nil {
			return gerror.Wrapf(err, "failed to write main.cf file: %s", PostfixMainCfPath)
		}

	} else {
		// exists, check if it's commented out or empty
		currentBindIP := strings.TrimSpace(smtpBindMatches[1])
		if currentBindIP == "" || strings.HasPrefix(currentBindIP, "#") {
			// commented out or empty, update to default
			newContentStr := smtpBindRegex.ReplaceAllString(contentStr, "smtp_bind_address = "+defaultBindIP)

			if err := ioutil.WriteFile(PostfixMainCfPath, []byte(newContentStr), 0644); err != nil {
				return gerror.Wrapf(err, "failed to write main.cf file: %s", PostfixMainCfPath)
			}
			g.Log().Info(ctx, "update: smtp_bind_address =", defaultBindIP)
		}
	}

	// Only need to ensure sender_dependent_relayhost_maps exists and contains pgsql_sender_relay_maps.cf
	requiredPgsqlMapEntry := "pgsql:/etc/postfix/sql/pgsql_sender_relay_maps.cf"

	// Find sender_dependent_relayhost_maps configuration line
	senderRelayRegex := regexp.MustCompile(`(?m)^[\s]*sender_dependent_relayhost_maps[\s]*=[\s]*(.*)$`)
	matches := senderRelayRegex.FindStringSubmatch(contentStr)

	// Need to enable multi-IP domain feature, ensure the configuration line exists and contains the required mapping file
	if matches == nil {
		// Configuration line does not exist, add entire line
		newLine := "\nsender_dependent_relayhost_maps = " + requiredPgsqlMapEntry + "\n"
		contentStr += newLine

		// Write back to file
		if err := ioutil.WriteFile(PostfixMainCfPath, []byte(contentStr), 0644); err != nil {
			return gerror.Wrapf(err, "failed to write main.cf file: %s", PostfixMainCfPath)
		}

	} else {
		// Configuration line exists, check if it contains the required mapping file
		currentValue := strings.TrimSpace(matches[1])
		if !strings.Contains(currentValue, requiredPgsqlMapEntry) {
			// Need to add the required mapping file
			var newValue string
			if strings.TrimSpace(currentValue) == "" || strings.HasPrefix(strings.TrimSpace(currentValue), "#") {
				newValue = requiredPgsqlMapEntry
			} else {
				newValue = requiredPgsqlMapEntry + ", " + currentValue
			}

			// Replace configuration line
			newContentStr := senderRelayRegex.ReplaceAllString(contentStr, "sender_dependent_relayhost_maps = "+newValue)

			// Write back to file
			if err := ioutil.WriteFile(PostfixMainCfPath, []byte(newContentStr), 0644); err != nil {
				return gerror.Wrapf(err, "failed to write main.cf file: %s", PostfixMainCfPath)
			}
		}
	}

	return nil
}

// updateMasterCf Use marker blocks to update master.cf
func (m *ConfigManager) updateMasterCf(ctx context.Context, configs []map[string]interface{}) error {

	content, err := ioutil.ReadFile(PostfixMasterCfPath)
	if err != nil {
		return err
	}
	contentStr := string(content)

	var blockBuilder strings.Builder
	blockBuilder.WriteString(masterCfBlockMarkerStart + "\n")

	for _, config := range configs {
		serviceName := gconv.String(config["smtp_server_name"])
		if serviceName == "" {
			continue
		}
		postfixIP := gconv.String(config["postfix_ip"])
		//domain := gconv.String(config["domain"])

		blockBuilder.WriteString(fmt.Sprintf("%-10s unix  -       -       n       -       -       smtp\n", serviceName))
		blockBuilder.WriteString(fmt.Sprintf("    -o smtp_bind_address=%s\n", postfixIP))

		//g.Log().Debugf(ctx, "Generating Postfix service %s for domain %s (ID=%d), binding IP: %s",
		//	domain, config["id"], serviceName, postfixIP)

	}
	blockBuilder.WriteString(masterCfBlockMarkerEnd)

	re := regexp.MustCompile(`(?s)` + masterCfBlockMarkerStart + `.*?` + masterCfBlockMarkerEnd)
	if re.MatchString(contentStr) {
		contentStr = re.ReplaceAllString(contentStr, blockBuilder.String())
	} else {
		contentStr += "\n" + blockBuilder.String()
	}

	return ioutil.WriteFile(PostfixMasterCfPath, []byte(contentStr), 0644)
}
