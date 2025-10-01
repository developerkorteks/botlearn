// Package services - XRay Converter service untuk conversion dan modification XRay links
package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/nabilulilalbab/promote/database"
	"github.com/nabilulilalbab/promote/utils"
)

// XRayConverterService mengelola konversi dan modifikasi XRay links
type XRayConverterService struct {
	repository database.Repository
	logger     *utils.Logger
}

// NewXRayConverterService membuat service baru untuk XRay converter
func NewXRayConverterService(repo database.Repository, logger *utils.Logger) *XRayConverterService {
	return &XRayConverterService{
		repository: repo,
		logger:     logger,
	}
}

// DetectXRayConfig mendeteksi dan parse konfigurasi dari XRay link
func (s *XRayConverterService) DetectXRayConfig(xrayLink string) (*database.DetectedXRayConfig, error) {
	// Trim whitespace dan newlines
	xrayLink = strings.TrimSpace(xrayLink)
	
	// Deteksi protocol dari prefix
	if strings.HasPrefix(xrayLink, "vmess://") {
		return s.parseVMESS(xrayLink)
	} else if strings.HasPrefix(xrayLink, "vless://") {
		return s.parseVLESS(xrayLink)
	} else if strings.HasPrefix(xrayLink, "trojan://") {
		return s.parseTrojan(xrayLink)
	} else if strings.HasPrefix(xrayLink, "ss://") {
		return s.parseShadowsocks(xrayLink)
	}
	
	return nil, fmt.Errorf("unsupported protocol or invalid XRay link")
}

