package askai

import (
	"billionmail-core/internal/service/public"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	PRODUCT_CONFIG_PATH = "../conf/askai"
	FILE_CDN_API        = "https://cdn.billionmail.com" // CDN API for file access
)

type KnowledgeInfo struct {
	Kid        string `json:"kid"`         // 知识库ID
	Title      string `json:"title"`       // 知识库标题
	Content    string `json:"content"`     // 知识库内容
	UpdateTime int64  `json:"update_time"` // 更新时间
}

type ProjectConfig struct {
	Domain        string          `json:"domain"`         // 域名
	Urls          []string        `json:"urls"`           // 相关URL列表
	ProjectName   string          `json:"project_name"`   // 项目名称
	Description   string          `json:"description"`    // 项目描述
	Industry      string          `json:"industry"`       // 行业
	PrimaryLogo   string          `json:"primary_logo"`   // 主Logo
	SecondaryLogo string          `json:"secondary_logo"` // 副Logo
	Favicon       string          `json:"favicon"`        // 网站图标
	KnowledgeBase []KnowledgeInfo `json:"knowledge_base"` // 知识库
	Status        bool            `json:"status"`         // 项目状态
	UpdateTime    int64           `json:"update_time"`    // 更新时间

}

// CompanyProfile represents the company profile information.
type CompanyProfile struct {
	LegalCompanyName string `json:"legal_company_name"` // 法定公司名称
	WebSite          string `json:"web_site"`           // 公司网站
	CompanyProfile   string `json:"company_profile"`    // 公司简介
	Email            string `json:"email"`              // 公司邮箱
	Phone            string `json:"phone"`              // 公司电话
	SupportUrl       string `json:"support_url"`        // 支持链接
	UpdateTime       int64  `json:"update_time"`        // 更新时间
}

// StyleConfig represents the style configuration for a project.
type StyleConfig struct {
	AccentColor         string `json:"accent_color"`         // 强调色
	TextColor           string `json:"text_color"`           // 文字颜色
	PageBackground      string `json:"page_background"`      // 页面背景色
	ContainerBackground string `json:"container_background"` // 容器背景色
	LinkSocialColor     string `json:"link_social_color"`    // 社交链接颜色
	LinkFooterColor     string `json:"link_footer_color"`    // 页脚链接颜色
	HeadingFont         string `json:"heading_font"`         // 标题字体
	BodyFont            string `json:"body_font"`            // 正文字体
	UpdateTime          int64  `json:"update_time"`          // 更新时间
}

// SiteMap represents a single entry in the site map of a project.
// It contains the title and URI path of the entry.
type SiteMap struct {
	Title string `json:"title"` // 网站标题
	// PageTitle    string `json:"page_title"`    // 页面标题
	// PageMarkdown string `json:"page_markdown"` // 页面Markdown内容
	UriPath    string `json:"uri_path"`    // URI路径
	UpdateTime int64  `json:"update_time"` // 更新时间
}

// FooterConfig represents the footer configuration for a project.
type FooterConfig struct {
	CopyrightText string `json:"copyright_text"` // 页脚版权文本
	Disclaimer    string `json:"disclaimer"`     // 免责声明
	Text          string `json:"text"`           // 页脚文本内容
	UpdateTime    int64  `json:"update_time"`    // 更新时间
}

type PromptConfig struct {
	Prompt     string `json:"prompt"`      // 提示内容
	UpdateTime int64  `json:"update_time"` // 更新时间
}

type ImageInfo struct {
	ImageId    string `json:"image_id"`    // 图片ID
	ImageUrl   string `json:"image_url"`   // 图片URL
	Filename   string `json:"filename"`    // 图片文件名
	AltText    string `json:"alt_text"`    // 图片替代文本
	ImageTag   string `json:"image_tag"`   // 图片标签
	UpdateTime int64  `json:"update_time"` // 更新时间
	Size       string `json:"size"`        // 图片大小 80x80
}

// ReadProjectConfig reads the project configuration from a JSON file based on the provided domain.
// It returns a ProjectConfig struct or an error if the file does not exist or cannot be read.
func ReadProjectConfig(Domain string) (ProjectConfig, error) {

	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/project.json", Domain)
	// Here you would implement the logic to read the project configuration from the file.
	// For now, we will just return a placeholder string.
	if !public.FileExists(filename) {
		return ProjectConfig{}, fmt.Errorf("project configuration file does not exist: %s", filename)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("error reading project configuration file: %v", err)
	}

	var config ProjectConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("error unmarshalling project configuration: %v", err)
	}
	return config, nil
}

// SaveProjectConfig saves the provided ProjectConfig to a JSON file based on the domain.
// It creates the directory if it does not exist and writes the configuration to project.json.
// If the file cannot be written, it returns an error.
// If the directory does not exist, it creates it with appropriate permissions.
func SaveProjectConfig(Domain string, config ProjectConfig) error {
	projectConfigPath := fmt.Sprintf("%s/%s", PRODUCT_CONFIG_PATH, Domain)
	if !public.FileExists(projectConfigPath) {
		os.MkdirAll(projectConfigPath, os.ModePerm)
	}
	filename := fmt.Sprintf("%s/project.json", projectConfigPath)

	configStr, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling project configuration: %v", err)
	}
	config.UpdateTime = public.GetNowTime() // Update the time before saving
	err = os.WriteFile(filename, configStr, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing project configuration file: %v", err)
	}
	return nil
}

// Create initializes a new project configuration with the provided domain and URLs.
// It sets default values for other fields and saves the configuration to a file.
func Create(Domain string, urls []string) error {
	urlsCount := len(urls)
	if urls == nil || urlsCount == 0 {
		urls = append(urls, "http://"+Domain) // Default URL if none provided
	}
	if urlsCount > 3 {
		return fmt.Errorf("Add up to 3 URLs")
	}
	// Ensure all URLs start with "http://"
	for i, urlStr := range urls {
		if !strings.HasPrefix(urlStr, "http") {
			urls[i] = "http://" + urlStr
		}
	}

	var config ProjectConfig = ProjectConfig{
		Domain:        Domain,
		Urls:          urls,
		ProjectName:   "",
		Description:   "",
		Industry:      "",
		PrimaryLogo:   "",
		SecondaryLogo: "",
		Favicon:       "",
		Status:        true, // Default status is active
		KnowledgeBase: []KnowledgeInfo{},
	}

	err := SaveProjectConfig(Domain, config)
	if err != nil {
		return fmt.Errorf("error creating project configuration: %v", err)
	}

	return nil
}

// GetBaseInfo retrieves the base information of a project configuration based on the provided domain.
// It reads the project configuration and returns a ProjectConfig struct containing only the base information.
func GetBaseInfo(Domain string) (ProjectConfig, error) {
	config, err := ReadProjectConfig(Domain)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("error reading project configuration: %v", err)
	}

	// Return only the base information
	baseInfo := ProjectConfig{
		Domain:        config.Domain,
		Urls:          config.Urls,
		ProjectName:   config.ProjectName,
		Description:   config.Description,
		Industry:      config.Industry,
		PrimaryLogo:   config.PrimaryLogo,
		SecondaryLogo: config.SecondaryLogo,
		Favicon:       config.Favicon,
		UpdateTime:    config.UpdateTime,
		Status:        config.Status,
	}

	// Get the knowledge base list
	// This will populate the KnowledgeBase field in the baseInfo struct.
	baseInfo.KnowledgeBase, err = GetKnowledgeBaseList(Domain)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("error getting knowledge base list: %v", err)
	}

	return baseInfo, nil
}

// GetProjectStatus retrieves the status of a project configuration based on the provided domain.
// It reads the project configuration and returns a boolean indicating whether the project is active or not.
func GetProjectStatus(Domain string) (bool, bool, error) {
	config, err := ReadProjectConfig(Domain)
	if err != nil {
		return false, false, nil
	}

	// Return the project status
	return config.Status, true, nil
}

// SetProjectStatus updates the status of a project configuration based on the provided domain.
// It reads the existing configuration, modifies the status, and saves the updated configuration back to the file.
func SetProjectStatus(Domain string, status bool) error {
	config, err := ReadProjectConfig(Domain)
	if err != nil {
		return fmt.Errorf("error reading project configuration: %v", err)
	}

	// Update the project status
	config.Status = status

	err = SaveProjectConfig(Domain, config)
	if err != nil {
		return fmt.Errorf("error saving project configuration: %v", err)
	}
	return nil
}

// ModifyBaseInfo updates the base information of a project configuration based on the provided domain.
// It reads the existing configuration, modifies the specified fields, and saves the updated configuration back to the file.
// If any field is empty, it will not be updated.
func ModifyBaseInfo(Domain string, projectName, description, industry, primaryLogo, secondaryLogo, favicon string, urls []string) error {
	config, err := ReadProjectConfig(Domain)
	if err != nil {
		return fmt.Errorf("error reading project configuration: %v", err)
	}

	// Update the base information
	if projectName != "" {
		config.ProjectName = projectName
	}
	if description != "" {
		config.Description = description
	}
	if industry != "" {
		config.Industry = industry
	}
	if primaryLogo != "" {
		config.PrimaryLogo = primaryLogo
	}
	if secondaryLogo != "" {
		config.SecondaryLogo = secondaryLogo
	}
	if favicon != "" {
		config.Favicon = favicon
	}
	urlsCount := len(urls)
	if urlsCount > 0 {
		if urlsCount > 3 {
			return fmt.Errorf("Add up to 3 URLs")
		}
		config.Urls = urls
	}

	err = SaveProjectConfig(Domain, config)
	if err != nil {
		return fmt.Errorf("error saving project configuration: %v", err)
	}
	return nil
}

// ReadKnowledgeBase reads a specific knowledge base from a JSON file based on the provided domain and knowledge ID.
// It returns a KnowledgeInfo struct or an error if the file does not exist or cannot be
func ReadKnowledgeBase(Domain string, Kid string) (KnowledgeInfo, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/knowledge/%s.json", Domain, Kid)
	if !public.FileExists(filename) {
		return KnowledgeInfo{}, fmt.Errorf("knowledge base file does not exist: %s", filename)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return KnowledgeInfo{}, fmt.Errorf("error reading knowledge base file: %v", err)
	}

	var knowledge KnowledgeInfo
	err = json.Unmarshal(data, &knowledge)
	if err != nil {
		return KnowledgeInfo{}, fmt.Errorf("error unmarshalling knowledge base: %v", err)
	}
	return knowledge, nil
}