// parseVMESS parsing VMESS link (supports both JSON base64 and URL format)
func (s *XRayConverterService) parseVMESS(vmessLink string) (*database.DetectedXRayConfig, error) {
	// Remove vmess:// prefix
	linkData := strings.TrimPrefix(vmessLink, "vmess://")
	
	// Check if this is URL format (contains @ symbol)
	if strings.Contains(linkData, "@") {
		// This is URL format VMESS, parse as URL
		return s.parseURLFormat(vmessLink, "vmess")
	}
	
	// This is traditional JSON base64 format
	// Decode base64
	jsonData, err := base64.StdEncoding.DecodeString(linkData)
	if err != nil {
		// Try URL-safe base64
		jsonData, err = base64.URLEncoding.DecodeString(linkData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
	}
	
	// Parse JSON
	var vmessConfig map[string]interface{}
	err = json.Unmarshal(jsonData, &vmessConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	
	// Extract configuration
	config := &database.DetectedXRayConfig{
		Protocol:  "vmess",
		RawConfig: vmessConfig,
	}
	
	// Extract fields with type checking
	if v, ok := vmessConfig["add"].(string); ok {
		config.Server = v
	}
	if v, ok := vmessConfig["port"].(string); ok {
		if port, err := strconv.Atoi(v); err == nil {
			config.Port = port
		}
	} else if v, ok := vmessConfig["port"].(float64); ok {
		config.Port = int(v)
	}
	if v, ok := vmessConfig["id"].(string); ok {
		config.UUID = v
	}
	if v, ok := vmessConfig["net"].(string); ok {
		config.Network = v
	}
	if v, ok := vmessConfig["tls"].(string); ok {
		config.TLS = (v == "tls")
	}
	if v, ok := vmessConfig["sni"].(string); ok {
		config.SNI = v
	}
	if v, ok := vmessConfig["host"].(string); ok {
		config.Host = v
	}
	if v, ok := vmessConfig["path"].(string); ok {
		config.Path = v
	}
	if v, ok := vmessConfig["aid"].(string); ok {
		if aid, err := strconv.Atoi(v); err == nil {
			config.AlterID = aid
		}
	} else if v, ok := vmessConfig["aid"].(float64); ok {
		config.AlterID = int(v)
	}
	if v, ok := vmessConfig["ps"].(string); ok {
		config.Remarks = v
	}
	
	// gRPC service name extraction
	if config.Network == "grpc" {
		if v, ok := vmessConfig["path"].(string); ok {
			config.ServiceName = v
		}
	}
	
	return config, nil
}

// parseVLESS parsing VLESS link (URL format)
func (s *XRayConverterService) parseVLESS(vlessLink string) (*database.DetectedXRayConfig, error) {
	// VLESS format: vless://uuid@server:port?parameters#remarks
	return s.parseURLFormat(vlessLink, "vless")
}

// parseTrojan parsing Trojan link (URL format)
func (s *XRayConverterService) parseTrojan(trojanLink string) (*database.DetectedXRayConfig, error) {
	// Trojan format: trojan://password@server:port?parameters#remarks
	return s.parseURLFormat(trojanLink, "trojan")
}

// parseShadowsocks parsing Shadowsocks link
func (s *XRayConverterService) parseShadowsocks(ssLink string) (*database.DetectedXRayConfig, error) {
	// SS format: ss://base64(method:password)@server:port#remarks
	return s.parseURLFormat(ssLink, "shadowsocks")
}

// parseURLFormat parsing URL format (VLESS, Trojan, Shadowsocks)
func (s *XRayConverterService) parseURLFormat(linkURL, protocol string) (*database.DetectedXRayConfig, error) {
	// Parse URL
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	
	config := &database.DetectedXRayConfig{
		Protocol: protocol,
		RawConfig: make(map[string]interface{}),
	}
	
	// Extract basic info
	config.Server = parsedURL.Hostname()
	config.Port, _ = strconv.Atoi(parsedURL.Port())
	if config.Port == 0 {
		config.Port = 443 // Default port
	}
	
	// Extract UUID/Password from user info
	if parsedURL.User != nil {
		config.UUID = parsedURL.User.Username()
	}
	
	// Extract remarks from fragment
	config.Remarks = parsedURL.Fragment
	
	// Parse query parameters
	queryParams := parsedURL.Query()
	
	// Network type
	config.Network = queryParams.Get("type")
	if config.Network == "" {
		config.Network = "tcp" // Default
	}
	
	// TLS settings
	security := queryParams.Get("security")
	config.TLS = (security == "tls")
	
	// SNI
	config.SNI = queryParams.Get("sni")
	if config.SNI == "" {
		config.SNI = queryParams.Get("host")
	}
	
	// Host
	config.Host = queryParams.Get("host")
	if config.Host == "" {
		config.Host = config.Server
	}
	
	// Path (for WebSocket)
	config.Path = queryParams.Get("path")
	if config.Path == "" && (config.Network == "ws" || config.Network == "httpupgrade") {
		config.Path = parsedURL.Path
	}
	// Handle multiple levels of URL encoding (e.g., %252F -> %2F -> /vless)
	if config.Path != "" {
		// Decode until we get the actual path
		decodedPath := config.Path
		for strings.Contains(decodedPath, "%") {
			if newPath, err := url.QueryUnescape(decodedPath); err == nil && newPath != decodedPath {
				decodedPath = newPath
			} else {
				break
			}
		}
		config.Path = decodedPath
	}
	
	// gRPC service name
	config.ServiceName = queryParams.Get("serviceName")
	if config.ServiceName == "" {
		config.ServiceName = queryParams.Get("service")
	}
	
	// KCP header type
	config.HeaderType = queryParams.Get("headerType")
	if config.HeaderType == "" {
		config.HeaderType = queryParams.Get("header")
	}
	
	// Build raw config for reconstruction
	rawConfig := map[string]interface{}{
		"protocol": protocol,
		"server":   config.Server,
		"port":     fmt.Sprintf("%d", config.Port),
		"uuid":     config.UUID,
		"network":  config.Network,
		"security": security,
		"remarks":  config.Remarks,
	}
	
	if config.TLS {
		rawConfig["tls"] = "tls"
		rawConfig["sni"] = config.SNI
	}
	
	if config.Host != "" {
		rawConfig["host"] = config.Host
	}
	
	if config.Path != "" {
		rawConfig["path"] = config.Path
	}
	
	if config.ServiceName != "" {
		rawConfig["serviceName"] = config.ServiceName
	}
	
	if config.HeaderType != "" {
		rawConfig["headerType"] = config.HeaderType
	}
	
	// Add all query parameters to raw config
	for key, values := range queryParams {
		if len(values) > 0 {
			rawConfig[key] = values[0]
		}
	}
	
	config.RawConfig = rawConfig
	
	return config, nil
}

// ModifyXRayConfig melakukan modifikasi berdasarkan converter rules
func (s *XRayConverterService) ModifyXRayConfig(detected *database.DetectedXRayConfig, converter *database.XRayConverter) (*database.ModifiedXRayConfig, error) {
	// Clone original config
	modifiedConfig := make(map[string]interface{})
	for k, v := range detected.RawConfig {
		modifiedConfig[k] = v
	}
	
	result := &database.ModifiedXRayConfig{
		DetectedConfig: detected,
		ModifyType:     converter.ModifyType,
		BugHost:        converter.BugHost,
	}
	
	// Process templates with fallback to legacy modify types
	result.ModifiedServer = s.processTemplate(converter.ServerTemplate, converter, detected)
	result.ModifiedHost = s.processTemplate(converter.HostTemplate, converter, detected)
	result.ModifiedSNI = s.processTemplate(converter.SNITemplate, converter, detected)
	
	// Fallback to legacy modify types if templates are empty
	if converter.ServerTemplate == "" && converter.HostTemplate == "" && converter.SNITemplate == "" {
		switch converter.ModifyType {
		case "wildcard":
			result.ModifiedServer = converter.BugHost
			result.ModifiedHost = fmt.Sprintf("%s.%s", converter.BugHost, detected.Server)
			result.ModifiedSNI = fmt.Sprintf("%s.%s", converter.BugHost, detected.Server)
		case "sni":
			result.ModifiedServer = detected.Server
			result.ModifiedHost = detected.Host
			result.ModifiedSNI = converter.BugHost
		case "ws", "grpc":
			result.ModifiedServer = converter.BugHost
			result.ModifiedHost = detected.Host
			result.ModifiedSNI = detected.SNI
		}
	}
	
	// Update config based on protocol and format
	if detected.Protocol == "vmess" && isVMESSJSONFormat(detected.RawConfig) {
		// VMESS JSON format
		modifiedConfig["add"] = result.ModifiedServer
		modifiedConfig["host"] = result.ModifiedHost
		if detected.TLS {
			modifiedConfig["sni"] = result.ModifiedSNI
		}
	} else {
		// URL format (VLESS, Trojan, VMESS URL)
		modifiedConfig["server"] = result.ModifiedServer
		modifiedConfig["host"] = result.ModifiedHost
		if detected.TLS {
			modifiedConfig["sni"] = result.ModifiedSNI
		}
	}
	
	// Apply path template only for specific modify types that need it
	// For wildcard and sni, keep original path unless specifically needed
	if converter.PathTemplate != "" && (detected.Network == "ws" || detected.Network == "httpupgrade" || detected.Network == "h2") {
		// Only apply path template for "ws" and "grpc" modify types
		// For "wildcard" and "sni", keep original path to preserve user's configuration
		if converter.ModifyType == "ws" || converter.ModifyType == "grpc" {
			modifiedConfig["path"] = converter.PathTemplate
		}
		// For wildcard and sni, path stays original (already in modifiedConfig)
	}
	
	// Apply port override if provided
	if converter.PortOverride != nil {
		modifiedConfig["port"] = strconv.Itoa(*converter.PortOverride)
	}
	
	// Generate new link based on protocol and format
	switch detected.Protocol {
	case "vmess":
		if isVMESSJSONFormat(detected.RawConfig) {
			// VMESS JSON format
			newLink, err := s.generateVMESSLink(modifiedConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to generate VMESS link: %v", err)
			}
			result.ModifiedLink = newLink
		} else {
			// VMESS URL format
			newLink, err := s.generateURLFormatLink(modifiedConfig, detected.Protocol)
			if err != nil {
				return nil, fmt.Errorf("failed to generate VMESS URL link: %v", err)
			}
			result.ModifiedLink = newLink
		}
	case "vless", "trojan", "shadowsocks":
		newLink, err := s.generateURLFormatLink(modifiedConfig, detected.Protocol)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s link: %v", detected.Protocol, err)
		}
		result.ModifiedLink = newLink
	default:
		return nil, fmt.Errorf("link generation for %s not implemented yet", detected.Protocol)
	}
	
	// Generate YAML config
	yamlConfig, err := s.generateYAMLConfig(detected, modifiedConfig, converter)
	if err != nil {
		return nil, fmt.Errorf("failed to generate YAML config: %v", err)
	}
	result.YAMLConfig = yamlConfig
	
	return result, nil
}

// generateVMESSLink generate VMESS link dari modified config
func (s *XRayConverterService) generateVMESSLink(config map[string]interface{}) (string, error) {
	// Convert to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	
	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(jsonData)
	
	return "vmess://" + base64Data, nil
}

// generateURLFormatLink generate URL format link (VLESS, Trojan, Shadowsocks)
func (s *XRayConverterService) generateURLFormatLink(config map[string]interface{}, protocol string) (string, error) {
	// Extract values from config
	server := getString(config, "server")
	port := getString(config, "port")
	uuid := getString(config, "uuid")
	remarks := getString(config, "remarks")
	
	if server == "" || port == "" || uuid == "" {
		return "", fmt.Errorf("missing required fields for %s link", protocol)
	}
	
	// Build URL
	linkURL := fmt.Sprintf("%s://%s@%s:%s", protocol, uuid, server, port)
	
	// Build query parameters
	params := url.Values{}
	
	// Add network type
	if network := getString(config, "network"); network != "" && network != "tcp" {
		params.Set("type", network)
	}
	
	// Add security
	if security := getString(config, "security"); security != "" {
		params.Set("security", security)
	}
	
	// Add TLS-related parameters
	if getBool(config, "tls") || getString(config, "security") == "tls" {
		params.Set("security", "tls")
		if sni := getString(config, "sni"); sni != "" {
			params.Set("sni", sni)
		}
	}
	
	// Add host
	if host := getString(config, "host"); host != "" && host != server {
		params.Set("host", host)
	}
	
	// Add path (for WebSocket/HTTPUpgrade) - smart encoding
	if path := getString(config, "path"); path != "" {
		// Always use the decoded path and let URL encoding handle it properly
		params.Set("path", path)
	}
	
	// Add service name (for gRPC)
	if serviceName := getString(config, "serviceName"); serviceName != "" {
		params.Set("serviceName", serviceName)
	}
	
	// Add other parameters from original config
	for key, value := range config {
		keyStr := fmt.Sprintf("%v", key)
		valueStr := fmt.Sprintf("%v", value)
		
		// Skip already handled parameters
		switch keyStr {
		case "protocol", "server", "port", "uuid", "remarks", "network", "security", "tls", "sni", "host", "path", "serviceName":
			continue
		}
		
		// Add other parameters
		if valueStr != "" && valueStr != "0" && valueStr != "false" {
			params.Set(keyStr, valueStr)
		}
	}
	
	// Add query parameters to URL manually to avoid double encoding
	if len(params) > 0 {
		queryParts := make([]string, 0, len(params))
		for key, values := range params {
			for _, value := range values {
				// Special handling for path to avoid re-encoding
				if key == "path" {
					queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
				} else {
					queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
				}
			}
		}
		linkURL += "?" + strings.Join(queryParts, "&")
	}
	
	// Add remarks as fragment
	if remarks != "" {
		linkURL += "#" + url.QueryEscape(remarks)
	}
	
	return linkURL, nil
}

// Helper functions for type conversion
func getString(config map[string]interface{}, key string) string {
	if val, ok := config[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func getBool(config map[string]interface{}, key string) bool {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "tls"
		}
	}
	return false
}

// processTemplate memproses template dengan placeholder
func (s *XRayConverterService) processTemplate(template string, converter *database.XRayConverter, detected *database.DetectedXRayConfig) string {
	if template == "" {
		return ""
	}
	
	// For now, bug_ip same as bug_host (DNS resolution can be added later)
	bugIP := converter.BugHost
	
	// Available placeholders:
	// {bug_host} - Bug host domain
	// {bug_ip} - Bug host IP (same as domain for now, can be enhanced)
	// {original_server} - Original server from XRay link
	// {original_host} - Original host from XRay link
	// {original_sni} - Original SNI from XRay link
	
	result := template
	result = strings.ReplaceAll(result, "{bug_host}", converter.BugHost)
	result = strings.ReplaceAll(result, "{bug_ip}", bugIP)
	result = strings.ReplaceAll(result, "{original_server}", detected.Server)
	result = strings.ReplaceAll(result, "{original_host}", detected.Host)
	result = strings.ReplaceAll(result, "{original_sni}", detected.SNI)
	
	return result
}

// isVMESSJSONFormat check if VMESS config is JSON format (has "add" field)
func isVMESSJSONFormat(config map[string]interface{}) bool {
	_, hasAdd := config["add"]
	return hasAdd
}

// generateYAMLConfig generate YAML config untuk Clash/OpenClash
func (s *XRayConverterService) generateYAMLConfig(detected *database.DetectedXRayConfig, modifiedConfig map[string]interface{}, converter *database.XRayConverter) (string, error) {
	var yamlBuilder strings.Builder
	
	// Proxy name
	proxyName := fmt.Sprintf("%s-%s-%d", converter.DisplayName, detected.Protocol, detected.Port)
	
	yamlBuilder.WriteString("proxies:\n")
	yamlBuilder.WriteString(fmt.Sprintf("  - name: \"%s\"\n", proxyName))
	yamlBuilder.WriteString(fmt.Sprintf("    type: %s\n", detected.Protocol))
	yamlBuilder.WriteString("    udp: true\n")
	
	// Server and port
	if server, ok := modifiedConfig["add"].(string); ok {
		yamlBuilder.WriteString(fmt.Sprintf("    server: %s\n", server))
	} else if server, ok := modifiedConfig["server"].(string); ok {
		yamlBuilder.WriteString(fmt.Sprintf("    server: %s\n", server))
	}
	
	port := detected.Port
	if converter.PortOverride != nil {
		port = *converter.PortOverride
	}
	yamlBuilder.WriteString(fmt.Sprintf("    port: %d\n", port))
	
	// Protocol specific configs
	switch detected.Protocol {
	case "vmess":
		if uuid, ok := modifiedConfig["id"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("    uuid: %s\n", uuid))
		}
		yamlBuilder.WriteString(fmt.Sprintf("    alterId: %d\n", detected.AlterID))
		yamlBuilder.WriteString("    cipher: auto\n")
	case "vless":
		if uuid, ok := modifiedConfig["uuid"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("    uuid: %s\n", uuid))
		}
	case "trojan":
		if password, ok := modifiedConfig["uuid"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("    password: %s\n", password))
		}
	case "shadowsocks":
		if password, ok := modifiedConfig["uuid"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("    password: %s\n", password))
		}
		if cipher, ok := modifiedConfig["cipher"].(string); ok && cipher != "" {
			yamlBuilder.WriteString(fmt.Sprintf("    cipher: %s\n", cipher))
		}
	}
	
	// Network config
	yamlBuilder.WriteString(fmt.Sprintf("    network: %s\n", detected.Network))
	
	// TLS config
	if detected.TLS {
		yamlBuilder.WriteString("    tls: true\n")
		if sni, ok := modifiedConfig["sni"].(string); ok && sni != "" {
			yamlBuilder.WriteString(fmt.Sprintf("    servername: %s\n", sni))
		}
		yamlBuilder.WriteString("    skip-cert-verify: true\n")
	} else {
		yamlBuilder.WriteString("    tls: false\n")
	}
	
	// Network specific options
	switch detected.Network {
	case "ws":
		yamlBuilder.WriteString("    ws-opts:\n")
		if path, ok := modifiedConfig["path"].(string); ok && path != "" {
			yamlBuilder.WriteString(fmt.Sprintf("      path: %s\n", path))
		}
		if host, ok := modifiedConfig["host"].(string); ok && host != "" {
			yamlBuilder.WriteString("      headers:\n")
			yamlBuilder.WriteString(fmt.Sprintf("        Host: %s\n", host))
		}
		
	case "grpc":
		yamlBuilder.WriteString("    grpc-opts:\n")
		serviceName := "grpc-service"
		if path, ok := modifiedConfig["path"].(string); ok && path != "" {
			serviceName = path
		}
		yamlBuilder.WriteString(fmt.Sprintf("      grpc-service-name: \"%s\"\n", serviceName))
		
	case "httpupgrade":
		yamlBuilder.WriteString("    httpupgrade-opts:\n")
		if path, ok := modifiedConfig["path"].(string); ok && path != "" {
			yamlBuilder.WriteString(fmt.Sprintf("      path: %s\n", path))
		}
		if host, ok := modifiedConfig["host"].(string); ok && host != "" {
			yamlBuilder.WriteString("      headers:\n")
			yamlBuilder.WriteString(fmt.Sprintf("        Host: %s\n", host))
		}
	}
	
	return yamlBuilder.String(), nil
}