// SaveKnowledgeBase saves the provided KnowledgeInfo to a JSON file based on the domain and knowledge ID.
// It creates the directory if it does not exist and writes the knowledge information to a JSON file
func SaveKnowledgeBase(Domain string, knowledge KnowledgeInfo) error {
	knowledgePath := fmt.Sprintf("%s/%s/knowledge", PRODUCT_CONFIG_PATH, Domain)
	if !public.FileExists(knowledgePath) {
		os.MkdirAll(knowledgePath, os.ModePerm)
	}
	filename := fmt.Sprintf("%s/%s.json", knowledgePath, knowledge.Kid)

	configStr, err := json.MarshalIndent(knowledge, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling knowledge base: %v", err)
	}

	err = os.WriteFile(filename, configStr, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing knowledge base file: %v", err)
	}
	return nil
}

// GetKnowledgeBaseList retrieves a list of all knowledge bases for a given domain.
// It reads the knowledge base directory, iterates through the files, and returns a slice of KnowledgeInfo structs.
// If the directory does not exist or cannot be read, it returns an error.
func GetKnowledgeBaseList(Domain string) ([]KnowledgeInfo, error) {
	knowledgePath := fmt.Sprintf("%s/%s/knowledge", PRODUCT_CONFIG_PATH, Domain)
	if !public.FileExists(knowledgePath) {
		os.MkdirAll(knowledgePath, os.ModePerm)
		// If the knowledge base directory does not exist, create it
	}

	files, err := os.ReadDir(knowledgePath)
	if err != nil {
		return nil, fmt.Errorf("error reading knowledge base directory: %v", err)
	}

	var knowledgeList []KnowledgeInfo = []KnowledgeInfo{}
	for _, file := range files {
		if !file.IsDir() {
			knowledge, err := ReadKnowledgeBase(Domain, strings.Split(file.Name(), ".")[0])
			if err == nil {
				knowledgeList = append(knowledgeList, knowledge)
			}
		}
	}
	return knowledgeList, nil
}

// GetUUID generates a new UUID using the UUID version 7 and returns it as a string.
// UUID version 7 is a time-based UUID that is suitable for generating unique identifiers.
func GetUUID() string {
	return uuid.Must(uuid.NewV7()).String()

}

// CreateKnowledgeBaseFile creates a new knowledge base file with the provided domain, title, and content.
// It generates a unique ID for the knowledge base, constructs a KnowledgeInfo struct, and saves it using the SaveKnowledgeBase function.
func CreateKnowledgeBaseFile(Domain string, Title string, Content string, Tid string) error {
	if Tid == "" {
		Tid = GetUUID()
	}

	knowledge := KnowledgeInfo{
		Kid:        Tid,
		Title:      Title,
		Content:    Content,
		UpdateTime: public.GetNowTime(),
	}
	return SaveKnowledgeBase(Domain, knowledge)
}

// ModifyKnowledgeBaseFile updates the title and content of an existing knowledge base file.
// It reads the existing knowledge base, modifies the specified fields, and saves the updated knowledge base.
// If the knowledge base does not exist or cannot be read, it returns an error.
func ModifyKnowledgeBaseFile(Domain string, Kid string, Title string, Content string) error {
	knowledge, err := ReadKnowledgeBase(Domain, Kid)
	if err != nil {
		return fmt.Errorf("error reading knowledge base: %v", err)
	}

	// Update the knowledge base information
	if Title != "" {
		knowledge.Title = Title
	}
	if Content != "" {
		knowledge.Content = Content
	}

	knowledge.UpdateTime = public.GetNowTime()
	err = SaveKnowledgeBase(Domain, knowledge)
	if err != nil {
		return fmt.Errorf("error saving knowledge base: %v", err)
	}
	return nil
}

// RemoveKnowledgeBaseFile removes a knowledge base file based on the provided domain and knowledge ID.
// It constructs the file path, checks if the file exists, and removes it.
func RemoveKnowledgeBaseFile(Domain string, Kid string) error {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/knowledge/%s.json", Domain, Kid)
	if !public.FileExists(filename) {
		return fmt.Errorf("knowledge base file does not exist: %s", filename)
	}
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("error removing knowledge base file: %v", err)
	}
	return nil
}

// ReadCompanyProfile reads the company profile from a JSON file based on the provided domain.
// It returns a CompanyProfile struct or an error if the file does not exist or cannot be read.
func ReadCompanyProfile(Domain string) (CompanyProfile, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/company_profile.json", Domain)
	if !public.FileExists(filename) {
		companyProfileDefault := CompanyProfile{
			LegalCompanyName: "",
			WebSite:          "",
			CompanyProfile:   "",
			Email:            "",
			Phone:            "",
			SupportUrl:       "",
		}
		// If the company profile file does not exist, return a default profile
		// This allows the system to handle cases where the profile has not been set up yet
		// and avoids errors when trying to read a non-existent file.
		// It also allows the user to create a new profile without needing to handle file not
		// found errors.
		companyProfileDefault.UpdateTime = public.GetNowTime()
		companyProfileJson, err := json.MarshalIndent(companyProfileDefault, "", "  ")
		if err != nil {
			return CompanyProfile{}, fmt.Errorf("error marshalling company profile: %v", err)
		}
		err = os.WriteFile(filename, companyProfileJson, 0644)
		if err != nil {
			return CompanyProfile{}, fmt.Errorf("error creating company profile file: %v", err)
		}
		return companyProfileDefault, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return CompanyProfile{}, fmt.Errorf("error reading company profile file: %v", err)
	}
	var profile CompanyProfile
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return CompanyProfile{}, fmt.Errorf("error unmarshalling company profile: %v", err)
	}
	return profile, nil
}

// SaveCompanyProfile saves the provided CompanyProfile to a JSON file based on the domain.
// It creates the directory if it does not exist and writes the profile information to company_profile.json
func SaveCompanyProfile(Domain string, profile CompanyProfile) error {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/company_profile.json", Domain)
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("error marshalling company profile: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving company profile file: %v", err)
	}
	return nil
}

// GetCompanyProfile retrieves the company profile for a given domain.
// It reads the company profile from the file and returns a CompanyProfile struct.
// If the profile does not exist or cannot be read, it returns an error.
func GetCompanyProfile(Domain string) (CompanyProfile, error) {
	companyProfile, err := ReadCompanyProfile(Domain)
	if err != nil {
		return CompanyProfile{}, fmt.Errorf("error reading company profile: %v", err)
	}
	return companyProfile, nil
}

// ModifyCompanyProfile updates the company profile information based on the provided parameters.
// It reads the existing profile, modifies the specified fields, and saves the updated profile back to the file.
// If any field is empty, it will not be updated.
func ModifyCompanyProfile(Domain string, legalCompanyName, webSite, companyProfile, email, phone, supportUrl string) error {
	profile, err := ReadCompanyProfile(Domain)
	if err != nil {
		return fmt.Errorf("error reading company profile: %v", err)
	}

	// Update the company profile information
	if legalCompanyName != "" {
		profile.LegalCompanyName = legalCompanyName
	}
	if webSite != "" {
		profile.WebSite = webSite
	}
	if companyProfile != "" {
		profile.CompanyProfile = companyProfile
	}
	if email != "" {
		profile.Email = email
	}
	if phone != "" {
		profile.Phone = phone
	}
	if supportUrl != "" {
		profile.SupportUrl = supportUrl
	}

	profile.UpdateTime = public.GetNowTime()
	err = SaveCompanyProfile(Domain, profile)
	if err != nil {
		return fmt.Errorf("error saving company profile: %v", err)
	}
	return nil
}

// GetStyleConfig retrieves the style configuration for a given domain from a JSON file.
func GetStyleConfig(Domain string) (StyleConfig, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/style_config.json", Domain)
	if !public.FileExists(filename) {
		styleConfigDefault := StyleConfig{
			AccentColor:         "#007bff",           // Default accent color
			TextColor:           "#000000",           // Default text color
			PageBackground:      "#ffffff",           // Default page background color
			ContainerBackground: "#f8f9fa",           // Default container background color
			LinkSocialColor:     "#007bff",           // Default social link color
			LinkFooterColor:     "#6c757d",           // Default footer link color
			HeadingFont:         "Arial, sans-serif", // Default heading font
			BodyFont:            "Arial, sans-serif", // Default body font
			UpdateTime:          public.GetNowTime(), // Current time as update time
		}
		// If the style config file does not exist, return a default style config
		// This allows the system to handle cases where the style configuration has not been set up yet
		// and avoids errors when trying to read a non-existent file.
		styleConfigJson, err := json.MarshalIndent(styleConfigDefault, "", "  ")
		if err != nil {
			return StyleConfig{}, fmt.Errorf("error marshalling style config: %v", err)
		}
		err = os.WriteFile(filename, styleConfigJson, 0644)
		if err != nil {
			return StyleConfig{}, fmt.Errorf("error saving style config file: %v", err)
		}
		return styleConfigDefault, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return StyleConfig{}, fmt.Errorf("error reading style config file: %v", err)
	}

	var config StyleConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return StyleConfig{}, fmt.Errorf("error unmarshalling style config: %v", err)
	}
	return config, nil
}

func SaveStyleConfig(Domain string, config StyleConfig) error {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/style_config.json", Domain)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling style config: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving style config file: %v", err)
	}
	return nil
}

// ModifyStyleConfig saves the provided StyleConfig to a JSON file based on the domain.
// It creates the directory if it does not exist and writes the style configuration to style_config.json
func ModifyStyleConfig(Domain string, AccentColor string, TextColor string, PageBackground, ContainerBackground, LinkSocialColor, LinkFooterColor, HeadingFont, BodyFont string) error {
	config := StyleConfig{
		AccentColor:         AccentColor,
		TextColor:           TextColor,
		PageBackground:      PageBackground,
		ContainerBackground: ContainerBackground,
		LinkSocialColor:     LinkSocialColor,
		LinkFooterColor:     LinkFooterColor,
		HeadingFont:         HeadingFont,
		BodyFont:            BodyFont,
		UpdateTime:          public.GetNowTime(),
	}

	err := SaveStyleConfig(Domain, config)
	if err != nil {
		return fmt.Errorf("error saving style config file: %v", err)
	}
	return nil
}

// GetSiteMap retrieves the site map for a given domain from a JSON file.
// It reads the site map file and returns a slice of SiteMap structs.
// If the file does not exist or cannot be read, it returns an error.
func GetSiteMap(Domain string) ([]SiteMap, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/sitemap.json", Domain)
	if !public.FileExists(filename) {
		// If the site map file does not exist, return an empty slice
		// This allows the system to handle cases where the site map has not been set up
		// and avoids errors when trying to read a non-existent file.
		// It also allows the user to create a new site map without needing to handle file not found errors.
		emptySiteMap := []SiteMap{}
		emptySiteMapJson, err := json.MarshalIndent(emptySiteMap, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling empty site map: %v", err)
		}
		err = os.WriteFile(filename, emptySiteMapJson, 0644)
		if err != nil {
			return nil, fmt.Errorf("error saving empty site map file: %v", err)
		}
		return emptySiteMap, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading site map file: %v", err)
	}

	var siteMap []SiteMap
	err = json.Unmarshal(data, &siteMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling site map: %v", err)
	}
	return siteMap, nil
}