// ProcessConversion memproses conversion lengkap dari XRay link
func (s *XRayConverterService) ProcessConversion(converterName, xrayLink, userJID, groupJID string) (*database.ModifiedXRayConfig, error) {
	// Get converter config
	converter, err := s.repository.GetXRayConverter(converterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get converter: %v", err)
	}
	
	if converter == nil {
		return nil, fmt.Errorf("converter not found: %s", converterName)
	}
	
	if !converter.IsActive {
		return nil, fmt.Errorf("converter is inactive: %s", converterName)
	}
	
	// Detect XRay config
	detected, err := s.DetectXRayConfig(xrayLink)
	if err != nil {
		// Log conversion failure
		errMsg := err.Error()
		logEntry := &database.XRayConversionLog{
			ConverterName:    converterName,
			UserJID:          userJID,
			GroupJID:         groupJID,
			OriginalProtocol: "unknown",
			OriginalNetwork:  "",
			OriginalServer:   "",
			ModifiedServer:   "",
			Success:          false,
			ErrorMessage:     &errMsg,
		}
		s.repository.LogXRayConversion(logEntry)
		
		return nil, fmt.Errorf("failed to detect XRay config: %v", err)
	}
	
	// Modify config
	result, err := s.ModifyXRayConfig(detected, converter)
	if err != nil {
		// Log conversion failure
		errMsg := err.Error()
		logEntry := &database.XRayConversionLog{
			ConverterName:    converterName,
			UserJID:          userJID,
			GroupJID:         groupJID,
			OriginalProtocol: detected.Protocol,
			OriginalNetwork:  detected.Network,
			OriginalServer:   detected.Server,
			ModifiedServer:   "",
			Success:          false,
			ErrorMessage:     &errMsg,
		}
		s.repository.LogXRayConversion(logEntry)
		
		return nil, fmt.Errorf("failed to modify XRay config: %v", err)
	}
	
	// Log successful conversion
	logEntry := &database.XRayConversionLog{
		ConverterName:    converterName,
		UserJID:          userJID,
		GroupJID:         groupJID,
		OriginalProtocol: detected.Protocol,
		OriginalNetwork:  detected.Network,
		OriginalServer:   detected.Server,
		ModifiedServer:   result.ModifiedServer,
		Success:          true,
	}
	s.repository.LogXRayConversion(logEntry)
	
	// Increment usage count
	s.repository.IncrementConverterUsage(converterName)
	
	s.logger.Infof("XRay conversion successful: %s -> %s (%s)", detected.Server, result.ModifiedServer, converter.ModifyType)
	
	return result, nil
}

// GetAllConverters mendapatkan semua converter yang tersedia
func (s *XRayConverterService) GetAllConverters() ([]database.XRayConverter, error) {
	return s.repository.GetAllXRayConverters()
}

// GetActiveConverters mendapatkan converter yang aktif
func (s *XRayConverterService) GetActiveConverters() ([]database.XRayConverter, error) {
	return s.repository.GetActiveXRayConverters()
}

// CreateConverter membuat converter baru
func (s *XRayConverterService) CreateConverter(converter *database.XRayConverter) error {
	return s.repository.CreateXRayConverter(converter)
}

// UpdateConverter update converter
func (s *XRayConverterService) UpdateConverter(converter *database.XRayConverter) error {
	return s.repository.UpdateXRayConverter(converter)
}

// DeleteConverter hapus converter
func (s *XRayConverterService) DeleteConverter(commandName string) error {
	return s.repository.DeleteXRayConverter(commandName)
}

// GetConversionStats mendapatkan statistik conversion
func (s *XRayConverterService) GetConversionStats(days int) (map[string]int, error) {
	return s.repository.GetXRayConversionStats(days)
}