func SaveSiteMap(Domain string, siteMap []SiteMap) error {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/sitemap.json", Domain)
	data, err := json.MarshalIndent(siteMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling site map: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving site map file: %v", err)
	}
	return nil
}

// AddSiteMapNode adds a new site map node with the provided title and URI path to the site map of the specified domain.
// It reads the existing site map, appends the new node, and saves the updated site map back to the file.
// If the site map file does not exist or cannot be read, it returns an error.
func AddSiteMapNode(Domain string, Title string, UriPath string) error {
	siteMap, err := GetSiteMap(Domain)
	if err != nil {
		return fmt.Errorf("error getting site map: %v", err)
	}

	newNode := SiteMap{
		Title:      Title,
		UriPath:    UriPath,
		UpdateTime: public.GetNowTime(),
	}
	siteMap = append(siteMap, newNode)

	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/sitemap.json", Domain)
	data, err := json.MarshalIndent(siteMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling site map: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving site map file: %v", err)
	}
	return nil
}

// RemoveSiteMapNode removes a site map node based on the provided domain and URI path.
// It reads the existing site map, filters out the node with the specified URI path, and saves the updated site map back to the file.
// If the site map file does not exist or cannot be read, it returns an error.
func RemoveSiteMapNode(Domain string, UriPath string) error {
	siteMap, err := GetSiteMap(Domain)
	if err != nil {
		return fmt.Errorf("error getting site map: %v", err)
	}

	var updatedSiteMap []SiteMap
	for _, node := range siteMap {
		if node.UriPath != UriPath {
			updatedSiteMap = append(updatedSiteMap, node)
		}
	}

	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/sitemap.json", Domain)
	data, err := json.MarshalIndent(updatedSiteMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling site map: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving site map file: %v", err)
	}
	return nil
}

// GetFooter retrieves the footer configuration for a given domain from a JSON file.
// It reads the footer configuration file and returns a FooterConfig struct.
// If the file does not exist or cannot be read, it returns an error.
func GetFooter(Domain string) (FooterConfig, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/footer_config.json", Domain)
	if !public.FileExists(filename) {
		// If the footer config file does not exist, return a default footer config
		// This allows the system to handle cases where the footer configuration has not been set up
		// and avoids errors when trying to read a non-existent file.
		// It also allows the user to create a new footer configuration without needing to handle file
		// not found errors.
		defaultFooterConfig := FooterConfig{
			CopyrightText: "© 2023 Your Company. All rights reserved.",
			Disclaimer:    "This is a sample disclaimer.",
		}
		defaultFooterConfig.UpdateTime = public.GetNowTime()
		defaultFooterJson, err := json.MarshalIndent(defaultFooterConfig, "", "  ")
		if err != nil {
			return FooterConfig{}, fmt.Errorf("error marshalling default footer config: %v", err)
		}
		err = os.WriteFile(filename, defaultFooterJson, 0644)
		if err != nil {
			return FooterConfig{}, fmt.Errorf("error saving default footer config file: %v", err)
		}
		// Return the default footer config if the file does not exist
		return defaultFooterConfig, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return FooterConfig{}, fmt.Errorf("error reading footer config file: %v", err)
	}

	var config FooterConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return FooterConfig{}, fmt.Errorf("error unmarshalling footer config: %v", err)
	}
	return config, nil
}

// ModifyFooter updates the footer configuration for a given domain.
// It reads the existing footer configuration, modifies the specified fields, and saves the updated configuration back to the file.
// If any field is empty, it will not be updated.
func ModifyFooter(Domain string, CopyrightText, Disclaimer string, Text string) error {
	config, err := GetFooter(Domain)
	if err != nil {
		return fmt.Errorf("error getting footer config: %v", err)
	}

	// Update the footer configuration
	if CopyrightText != "" {
		config.CopyrightText = CopyrightText
	}
	if Disclaimer != "" {
		config.Disclaimer = Disclaimer
	}
	if Text != "" {
		config.Text = Text
	}

	config.UpdateTime = public.GetNowTime()
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/footer_config.json", Domain)
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling footer config: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving footer config file: %v", err)
	}
	return nil
}

// GetPrompt retrieves the prompt configuration for a given domain from a JSON file.
// It reads the prompt configuration file and returns a PromptConfig struct.
// If the file does not exist or cannot be read, it returns an error.
func GetPrompt(Domain string) (PromptConfig, error) {
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/prompt_config.json", Domain)
	if !public.FileExists(filename) {
		// If the prompt config file does not exist, return a default prompt config
		// This allows the system to handle cases where the prompt configuration has not been set up
		// and avoids errors when trying to read a non-existent file.
		// It also allows the user to create a new prompt configuration without needing to handle file
		// not found errors.
		defaultPromptConfig := PromptConfig{
			Prompt: "This is a default prompt. Please customize it.",
		}
		defaultPromptConfig.UpdateTime = public.GetNowTime()
		defaultPromptJson, err := json.MarshalIndent(defaultPromptConfig, "", "  ")

		if err != nil {
			return PromptConfig{}, fmt.Errorf("error marshalling default prompt config: %v", err)
		}
		err = os.WriteFile(filename, defaultPromptJson, 0644)
		if err != nil {
			return PromptConfig{}, fmt.Errorf("error saving default prompt config file: %v", err)
		}
		return defaultPromptConfig, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return PromptConfig{}, fmt.Errorf("error reading prompt config file: %v", err)
	}

	var config PromptConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return PromptConfig{}, fmt.Errorf("error unmarshalling prompt config: %v", err)
	}
	return config, nil
}

// ModifyPrompt updates the prompt configuration for a given domain.
// It reads the existing prompt configuration, modifies the specified fields, and saves the updated configuration back to the file.
// If the prompt is empty, it will not be updated.
func ModifyPrompt(Domain string, Prompt string) error {
	config, err := GetPrompt(Domain)
	if err != nil {
		return fmt.Errorf("error getting prompt config: %v", err)
	}

	// Update the prompt configuration
	if Prompt != "" {
		config.Prompt = Prompt
	}

	config.UpdateTime = public.GetNowTime()
	filename := fmt.Sprintf(PRODUCT_CONFIG_PATH+"/%s/prompt_config.json", Domain)
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling prompt config: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving prompt config file: %v", err)
	}
	return nil
}

type BotResponse struct {
	Data   BotData `json:"data"`
	Msg    string  `json:"msg"`
	Status bool    `json:"status"`
}

type BotData struct {
	Description   string       `json:"description"`
	Domain        string       `json:"domain"`
	Markdown      string       `json:"markdown"`
	Footer        BotFooter    `json:"footer"`
	Icon          string       `json:"icon"`
	Images        []ImageInfo  `json:"images"`
	Keywords      string       `json:"keywords"`
	PrimaryLogo   string       `json:"primary_logo"`
	SecondaryLogo string       `json:"secondary_logo"`
	Sitemap       []BotSitemap `json:"sitemap"`
	Source        string       `json:"source"`
	Style         BotStyle     `json:"style"`
	Title         string       `json:"title"`
	URL           string       `json:"url"`
	Email         string       `json:"email"`
	Phone         string       `json:"phone"`
	SupportUrl    string       `json:"support_url"`
}

type BotFooter struct {
	Copyright  string `json:"copyright"`
	Disclaimer string `json:"disclaimer"`
	Text       string `json:"text"`
}

type BotSitemap struct {
	PageMarkdown string `json:"page_markdown"`
	PageTitle    string `json:"page_title"`
	Title        string `json:"title"`
	URIPath      string `json:"uri_path"`
}

type BotStyle struct {
	AccentColor         string `json:"accent_color"`
	BodyFont            string `json:"body_font"`
	ContainerBackground string `json:"container_background"`
	HeadingFont         string `json:"heading_font"`
	LinkFooterColor     string `json:"link_footer_color"`
	LinkSocialColor     string `json:"link_social_color"`
	PageBackground      string `json:"page_background"`
	TextColor           string `json:"text_color"`
}

func SaveImagesConfig(Domain string, images []ImageInfo) error {
	imagesPath := fmt.Sprintf("%s/%s/images.json", PRODUCT_CONFIG_PATH, Domain)
	if !public.FileExists(PRODUCT_CONFIG_PATH + "/" + Domain) {
		os.MkdirAll(PRODUCT_CONFIG_PATH+"/"+Domain, os.ModePerm)
	}
	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling images: %v", err)
	}
	err = os.WriteFile(imagesPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving images file: %v", err)
	}
	return nil
}

func ReadImagesConfig(Domain string) ([]ImageInfo, error) {
	imagesPath := fmt.Sprintf("%s/%s/images.json", PRODUCT_CONFIG_PATH, Domain)
	if !public.FileExists(imagesPath) {
		os.WriteFile(imagesPath, []byte("[]"), 0644)
	}
	data, err := os.ReadFile(imagesPath)
	if err != nil {
		return []ImageInfo{}, fmt.Errorf("error reading images configuration file: %v", err)
	}

	var images []ImageInfo
	err = json.Unmarshal(data, &images)
	if err != nil {
		return []ImageInfo{}, fmt.Errorf("error unmarshalling images configuration: %v", err)
	}
	return images, nil
}

type UploadImageResponse struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		URL string `json:"url"`
	} `json:"data"`
}

// UploadImage(req.Domain, req.Image, req.Filename, req.AltText, req.ImageTag)
func UploadImage(Domain string, Image string, Filename string, AltText string, ImageTag string) (string, error) {
	imageInfo := ImageInfo{
		ImageId:    GetUUID(),
		ImageUrl:   "",
		Filename:   Filename,
		AltText:    AltText,
		ImageTag:   ImageTag,
		UpdateTime: public.GetNowTime(),
	}
	apiUrl := fmt.Sprintf("%s/upload-image", FILE_CDN_API)
	params := url.Values{}
	params.Set("imageId", imageInfo.ImageId)
	params.Set("imageBase64", Image)
	params.Set("imageFilename", Filename)

	result, err := public.HttpPostSrc(apiUrl, params, 60)
	if err != nil {
		return "", err
	}
	var resultData UploadImageResponse
	err = json.Unmarshal([]byte(result), &resultData)
	if err != nil {
		return "", err
	}

	if resultData.Status == false {
		return "", fmt.Errorf("upload image failed: %s", resultData.Msg)
	}

	imageInfo.ImageUrl = resultData.Data.URL

	// Save the image information to the images.json file
	images, err := ReadImagesConfig(Domain)
	if err != nil {
		return "", fmt.Errorf("error reading images configuration: %v", err)
	}
	images = append(images, imageInfo)
	err = SaveImagesConfig(Domain, images)
	if err != nil {
		return "", fmt.Errorf("error saving images configuration: %v", err)
	}

	return imageInfo.ImageUrl, nil
}

// ModifyImage updates the alt text and image tag of an existing image in the images configuration.
// It reads the existing images configuration, modifies the specified fields for the image with the given ID and saves the updated configuration back to the file.
// If the images configuration file does not exist or cannot be read, it returns an error.
func ModifyImage(Domain string, ImageId string, AltText string, ImageTag string) error {
	images, err := ReadImagesConfig(Domain)
	if err != nil {
		return fmt.Errorf("error reading images configuration: %v", err)
	}
	for i, image := range images {
		if image.ImageId == ImageId {
			// Update the AltText and ImageTag if they are not empty
			if AltText != "" {
				images[i].AltText = AltText
			}
			if ImageTag != "" {
				images[i].ImageTag = ImageTag
			}
			images[i].UpdateTime = public.GetNowTime()
			break
		}
	}
	err = SaveImagesConfig(Domain, images)
	if err != nil {
		return fmt.Errorf("error saving images configuration: %v", err)
	}
	return nil
}

// RemoveImage removes an image from the images configuration based on the provided domain and image ID.
// It reads the existing images configuration, filters out the image with the specified ID, and saves the updated configuration back to the file.
// If the images configuration file does not exist or cannot be read, it returns an error.
func RemoveImage(Domain string, ImageId string) error {
	images, err := ReadImagesConfig(Domain)
	if err != nil {
		return fmt.Errorf("error reading images configuration: %v", err)
	}

	for i, image := range images {
		if image.ImageId == ImageId {
			images = append(images[:i], images[i+1:]...)
			break
		}
	}
	err = SaveImagesConfig(Domain, images)
	if err != nil {
		return fmt.Errorf("error saving images configuration: %v", err)
	}
	// Optionally, you can also delete the image file from the CDN if needed
	apiUrl := fmt.Sprintf("%s/remove-image", FILE_CDN_API)
	params := url.Values{}
	params.Set("imageId", ImageId)
	go public.HttpPostSrc(apiUrl, params, 60)
	return nil
}

// GetImages retrieves the images configuration for a given domain.
// It reads the images.json file and returns a slice of ImageInfo structs.
// If the file does not exist or cannot be read, it returns an error.
func GetImages(Domain string) ([]ImageInfo, error) {
	images, err := ReadImagesConfig(Domain)
	if err != nil {
		return []ImageInfo{}, fmt.Errorf("error reading images configuration: %v", err)
	}
	return images, nil
}

func AppendImages(domain string, Images []ImageInfo) {
	images, err := ReadImagesConfig(domain)
	if err != nil {
		fmt.Printf("Error reading images configuration for %s: %v\n", domain, err)
		return
	}

	for _, image := range Images {
		exists := false
		for _, img := range images {
			if img.ImageUrl == image.ImageUrl {
				exists = true
				break
			}
		}
		if !exists && image.ImageUrl != "" {
			image.ImageId = GetUUID()
			image.UpdateTime = public.GetNowTime()
			images = append(images, image)
		}
	}
	err = SaveImagesConfig(domain, images)
	if err != nil {
		fmt.Printf("Error saving images configuration for %s: %v\n", domain, err)
		return
	}
}

func AppendSitemap(domain string, Sitemap []BotSitemap) {
	siteMap, _ := GetSiteMap(domain)
	for _, sitemapNode := range Sitemap {
		// check if the node already exists in the site map
		exists := false
		for _, existingNode := range siteMap {
			if existingNode.UriPath == sitemapNode.URIPath {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		newNode := SiteMap{
			Title: sitemapNode.Title,
			// PageTitle:    sitemapNode.PageTitle,
			// PageMarkdown: sitemapNode.PageMarkdown,
			UriPath:    sitemapNode.URIPath,
			UpdateTime: public.GetNowTime(),
		}
		siteMap = append(siteMap, newNode)
	}
	SaveSiteMap(domain, siteMap)
}

func AppendStyle(domain string, Style BotStyle) {
	styleConfig, _ := GetStyleConfig(domain)
	styleConfig.AccentColor = Style.AccentColor
	styleConfig.TextColor = Style.TextColor
	styleConfig.PageBackground = Style.PageBackground
	styleConfig.ContainerBackground = Style.ContainerBackground
	styleConfig.LinkSocialColor = Style.LinkSocialColor
	styleConfig.LinkFooterColor = Style.LinkFooterColor
	styleConfig.HeadingFont = Style.HeadingFont
	styleConfig.BodyFont = Style.BodyFont
	styleConfig.UpdateTime = public.GetNowTime()
	SaveStyleConfig(domain, styleConfig)
}

func AppendCompanyConfig(domain string, Data BotData) {
	companyConfig, _ := GetCompanyProfile(domain)
	if companyConfig.LegalCompanyName == "" {
		companyConfig.LegalCompanyName = Data.Title
	}
	if companyConfig.WebSite == "" {
		companyConfig.WebSite = Data.URL
	}
	if companyConfig.CompanyProfile == "" {
		companyConfig.CompanyProfile = Data.Description
	}
	if companyConfig.Email == "" {
		companyConfig.Email = Data.Email
	}

	if companyConfig.Phone == "" {
		companyConfig.Phone = Data.Phone
	}
	if companyConfig.SupportUrl == "" {
		companyConfig.SupportUrl = Data.SupportUrl
	}
	companyConfig.UpdateTime = public.GetNowTime()
	SaveCompanyProfile(domain, companyConfig)
}

func AppendProjectConfig(domain string, config ProjectConfig, Data BotData) {
	config.Description = Data.Description
	config.Favicon = Data.Icon
	config.PrimaryLogo = Data.PrimaryLogo
	config.SecondaryLogo = Data.SecondaryLogo
	config.ProjectName = Data.Title
	config.UpdateTime = public.GetNowTime()
	SaveProjectConfig(domain, config)
}

func GetUrlIsRequest(domain string, urlMd5 string) bool {
	tipsPath := fmt.Sprintf("%s/%s/tips", PRODUCT_CONFIG_PATH, domain)
	if !public.FileExists(tipsPath) {
		os.MkdirAll(tipsPath, 0755)
	}

	tipsFile := fmt.Sprintf("%s/%s", tipsPath, urlMd5)
	return public.FileExists(tipsFile)
}

func SetUrlIsRequest(domain string, urlMd5 string) {
	tipsPath := fmt.Sprintf("%s/%s/tips", PRODUCT_CONFIG_PATH, domain)
	if !public.FileExists(tipsPath) {
		os.MkdirAll(tipsPath, 0755)
	}
	tipsFile := fmt.Sprintf("%s/%s", tipsPath, urlMd5)
	os.WriteFile(tipsFile, []byte(public.GetNowTimeStr()), 0600)
}

func RequestUrl(domain, toUrl string) (BotResponse, error) {
	if !strings.HasPrefix(toUrl, "http") {
		toUrl = "http://" + toUrl
	}
	var result BotResponse
	url := FILE_CDN_API + "/bot?url=" + toUrl
	resultBody, err := public.HttpGetSrc(url, 360)
	if err != nil {
		fmt.Printf("Error fetching project data for %s: %v\n", domain, err)
		return result, err
	}

	// Parse the JSON response
	err = json.Unmarshal([]byte(resultBody), &result)
	if err != nil {
		fmt.Printf("Error parsing project data for %s: %v\n", domain, err)
		return result, err
	}
	return result, nil
}

func AutoGetProjectInfo() {
	if err := os.MkdirAll(PRODUCT_CONFIG_PATH, 0755); err != nil {
		fmt.Println("Error creating project config directory:", err)
		return
	}
	dirs, err := os.ReadDir(PRODUCT_CONFIG_PATH)
	if err != nil {
		fmt.Println("Error reading project config directory:", err)
		return
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue // Skip files, only process directories
		}

		domain := dir.Name()
		projectConfigFile := fmt.Sprintf("%s/%s/project.json", PRODUCT_CONFIG_PATH, domain)
		if !public.FileExists(projectConfigFile) {
			continue // Skip if project.json does not exist for this domain
		}
		// Read the project configuration
		config, err := ReadProjectConfig(domain)
		if err != nil {
			fmt.Printf("Error reading project configuration for %s: %v\n", domain, err)
			continue
		}

		if config.UpdateTime > 0 {
			continue // Skip if the project configuration has already been updated
		}

		if len(config.Urls) == 0 {
			continue // Skip if there are no URLs to process
		}

		// Check if the project is already being processed
		// Use a cache to prevent multiple concurrent updates for the same domain
		isLock := public.GetCache(domain)
		if isLock != nil && isLock.(bool) {
			continue
		}
		public.SetCache(domain, true, 600) // Set a lock for 300 seconds

		for _, toUrl := range config.Urls {
			urlMd5 := public.Md5(toUrl)

			// check is requestd
			// if GetUrlIsRequest(domain, urlMd5) {
			// 	continue
			// }

			result, err := RequestUrl(domain, toUrl)
			if err != nil {
				continue
			}
			// if i == 0 {
			AppendProjectConfig(domain, config, result.Data)
			AppendCompanyConfig(domain, result.Data)
			AppendStyle(domain, result.Data.Style)
			if result.Data.Footer.Copyright != "" {
				ModifyFooter(domain, result.Data.Footer.Copyright, result.Data.Footer.Disclaimer, result.Data.Footer.Text)
			}
			// } else {
			// 	for i, sitemapNode := range result.Data.Sitemap {
			// 		if strings.HasPrefix(sitemapNode.URIPath, "http") {
			// 			continue
			// 		}
			// 		result.Data.Sitemap[i].URIPath = strings.Replace(result.Data.SupportUrl+result.Data.Sitemap[i].URIPath, "//", "/", 1)
			// 	}
			// }

			AppendSitemap(domain, result.Data.Sitemap)
			CreateKnowledgeBaseFile(domain, result.Data.Title, result.Data.Markdown, urlMd5)
			AppendImages(domain, result.Data.Images)

			// tip requestd
			// SetUrlIsRequest(domain, urlMd5)
			break // Break after processing the first URL
		}

		// Clear the cache for this domain
		public.RemoveCache(domain)

	}
}
