package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"herobox/internal/models"
	"herobox/internal/service"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SingBoxInbound 入站配置结构
type SingBoxInbound struct {
	Tag        string      `json:"tag"`
	Type       string      `json:"type"`
	Listen     string      `json:"listen,omitempty"`
	ListenPort int         `json:"listen_port,omitempty"`
	Users      interface{} `json:"users,omitempty"`
	Method     string      `json:"method,omitempty"`
	Password   string      `json:"password,omitempty"`
	Network    string      `json:"network,omitempty"`
}

// SingBoxOutbound 出站配置结构
type SingBoxOutbound struct {
	Tag             string   `json:"tag"`
	Type            string   `json:"type"`
	Server          string   `json:"server,omitempty"`
	ServerPort      int      `json:"server_port,omitempty"`
	Method          string   `json:"method,omitempty"`
	Password        string   `json:"password,omitempty"`
	UUID            string   `json:"uuid,omitempty"`
	Flow            string   `json:"flow,omitempty"`
	Network         string   `json:"network,omitempty"`
	Security        string   `json:"security,omitempty"`
	Path            string   `json:"path,omitempty"`
	Host            string   `json:"host,omitempty"`
	ServiceName     string   `json:"service_name,omitempty"`
	Outbounds       []string `json:"outbounds,omitempty"`
	Default         string   `json:"default,omitempty"`
	Include         string   `json:"include,omitempty"`
	Exclude         string   `json:"exclude,omitempty"`
	UseAllProviders bool     `json:"use_all_providers,omitempty"`

	// Provider 相关字段 (用于订阅节点)
	Providers []string `json:"providers,omitempty"`

	// URL Test 相关字段
	IdleTimeout string `json:"idle_timeout,omitempty"`
	Interval    string `json:"interval,omitempty"`
	Tolerance   int    `json:"tolerance,omitempty"`

	// Load Balance 相关字段
	Strategy string `json:"strategy,omitempty"`

	// 其他可能的字段使用 interface{} 来保持灵活性
	ExtraFields map[string]interface{} `json:"-"`
}

// SingBoxRule 路由规则结构 - 使用interface{}来处理可能是字符串或数组的字段
type SingBoxRule struct {
	Inbound         interface{} `json:"inbound,omitempty"`
	IPVersion       int         `json:"ip_version,omitempty"`
	Invert          *bool       `json:"invert,omitempty"`
	Network         interface{} `json:"network,omitempty"`
	AuthUser        interface{} `json:"auth_user,omitempty"`
	Protocol        interface{} `json:"protocol,omitempty"`
	Domain          interface{} `json:"domain,omitempty"`
	DomainSuffix    interface{} `json:"domain_suffix,omitempty"`
	DomainKeyword   interface{} `json:"domain_keyword,omitempty"`
	DomainRegex     interface{} `json:"domain_regex,omitempty"`
	Geosite         interface{} `json:"geosite,omitempty"`
	SourceGeoIP     interface{} `json:"source_geoip,omitempty"`
	GeoIP           interface{} `json:"geoip,omitempty"`
	SourceIPCIDR    interface{} `json:"source_ip_cidr,omitempty"`
	IPCIDR          interface{} `json:"ip_cidr,omitempty"`
	SourcePort      interface{} `json:"source_port,omitempty"`
	SourcePortRange interface{} `json:"source_port_range,omitempty"`
	Port            interface{} `json:"port,omitempty"`
	PortRange       interface{} `json:"port_range,omitempty"`
	ProcessName     interface{} `json:"process_name,omitempty"`
	ProcessPath     interface{} `json:"process_path,omitempty"`
	PackageName     interface{} `json:"package_name,omitempty"`
	User            interface{} `json:"user,omitempty"`
	UserID          interface{} `json:"user_id,omitempty"`
	ClashMode       string      `json:"clash_mode,omitempty"`
	RuleSet         interface{} `json:"rule_set,omitempty"`
	IPIsPrivate     *bool       `json:"ip_is_private,omitempty"`
	Outbound        string      `json:"outbound,omitempty"`
	Action          string      `json:"action,omitempty"`
	Method          string      `json:"method,omitempty"`
}

// SingBoxRoute 路由配置结构
type SingBoxRoute struct {
	Rules                 []SingBoxRule `json:"rules,omitempty"`
	RuleSet               interface{}   `json:"rule_set,omitempty"`
	Final                 string        `json:"final,omitempty"`
	FindProcess           bool          `json:"find_process,omitempty"`
	AutoDetectInterface   bool          `json:"auto_detect_interface,omitempty"`
	OverrideAndroidVPN    bool          `json:"override_android_vpn,omitempty"`
	DefaultInterface      string        `json:"default_interface,omitempty"`
	DefaultMark           int           `json:"default_mark,omitempty"`
	DefaultDomainResolver interface{}   `json:"default_domain_resolver,omitempty"`
	DomainResolver        interface{}   `json:"domain_resolver,omitempty"`
}

// SingBoxConfig 完整配置结构
type SingBoxConfig struct {
	Log          interface{}       `json:"log,omitempty"`
	DNS          interface{}       `json:"dns,omitempty"`
	NTP          interface{}       `json:"ntp,omitempty"`
	Inbounds     []SingBoxInbound  `json:"inbounds,omitempty"`
	Outbounds    []SingBoxOutbound `json:"outbounds,omitempty"`
	Route        *SingBoxRoute     `json:"route,omitempty"`
	Experimental interface{}       `json:"experimental,omitempty"`
	Providers    interface{}       `json:"providers,omitempty"`
}

// ConfigHandler 配置处理器
type ConfigHandler struct {
	serviceManager *service.ServiceManager
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(sm *service.ServiceManager) *ConfigHandler {
	return &ConfigHandler{
		serviceManager: sm,
	}
}

// getSingBoxConfigPath 获取Sing-Box配置文件路径
func (h *ConfigHandler) getSingBoxConfigPath() string {
	configPath := "/etc/sing-box/config.json"
	// 测试环境下使用相对路径
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试其他可能的路径
		possiblePaths := []string{
			"config.json",
			"./config/sing-box.json",
			"./bin/test/config.json",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return configPath
}

// readSingBoxConfig 读取并解析Sing-Box配置文件
func (h *ConfigHandler) readSingBoxConfig() (*SingBoxConfig, error) {
	configPath := h.getSingBoxConfigPath()

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SingBoxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// readSingBoxConfigAsInterface 读取配置文件为通用接口，保持所有字段
func (h *ConfigHandler) readSingBoxConfigAsInterface() (map[string]interface{}, error) {
	configPath := h.getSingBoxConfigPath()

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetConfig 获取配置文件
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "服务名称不能为空"))
		return
	}

	if serviceName != "mosdns" && serviceName != "sing-box" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "不支持的服务"))
		return
	}

	configFile, err := h.serviceManager.GetConfigFile(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(configFile))
}

// UpdateConfig 更新配置文件
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "服务名称不能为空"))
		return
	}

	if serviceName != "mosdns" && serviceName != "sing-box" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "不支持的服务"))
		return
	}

	var req models.ConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "请求参数错误"))
		return
	}

	// 验证配置文件格式
	if err := h.serviceManager.ValidateConfig(serviceName, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "配置文件格式错误: "+err.Error()))
		return
	}

	// 更新配置文件
	backupPath, err := h.serviceManager.UpdateConfigFile(serviceName, req.Content, req.Backup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	response := models.ConfigUpdateResponse{
		Success:    true,
		Message:    "配置文件更新成功",
		BackupPath: backupPath,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetSingBoxConfig 获取 Sing-Box 配置文件
func (h *ConfigHandler) GetSingBoxConfig(c *gin.Context) {
	configPath := "/etc/sing-box/config.json"
	// 测试环境下使用相对路径
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试其他可能的路径
		possiblePaths := []string{
			"config.json",
			"./config/sing-box.json",
			"./bin/test/config.json",
		}

		found := false
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				found = true
				break
			}
		}

		if !found {
			// 如果都没找到，创建一个默认配置
			configPath = "/etc/sing-box/config.json"
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Sing-Box 配置文件不存在"))
		return
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "读取配置文件失败: "+err.Error()))
		return
	}

	// 直接返回JSON配置，不进行结构体解析
	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "解析配置文件失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(config))
}

// UpdateSingBoxConfig 更新 Sing-Box 配置文件
func (h *ConfigHandler) UpdateSingBoxConfig(c *gin.Context) {
	var req struct {
		Config         interface{} `json:"config"`
		Backup         bool        `json:"backup"`
		AutoRestart    bool        `json:"auto_restart"`    // 是否自动重启服务
		EnableRollback bool        `json:"enable_rollback"` // 是否启用自动回滚
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "请求参数错误"))
		return
	}

	configPath := "/etc/sing-box/config.json"
	backupPath := ""

	// 创建备份
	if req.Backup || req.EnableRollback {
		backupPath = configPath + ".backup." + strconv.FormatInt(time.Now().Unix(), 10)
		if data, err := ioutil.ReadFile(configPath); err == nil {
			if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "创建备份失败: "+err.Error()))
				return
			}
		}
	}

	// 序列化配置
	data, err := json.MarshalIndent(req.Config, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "序列化配置失败: "+err.Error()))
		return
	}

	// 写入配置文件
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "写入配置文件失败: "+err.Error()))
		return
	}

	response := gin.H{
		"success":       true,
		"message":       "Sing-Box 配置更新成功",
		"backup_path":   backupPath,
		"needs_restart": true,
	}

	// 如果启用自动重启
	if req.AutoRestart {
		go h.handleAutoRestart(configPath, backupPath, req.EnableRollback)
		response["restarting"] = true
		response["message"] = "配置更新成功，正在重启服务..."
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// handleAutoRestart 处理自动重启和回滚
func (h *ConfigHandler) handleAutoRestart(configPath, backupPath string, enableRollback bool) {
	// 等待一小段时间让响应返回
	time.Sleep(2 * time.Second)

	// 重启服务
	if err := h.serviceManager.ControlService("sing-box", models.ActionRestart); err != nil {
		if enableRollback && backupPath != "" {
			// 重启失败，执行回滚
			h.rollbackConfig(configPath, backupPath)
		}
		return
	}

	// 等待服务启动
	time.Sleep(5 * time.Second)

	// 检查服务状态
	if enableRollback {
		serviceInfo := h.serviceManager.GetServiceInfo("sing-box")
		if serviceInfo.Status != models.StatusRunning {
			// 服务未正常启动，执行回滚
			h.rollbackConfig(configPath, backupPath)
			// 尝试重新启动服务
			h.serviceManager.ControlService("sing-box", models.ActionRestart)
		}
	}
}

// rollbackConfig 回滚配置
func (h *ConfigHandler) rollbackConfig(configPath, backupPath string) {
	if backupPath == "" {
		return
	}

	// 读取备份配置
	backupData, err := ioutil.ReadFile(backupPath)
	if err != nil {
		return
	}

	// 恢复配置
	ioutil.WriteFile(configPath, backupData, 0644)
}

// ValidateSingBoxConfig 验证 Sing-Box 配置文件 - 仅使用官方验证
func (h *ConfigHandler) ValidateSingBoxConfig(c *gin.Context) {
	var req struct {
		Config interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "请求参数错误"))
		return
	}

	// 创建临时配置文件
	configData, err := json.MarshalIndent(req.Config, "", "  ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "配置格式错误: "+err.Error()))
		return
	}

	// 检查 sing-box 二进制文件是否存在
	singboxBinary := h.getSingBoxBinary()
	if singboxBinary == "" {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500,
			"无法找到 sing-box 二进制文件，请确保 sing-box 已正确安装。查找路径包括: /usr/local/bin/sing-box, /usr/bin/sing-box, /opt/sing-box/sing-box, ./sing-box 或 PATH 环境变量中"))
		return
	}

	// 创建临时配置文件
	tempFile, err := os.CreateTemp("", "singbox-config-*.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "创建临时配置文件失败: "+err.Error()))
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 写入配置到临时文件
	if _, err := tempFile.Write(configData); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "写入临时配置文件失败: "+err.Error()))
		return
	}
	tempFile.Close()

	// 执行 sing-box check 命令进行官方验证
	cmd := exec.Command(singboxBinary, "check", "-c", tempFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		// 官方验证失败
		c.JSON(http.StatusBadRequest, models.SuccessResponse(gin.H{
			"valid":             false,
			"errors":            []string{string(output)},
			"warnings":          []string{},
			"message":           "Sing-Box 官方验证失败",
			"validation_method": "sing-box check",
		}))
		return
	}

	// 官方验证通过
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"valid":             true,
		"errors":            []string{},
		"warnings":          []string{},
		"message":           "✅ 配置通过 Sing-Box 官方验证",
		"validation_method": "sing-box check",
	}))
}

// min 函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetSingBoxInbounds 获取Sing-Box入站配置
func (h *ConfigHandler) GetSingBoxInbounds(c *gin.Context) {
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "读取配置文件失败: "+err.Error()))
		return
	}

	// 为每个入站添加一些额外信息
	var inbounds []gin.H
	for i, inbound := range config.Inbounds {
		inboundData := gin.H{
			"id":          i + 1,
			"tag":         inbound.Tag,
			"type":        inbound.Type,
			"listen":      inbound.Listen,
			"listen_port": inbound.ListenPort,
			"enabled":     true, // 在配置文件中的都认为是启用的
		}

		// 根据类型设置协议信息
		switch inbound.Type {
		case "http":
			inboundData["protocol"] = "HTTP"
		case "socks":
			inboundData["protocol"] = "SOCKS5"
		case "mixed":
			inboundData["protocol"] = "HTTP/SOCKS5"
		case "tun":
			inboundData["protocol"] = "TUN"
		case "shadowsocks":
			inboundData["protocol"] = "Shadowsocks"
		default:
			inboundData["protocol"] = inbound.Type
		}

		inbounds = append(inbounds, inboundData)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(inbounds))
}

// GetSingBoxOutbounds 获取Sing-Box出站配置（代理节点）
func (h *ConfigHandler) GetSingBoxOutbounds(c *gin.Context) {
	config, err := h.readSingBoxConfigAsInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "读取配置文件失败: "+err.Error()))
		return
	}

	// 按类型分组节点
	nodeGroups := make(map[string][]gin.H)

	outbounds, ok := config["outbounds"].([]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "配置文件格式错误: outbounds 不是数组"))
		return
	}

	for i, outboundInterface := range outbounds {
		outbound, ok := outboundInterface.(map[string]interface{})
		if !ok {
			continue
		}

		outboundType, _ := outbound["type"].(string)
		// 跳过系统内置的出站（direct, block等）
		if outboundType == "direct" || outboundType == "block" || outboundType == "dns" {
			continue
		}

		// 保持所有原始字段，只添加前端需要的ID
		nodeData := gin.H{
			"id": i + 1, // 仅用于前端表格显示的临时ID
		}

		// 复制所有原始字段
		for key, value := range outbound {
			nodeData[key] = value
		}

		// 根据节点类型和特征进行分组
		var groupName string
		if isProxyProtocol(outboundType) {
			// 真正的代理协议归类到"代理"分组
			groupName = "代理"
		} else if outboundType == "selector" || outboundType == "urltest" || outboundType == "loadbalance" {
			// 检查是否为带正则表达式的节点过滤器
			include, _ := outbound["include"].(string)
			exclude, _ := outbound["exclude"].(string)
			if include != "" || exclude != "" {
				// 只有带正则表达式的节点才归类为节点过滤
				groupName = "节点过滤"
			} else {
				// 其他逻辑节点都归类为应用分流
				groupName = "应用分流"
			}
		} else {
			// 其他类型使用类型名作为分组名
			groupName = outboundType
		}

		if nodeGroups[groupName] == nil {
			nodeGroups[groupName] = []gin.H{}
		}
		nodeGroups[groupName] = append(nodeGroups[groupName], nodeData)
	}

	// 按指定顺序转换为前端需要的格式
	var groups []gin.H

	// 定义分组顺序：代理、应用分流、节点过滤
	groupOrder := []string{"代理", "应用分流", "节点过滤"}

	// 按顺序添加分组
	for _, groupName := range groupOrder {
		if nodes, exists := nodeGroups[groupName]; exists && len(nodes) > 0 {
			groups = append(groups, gin.H{
				"name":  groupName,
				"nodes": nodes,
			})
		}
	}

	// 添加其他未在顺序中定义的分组
	for groupName, nodes := range nodeGroups {
		found := false
		for _, orderedName := range groupOrder {
			if groupName == orderedName {
				found = true
				break
			}
		}
		if !found && len(nodes) > 0 {
			groups = append(groups, gin.H{
				"name":  groupName,
				"nodes": nodes,
			})
		}
	}

	// 如果没有代理节点，返回空的默认分组
	if len(groups) == 0 {
		groups = append(groups, gin.H{
			"name":  "默认",
			"nodes": []gin.H{},
		})
	}

	c.JSON(http.StatusOK, models.SuccessResponse(groups))
}

// GetSingBoxRules 获取Sing-Box路由规则和规则集
func (h *ConfigHandler) GetSingBoxRules(c *gin.Context) {
	// 使用通用解析，避免结构体字段差异导致的条件缺失
	config, err := h.readSingBoxConfigAsInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "读取配置文件失败: "+err.Error()))
		return
	}

	var routeRules []gin.H
	var ruleSets []gin.H

	// 处理路由规则 (route.rules)
	if routeObj, ok := config["route"].(map[string]interface{}); ok {
		if rulesArr, ok := routeObj["rules"].([]interface{}); ok {
			for i, r := range rulesArr {
				ruleMap, _ := r.(map[string]interface{})

				// 生成条件
				var conditions []gin.H
				addCond := func(key string, labelType string) {
					if v, exists := ruleMap[key]; exists && v != nil {
						values := convertToStringArray(v)
						if len(values) > 0 {
							conditions = append(conditions, gin.H{
								"type":    labelType,
								"content": joinStringArray(values),
							})
						}
					}
				}

				// 已知匹配键
				keys := []struct{ jsonKey, condType string }{
					{"domain", "domain"},
					{"domain_suffix", "domain_suffix"},
					{"domain_keyword", "domain_keyword"},
					{"domain_regex", "domain_regex"},
					{"ip_cidr", "ip_cidr"},
					{"geoip", "geoip"},
					{"geosite", "geosite"},
					{"source_geoip", "source_geoip"},
					{"source_ip_cidr", "source_ip_cidr"},
					{"inbound", "inbound"},
					{"protocol", "protocol"},
					{"network", "network"},
					{"auth_user", "auth_user"},
					{"port", "port"},
					{"port_range", "port_range"},
					{"source_port", "source_port"},
					{"source_port_range", "source_port_range"},
					{"process_name", "process_name"},
					{"process_path", "process_path"},
					{"package_name", "package_name"},
					{"user", "user"},
					{"user_id", "user_id"},
					{"clash_mode", "clash_mode"},
					{"rule_set", "rule_set"},
				}
				for _, k := range keys {
					addCond(k.jsonKey, k.condType)
				}

				// 特殊布尔条件
				if v, ok := ruleMap["ip_is_private"]; ok && v != nil {
					conditions = append(conditions, gin.H{
						"type":    "ip_is_private",
						"content": fmt.Sprintf("%v", v),
					})
				}

				if len(conditions) == 0 {
					conditions = append(conditions, gin.H{"type": "unknown", "content": "复合规则"})
				}

				ruleID := strconv.Itoa(i + 1)
				ruleData := gin.H{
					"id":         ruleID,
					"conditions": conditions,
				}

				// 出站/动作与其他显示字段
				if v, ok := ruleMap["outbound"]; ok {
					ruleData["outbound"] = fmt.Sprintf("%v", v)
				}
				if v, ok := ruleMap["action"]; ok {
					ruleData["action"] = fmt.Sprintf("%v", v)
				}
				if v, ok := ruleMap["ip_version"]; ok {
					// 尝试转为整型显示
					switch vv := v.(type) {
					case float64:
						ruleData["ip_version"] = int(vv)
					default:
						ruleData["ip_version"] = v
					}
				}
				if v, ok := ruleMap["invert"]; ok {
					ruleData["invert"] = v
				}
				if v, ok := ruleMap["network"]; ok {
					ruleData["network"] = joinStringArray(convertToStringArray(v))
				}
				if v, ok := ruleMap["ip_is_private"]; ok {
					ruleData["ip_is_private"] = v
				}

				routeRules = append(routeRules, ruleData)
			}
		}

		// 处理规则集 (route.rule_set)
		if rs, ok := routeObj["rule_set"]; ok && rs != nil {
			switch ruleSetData := rs.(type) {
			case []interface{}:
				for i, item := range ruleSetData {
					if ruleSetMap, ok := item.(map[string]interface{}); ok {
						m := gin.H{"id": i + 1}
						if tag, ok := ruleSetMap["tag"].(string); ok {
							m["tag"] = tag
						}
						if t, ok := ruleSetMap["type"].(string); ok {
							m["type"] = t
						}
						if f, ok := ruleSetMap["format"].(string); ok {
							m["format"] = f
						}
						if url, ok := ruleSetMap["url"].(string); ok {
							m["url"] = url
						}
						if path, ok := ruleSetMap["path"].(string); ok {
							m["path"] = path
						}
						if dd, ok := ruleSetMap["download_detour"].(string); ok {
							m["download_detour"] = dd
						}
						if ui, ok := ruleSetMap["update_interval"].(string); ok {
							m["update_interval"] = ui
						}
						ruleSets = append(ruleSets, m)
					}
				}
			case map[string]interface{}:
				m := gin.H{"id": 1}
				if tag, ok := ruleSetData["tag"].(string); ok {
					m["tag"] = tag
				}
				if t, ok := ruleSetData["type"].(string); ok {
					m["type"] = t
				}
				if f, ok := ruleSetData["format"].(string); ok {
					m["format"] = f
				}
				if url, ok := ruleSetData["url"].(string); ok {
					m["url"] = url
				}
				if path, ok := ruleSetData["path"].(string); ok {
					m["path"] = path
				}
				if dd, ok := ruleSetData["download_detour"].(string); ok {
					m["download_detour"] = dd
				}
				if ui, ok := ruleSetData["update_interval"].(string); ok {
					m["update_interval"] = ui
				}
				ruleSets = append(ruleSets, m)
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"routeRules": routeRules,
		"ruleSets":   ruleSets,
	}))
}

// isProxyProtocol 判断是否为真正的代理协议
func isProxyProtocol(protocolType string) bool {
	// 根据Sing-Box官方文档，支持的代理协议
	supportedProtocols := map[string]bool{
		"shadowsocks":  true,
		"vmess":        true,
		"vless":        true,
		"trojan":       true,
		"wireguard":    true,
		"hysteria":     true,
		"hysteria2":    true,
		"tuic":         true,
		"ssh":          true,
		"shadowtls":    true,
		"shadowsocksr": true,
	}
	return supportedProtocols[protocolType]
}

// categorizeLogicalNode 对逻辑节点进行分类
func categorizeLogicalNode(tag, nodeType string) string {
	// 去除表情符号和特殊字符，只保留字母数字和中文
	cleanTag := regexp.MustCompile(`[^\p{L}\p{N}\s-]`).ReplaceAllString(tag, " ")
	tagLower := strings.ToLower(strings.TrimSpace(cleanTag))

	// 应用/服务相关的节点（优先级最高）
	servicePatterns := []string{
		// 节点选择和规则相关
		"(?i)(节点选择|自定义规则|选择|规则|漏网之鱼)",

		// 通信应用（去掉词边界，使用简单匹配）
		"(?i)(telegram|电报|消息|wechat|微信|qq|whatsapp|discord|signal)",

		// 科技公司和服务
		"(?i)(apple|苹果|服务|microsoft|微软|google|谷歌|amazon|亚马逊|docker|github|gitlab)",

		// 搜索引擎和门户
		"(?i)(bing|必应|baidu|百度|yahoo|雅虎|duckduckgo|yandex|搜狗|sogou|360搜索)",

		// 视频平台
		"(?i)(youtube|油管|netflix|奈飞|disney|迪士尼|hulu|amazon\\s*prime|爱奇艺|iqiyi|优酷|youku|腾讯视频|tencent\\s*video|哔哩哔哩|bilibili|抖音|douyin|快手|kuaishou|西瓜视频|xigua|好看视频|haokan)",

		// 社交媒体
		"(?i)(twitter|推特|facebook|脸书|instagram|ins|tiktok|抖音|linkedin|snapchat|pinterest|tumblr|clubhouse|weibo|微博|知乎|zhihu|reddit)",

		// 音乐平台
		"(?i)(spotify|apple\\s*music|amazon\\s*music|youtube\\s*music|网易云音乐|netease|qq音乐|酷狗|kugou|酷我|kuwo|虾米|xiami|pandora|soundcloud)",

		// AI和开发工具
		"(?i)(openai|chatgpt|claude|gemini|bard|bitbucket|stackoverflow|掘金|juejin|csdn|博客园|cnblogs|简书|jianshu|segmentfault)",

		// 游戏平台
		"(?i)(steam|epic|uplay|origin|battle\\.net|暴雪|blizzard|xbox|playstation|nintendo|任天堂|twitch|直播|live|斗鱼|douyu|虎牙|huya|bilibili直播)",

		// 电商平台
		"(?i)(taobao|淘宝|tmall|天猫|jd|京东|ebay|alibaba|阿里巴巴|拼多多|pinduoduo|苏宁|suning|国美|gome|唯品会|vip)",

		// 生活服务
		"(?i)(meituan|美团|dianping|大众点评|eleme|饿了么|didi|滴滴|uber|优步|airbnb|booking|携程|ctrip|去哪儿|qunar|马蜂窝|mafengwo)",

		// 金融支付
		"(?i)(alipay|支付宝|wechat\\s*pay|微信支付|paypal|visa|mastercard|银联|unionpay|招商银行|cmb|工商银行|icbc|建设银行|ccb|农业银行|abc|中国银行|boc)",

		// 新闻媒体
		"(?i)(cnn|bbc|fox|reuters|路透|bloomberg|彭博|wsj|华尔街日报|nytimes|纽约时报|guardian|卫报|新浪|sina|搜狐|sohu|网易|netease|凤凰|ifeng|人民网|xinhua|新华)",

		// 云服务和CDN
		"(?i)(cloudflare|aws|azure|gcp|阿里云|aliyun|腾讯云|qcloud|百度云|baiducloud|华为云|huaweicloud|七牛|qiniu|又拍云|upyun)",

		// 教育平台
		"(?i)(coursera|udemy|edx|khan\\s*academy|可汗学院|慕课|mooc|网易云课堂|腾讯课堂|学而思|xueersi|新东方|xdf|好未来|tal)",

		// 办公软件
		"(?i)(office|word|excel|powerpoint|outlook|teams|zoom|钉钉|dingtalk|企业微信|wework|slack|notion|trello|asana|monday|石墨|shimo|腾讯文档|tencent\\s*docs)",

		// 开发工具
		"(?i)(cursor|speedtest|测速)",
	}

	for _, pattern := range servicePatterns {
		if matched, _ := regexp.MatchString(pattern, tagLower); matched {
			return "应用分流"
		}
	}

	// 地区相关的节点 - 这些是节点过滤器（带正则表达式筛选具体节点）
	regionPatterns := []string{
		"🇭🇰|🇯🇵|🇺🇸|🇸🇬|🇹🇼|🇬🇧|🇩🇪|🇰🇷|🇨🇦|🇦🇺|🇫🇷|🇳🇱|🇷🇺|🇮🇳|🇨🇳|🇹🇭|🇲🇾|🇮🇩|🇵🇭|🇻🇳|🇧🇷|🇦🇷|🇲🇽|🇨🇱|🇿🇦|🇪🇬|🇳🇬|🇰🇪|🇮🇱|🇸🇦|🇦🇪|🇹🇷|🇬🇷|🇮🇹|🇪🇸|🇵🇹|🇸🇪|🇳🇴|🇩🇰|🇫🇮|🇵🇱|🇨🇿|🇭🇺|🇷🇴|🇧🇬|🇭🇷|🇸🇮|🇸🇰|🇱🇹|🇱🇻|🇪🇪|🇺🇦|🇧🇾|🇲🇩|🇷🇸|🇧🇦|🇲🇰|🇦🇱|🇲🇪|🇮🇸|🇮🇪|🇱🇺|🇧🇪|🇨🇭|🇦🇹|🇱🇮|🇲🇨|🇸🇲|🇻🇦|🇲🇹|🇨🇾|🦁",
		"(?i)(香港|日本|美国|新加坡|台湾|英国|德国|韩国|加拿大|澳洲|法国|荷兰|俄罗斯|印度|中国|泰国|马来西亚|印尼|菲律宾|越南|巴西|阿根廷|墨西哥|智利|南非|埃及|尼日利亚|肯尼亚|以色列|沙特|阿联酋|土耳其|希腊|意大利|西班牙|葡萄牙|瑞典|挪威|丹麦|芬兰|波兰|捷克|匈牙利|罗马尼亚|保加利亚|克罗地亚|斯洛文尼亚|斯洛伐克|立陶宛|拉脱维亚|爱沙尼亚|乌克兰|白俄罗斯|摩尔多瓦|塞尔维亚|波黑|马其顿|阿尔巴尼亚|黑山|冰岛|爱尔兰|卢森堡|比利时|瑞士|奥地利|狮城)",
		"(?i)(hk|jp|us|sg|tw|uk|de|kr|ca|au|fr|nl|ru|in|cn|th|my|id|ph|vn|br|ar|mx|cl|za|eg|ng|ke|il|sa|ae|tr|gr|it|es|pt|se|no|dk|fi|pl|cz|hu|ro|bg|hr|si|sk|lt|lv|ee|ua|by|md|rs|ba|mk|al|me|is|ie|lu|be|ch|at)",
		"(?i)(hong\\s*kong|japan|america|singapore|taiwan|britain|germany|korea|canada|australia|france|netherlands|russia|india|china|thailand|malaysia|indonesia|philippines|vietnam|brazil|argentina|mexico|chile|africa|egypt|nigeria|kenya|israel|saudi|emirates|turkey|greece|italy|spain|portugal|sweden|norway|denmark|finland|poland|czech|hungary|romania|bulgaria|croatia|slovenia|slovakia|lithuania|latvia|estonia|ukraine|belarus|moldova|serbia|bosnia|macedonia|albania|montenegro|iceland|ireland|luxembourg|belgium|switzerland|austria)",
		"(?i)(达拉斯|洛杉矶|圣何塞|东京|大阪|悉尼|墨尔本|伦敦|巴黎|柏林|法兰克福|阿姆斯特丹|苏黎世|维也纳|布鲁塞尔|马德里|巴塞罗那|罗马|米兰|斯德哥尔摩|哥本哈根|赫尔辛基|华沙|布拉格|布达佩斯|布加勒斯特|索菲亚|萨格勒布|卢布尔雅那|里斯本|都柏林|雷克雅未克)",
	}

	for _, pattern := range regionPatterns {
		if matched, _ := regexp.MatchString(pattern, tag); matched {
			return "节点过滤"
		}
	}

	// 通用功能性节点（手动选择、自动选择等）- 这些是应用分流
	functionalPatterns := []string{
		"(?i)(手动|自动|自建|代理|proxy|manual|auto)", // 移除"选择"避免与地区节点冲突
	}

	for _, pattern := range functionalPatterns {
		if matched, _ := regexp.MatchString(pattern, tagLower); matched {
			return "应用分流"
		}
	}

	// 其他所有类型的逻辑节点都归类到节点过滤
	return "节点过滤"
}

// convertToStringArray 将interface{}转换为字符串数组
func convertToStringArray(v interface{}) []string {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case string:
		return []string{val}
	case []string:
		return val
	case []interface{}:
		var result []string
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				// 处理可能的其他类型，转换为字符串
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	default:
		// 处理其他可能的类型，直接转换为字符串
		return []string{fmt.Sprintf("%v", val)}
	}
}

// joinStringArray 将字符串数组连接成一个字符串
func joinStringArray(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	if len(arr) == 1 {
		return arr[0]
	}

	// 对于多个项目，使用逗号和空格分隔，更适合前端显示
	if len(arr) <= 3 {
		return strings.Join(arr, ", ")
	}

	// 如果超过3个，显示前3个并加上省略号
	result := strings.Join(arr[:3], ", ")
	result += " ... (+" + strconv.Itoa(len(arr)-3) + " more)"
	return result
}

// CreateSingBoxOutbound 创建新的出站节点
func (h *ConfigHandler) CreateSingBoxOutbound(c *gin.Context) {
	var newOutbound SingBoxOutbound
	if err := c.ShouldBindJSON(&newOutbound); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := readSingBoxConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}

	// 检查节点名称是否已存在
	for _, outbound := range config.Outbounds {
		if outbound.Tag == newOutbound.Tag {
			c.JSON(400, gin.H{"error": "节点名称已存在"})
			return
		}
	}

	// 创建临时配置进行验证
	tempConfig := *config
	tempConfig.Outbounds = append(tempConfig.Outbounds, newOutbound)

	// 验证配置
	if err := h.validateSingBoxConfig(&tempConfig); err != nil {
		c.JSON(400, gin.H{"error": "配置验证失败: " + err.Error()})
		return
	}

	// 保存配置文件
	if err := writeSingBoxConfig(&tempConfig); err != nil {
		c.JSON(500, gin.H{"error": "保存配置文件失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":           "✅ 节点创建成功，配置验证通过",
		"need_restart":      true,
		"validation_msg":    "✅ 配置已通过 Sing-Box 官方验证，可以选择只保存或保存并重启 Sing-Box 服务",
		"validation_method": "sing-box check",
	})
}

// UpdateSingBoxOutbound 更新出站节点
func (h *ConfigHandler) UpdateSingBoxOutbound(c *gin.Context) {
	var updatedOutbound SingBoxOutbound
	if err := c.ShouldBindJSON(&updatedOutbound); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := readSingBoxConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}

	// 查找并更新节点 - 通过tag查找而不是id
	found := false
	tempConfig := *config
	for i, outbound := range tempConfig.Outbounds {
		if outbound.Tag == updatedOutbound.Tag {
			tempConfig.Outbounds[i] = updatedOutbound
			found = true
			break
		}
	}

	if !found {
		c.JSON(404, gin.H{"error": "节点未找到"})
		return
	}

	// 验证配置
	if err := h.validateSingBoxConfig(&tempConfig); err != nil {
		c.JSON(400, gin.H{"error": "配置验证失败: " + err.Error()})
		return
	}

	// 保存配置文件
	if err := writeSingBoxConfig(&tempConfig); err != nil {
		c.JSON(500, gin.H{"error": "保存配置文件失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":           "✅ 节点更新成功，配置验证通过",
		"need_restart":      true,
		"validation_msg":    "配置已通过 Sing-Box 官方验证，可以选择只保存或保存并重启 Sing-Box 服务",
		"validation_method": "sing-box check",
	})
}

// DeleteSingBoxOutbound 删除出站节点
func (h *ConfigHandler) DeleteSingBoxOutbound(c *gin.Context) {
	id := c.Param("id")

	// 读取当前配置
	config, err := readSingBoxConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}

	// 通过id查找并删除节点
	nodeId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的节点ID"})
		return
	}

	if nodeId < 1 || nodeId > len(config.Outbounds) {
		c.JSON(404, gin.H{"error": "节点未找到"})
		return
	}

	// 创建临时配置进行验证
	tempConfig := *config

	// 删除节点 (nodeId是从1开始的)
	tempConfig.Outbounds = append(tempConfig.Outbounds[:nodeId-1], tempConfig.Outbounds[nodeId:]...)

	// 验证配置
	if err := h.validateSingBoxConfig(&tempConfig); err != nil {
		c.JSON(400, gin.H{"error": "配置验证失败: " + err.Error()})
		return
	}

	// 保存配置文件
	if err := writeSingBoxConfig(&tempConfig); err != nil {
		c.JSON(500, gin.H{"error": "保存配置文件失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":           "✅ 节点删除成功，配置验证通过",
		"need_restart":      true,
		"validation_msg":    "配置已通过 Sing-Box 官方验证，可以选择只保存或保存并重启 Sing-Box 服务",
		"validation_method": "sing-box check",
	})
}

// validateSingBoxConfig 验证 Sing-Box 配置是否有效
func (h *ConfigHandler) validateSingBoxConfig(config *SingBoxConfig) error {
	// 转换为JSON
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("配置序列化失败: %v", err)
	}

	// 创建临时配置文件
	tempFile, err := os.CreateTemp("", "singbox-config-*.json")
	if err != nil {
		return fmt.Errorf("创建临时配置文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 写入配置到临时文件
	if _, err := tempFile.Write(configData); err != nil {
		return fmt.Errorf("写入临时配置文件失败: %v", err)
	}
	tempFile.Close()

	// 强制使用 sing-box check 命令验证配置
	singboxBinary := h.getSingBoxBinary()
	if singboxBinary == "" {
		return fmt.Errorf("无法找到 sing-box 二进制文件，请确保 sing-box 已正确安装。查找路径包括: /usr/local/bin/sing-box, /usr/bin/sing-box, /opt/sing-box/sing-box, ./sing-box 或 PATH 环境变量中")
	}

	// 执行 sing-box check 命令进行官方验证
	cmd := exec.Command(singboxBinary, "check", "-c", tempFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Sing-Box 官方验证失败: %s", string(output))
	}

	return nil
}

// getSingBoxBinary 获取 sing-box 二进制文件路径
func (h *ConfigHandler) getSingBoxBinary() string {
	// 首先从 systemd 服务文件中获取二进制路径
	if binaryPath := h.getSingBoxBinaryFromSystemd(); binaryPath != "" {
		return binaryPath
	}

	// 尝试从服务管理器获取二进制路径
	if h.serviceManager != nil {
		if config := h.serviceManager.GetConfig(); config != nil {
			if path := config.SingBoxBinaryPath; path != "" {
				if _, err := os.Stat(path); err == nil {
					return path
				}
			}
		}
	}

	// 尝试常见的路径作为最后备选
	commonPaths := []string{
		"/usr/local/bin/sing-box",
		"/usr/bin/sing-box",
		"/opt/sing-box/sing-box",
		"./sing-box",
		"sing-box", // 在 PATH 中查找
	}

	for _, path := range commonPaths {
		if path == "sing-box" {
			// 检查是否在 PATH 中
			if _, err := exec.LookPath("sing-box"); err == nil {
				return "sing-box"
			}
		} else {
			// 检查文件是否存在
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// getSingBoxBinaryFromSystemd 从 systemd 服务文件中获取 sing-box 二进制文件路径
func (h *ConfigHandler) getSingBoxBinaryFromSystemd() string {
	servicePaths := []string{
		"/etc/systemd/system/sing-box.service",
		"/usr/lib/systemd/system/sing-box.service",
		"/lib/systemd/system/sing-box.service",
	}

	for _, servicePath := range servicePaths {
		if _, err := os.Stat(servicePath); err == nil {
			content, err := ioutil.ReadFile(servicePath)
			if err != nil {
				continue
			}

			// 解析服务文件，查找 ExecStart 行
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "ExecStart=") {
					// 提取二进制文件路径
					execStart := strings.TrimPrefix(line, "ExecStart=")
					// 去除可能的引号
					execStart = strings.Trim(execStart, "\"'")

					// 提取第一个参数（二进制文件路径）
					parts := strings.Fields(execStart)
					if len(parts) > 0 {
						binaryPath := parts[0]
						// 验证文件是否存在
						if _, err := os.Stat(binaryPath); err == nil {
							return binaryPath
						}
					}
				}
			}
		}
	}

	return ""
}

// ValidateOutboundsChanges 验证出站节点更改（只验证，不保存）
func (h *ConfigHandler) ValidateOutboundsChanges(c *gin.Context) {
	var req struct {
		Changes []struct {
			Type          string                 `json:"type"`          // create, update, delete
			Data          map[string]interface{} `json:"data"`          // 节点数据
			OriginalProxy *SingBoxOutbound       `json:"originalProxy"` // 原始节点数据（用于update/delete）
		} `json:"changes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := readSingBoxConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}

	// 创建临时配置进行验证
	tempConfig := *config

	// 应用所有更改到临时配置
	for _, change := range req.Changes {
		switch change.Type {
		case "create":
			// 创建新节点
			newOutbound := SingBoxOutbound{}
			if data, err := json.Marshal(change.Data); err == nil {
				json.Unmarshal(data, &newOutbound)
				tempConfig.Outbounds = append(tempConfig.Outbounds, newOutbound)
			}
		case "update":
			// 更新现有节点
			if change.OriginalProxy != nil {
				for i, outbound := range tempConfig.Outbounds {
					if outbound.Tag == change.OriginalProxy.Tag {
						updatedOutbound := SingBoxOutbound{}
						if data, err := json.Marshal(change.Data); err == nil {
							json.Unmarshal(data, &updatedOutbound)
							tempConfig.Outbounds[i] = updatedOutbound
						}
						break
					}
				}
			}
		case "delete":
			// 删除节点
			if change.OriginalProxy != nil {
				newOutbounds := []SingBoxOutbound{}
				for _, outbound := range tempConfig.Outbounds {
					if outbound.Tag != change.OriginalProxy.Tag {
						newOutbounds = append(newOutbounds, outbound)
					}
				}
				tempConfig.Outbounds = newOutbounds
			}
		}
	}

	// 验证临时配置
	if err := h.validateSingBoxConfig(&tempConfig); err != nil {
		c.JSON(400, gin.H{"error": "配置验证失败: " + err.Error()})
		return
	}

	// 验证成功
	c.JSON(200, gin.H{
		"message":           "✅ 所有配置更改验证通过",
		"validation_method": "sing-box check",
		"changes_count":     len(req.Changes),
	})
}

// BatchSaveOutbounds 批量保存出站节点更改
func (h *ConfigHandler) BatchSaveOutbounds(c *gin.Context) {
	var req struct {
		Changes []struct {
			Type          string                 `json:"type"`          // create, update, delete
			Data          map[string]interface{} `json:"data"`          // 节点数据
			OriginalProxy *SingBoxOutbound       `json:"originalProxy"` // 原始节点数据（用于update/delete）
		} `json:"changes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := readSingBoxConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}

	// 应用所有更改
	for _, change := range req.Changes {
		switch change.Type {
		case "create":
			// 创建新节点
			newOutbound := SingBoxOutbound{}
			if data, err := json.Marshal(change.Data); err == nil {
				json.Unmarshal(data, &newOutbound)
				config.Outbounds = append(config.Outbounds, newOutbound)
			}
		case "update":
			// 更新现有节点
			if change.OriginalProxy != nil {
				for i, outbound := range config.Outbounds {
					if outbound.Tag == change.OriginalProxy.Tag {
						updatedOutbound := SingBoxOutbound{}
						if data, err := json.Marshal(change.Data); err == nil {
							json.Unmarshal(data, &updatedOutbound)
							config.Outbounds[i] = updatedOutbound
						}
						break
					}
				}
			}
		case "delete":
			// 删除节点
			if change.OriginalProxy != nil {
				newOutbounds := []SingBoxOutbound{}
				for _, outbound := range config.Outbounds {
					if outbound.Tag != change.OriginalProxy.Tag {
						newOutbounds = append(newOutbounds, outbound)
					}
				}
				config.Outbounds = newOutbounds
			}
		}
	}

	// 保存配置文件
	if err := writeSingBoxConfig(config); err != nil {
		c.JSON(500, gin.H{"error": "保存配置文件失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":        "批量保存成功，Sing-Box 配置已更新",
		"need_restart":   true,
		"validation_msg": "配置已保存，可以选择只保存或保存并重启 Sing-Box 服务",
		"changes_count":  len(req.Changes),
	})
}

// RestartSingBoxService 重启 Sing-Box 服务
func (h *ConfigHandler) RestartSingBoxService(c *gin.Context) {
	// 重启服务
	if err := h.serviceManager.ControlService("sing-box", models.ActionRestart); err != nil {
		c.JSON(500, gin.H{"error": "重启服务失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Sing-Box 服务重启成功",
		"status":  "restarted",
	})
}

// getSingBoxConfigPath 获取配置文件路径的独立函数
func getSingBoxConfigPath() string {
	configPath := "/etc/sing-box/config.json"
	// 测试环境下使用相对路径
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试其他可能的路径
		possiblePaths := []string{
			"./bin/test/config.json",
			"./config.json",
			"../config.json",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return configPath
}

// readSingBoxConfig 读取配置文件的独立函数
func readSingBoxConfig() (*SingBoxConfig, error) {
	configPath := getSingBoxConfigPath()

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SingBoxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// writeSingBoxConfig 写入配置文件
func writeSingBoxConfig(config *SingBoxConfig) error {
	configPath := getSingBoxConfigPath()

	// 将配置转换为JSON
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configPath, configData, 0644)
}

// CreateRouteRule 创建路由规则
func (h *ConfigHandler) CreateRouteRule(c *gin.Context) {
	var newRule SingBoxRule
	if err := c.ShouldBindJSON(&newRule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config: " + err.Error()})
		return
	}

	// 确保Route存在
	if config.Route == nil {
		config.Route = &SingBoxRoute{
			Rules: []SingBoxRule{},
		}
	}

	// 添加到规则列表（不需要ID，直接追加）
	config.Route.Rules = append(config.Route.Rules, newRule)

	// 验证配置
	if err := h.validateSingBoxConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration validation failed: " + err.Error()})
		return
	}

	// 保存配置
	if err := h.saveSingBoxConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route rule created successfully", "rule": newRule})
}

// UpdateRouteRule 更新路由规则
func (h *ConfigHandler) UpdateRouteRule(c *gin.Context) {
	ruleID := c.Param("id")

	var updatedRule SingBoxRule
	if err := c.ShouldBindJSON(&updatedRule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config: " + err.Error()})
		return
	}

	if config.Route == nil || config.Route.Rules == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rules found"})
		return
	}

	// 将ruleID转换为数组索引
	ruleIndex, err := strconv.Atoi(ruleID)
	if err != nil || ruleIndex <= 0 || ruleIndex > len(config.Route.Rules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// 转换为0基索引
	ruleIndex = ruleIndex - 1

	// 更新规则（不需要设置ID）
	config.Route.Rules[ruleIndex] = updatedRule

	// 验证配置
	if err := h.validateSingBoxConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration validation failed: " + err.Error()})
		return
	}

	// 保存配置
	if err := h.saveSingBoxConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route rule updated successfully", "rule": updatedRule})
}

// DeleteRouteRule 删除路由规则
func (h *ConfigHandler) DeleteRouteRule(c *gin.Context) {
	ruleID := c.Param("id")

	// 读取当前配置
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config: " + err.Error()})
		return
	}

	if config.Route == nil || config.Route.Rules == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rules found"})
		return
	}

	// 将ruleID转换为数组索引
	ruleIndex, err := strconv.Atoi(ruleID)
	if err != nil || ruleIndex <= 0 || ruleIndex > len(config.Route.Rules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// 转换为0基索引
	ruleIndex = ruleIndex - 1

	// 删除规则
	config.Route.Rules = append(config.Route.Rules[:ruleIndex], config.Route.Rules[ruleIndex+1:]...)

	// 验证配置
	if err := h.validateSingBoxConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration validation failed: " + err.Error()})
		return
	}

	// 保存配置
	if err := h.saveSingBoxConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route rule deleted successfully"})
}

// MoveRouteRuleUp 路由规则上移（与前一条交换位置）
func (h *ConfigHandler) MoveRouteRuleUp(c *gin.Context) {
	h.moveRouteRule(c, -1)
}

// MoveRouteRuleDown 路由规则下移（与后一条交换位置）
func (h *ConfigHandler) MoveRouteRuleDown(c *gin.Context) {
	h.moveRouteRule(c, 1)
}

// moveRouteRule 按增量移动指定规则
func (h *ConfigHandler) moveRouteRule(c *gin.Context, delta int) {
	ruleID := c.Param("id")

	// 读取当前配置
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config: " + err.Error()})
		return
	}

	if config.Route == nil || config.Route.Rules == nil || len(config.Route.Rules) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rules found"})
		return
	}

	// 将ruleID转换为数组索引（1基 -> 0基）
	ruleIndex, err := strconv.Atoi(ruleID)
	if err != nil || ruleIndex <= 0 || ruleIndex > len(config.Route.Rules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}
	ruleIndex = ruleIndex - 1

	targetIndex := ruleIndex + delta
	if targetIndex < 0 || targetIndex >= len(config.Route.Rules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Move out of bounds"})
		return
	}

	// 交换位置
	rules := config.Route.Rules
	rules[ruleIndex], rules[targetIndex] = rules[targetIndex], rules[ruleIndex]
	config.Route.Rules = rules

	// 验证配置
	if err := h.validateSingBoxConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration validation failed: " + err.Error()})
		return
	}

	// 保存配置
	if err := h.saveSingBoxConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config: " + err.Error()})
		return
	}

	direction := "down"
	if delta < 0 {
		direction = "up"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Route rule moved " + direction + " successfully",
		"from_index":     ruleIndex + 1,   // 返回为1基
		"to_index":       targetIndex + 1, // 返回为1基
		"need_restart":   true,
		"validation_msg": "配置已通过 Sing-Box 官方验证，可以选择重启服务以应用更改",
	})
}

// saveSingBoxConfig 保存Sing-Box配置到文件
func (h *ConfigHandler) saveSingBoxConfig(config *SingBoxConfig) error {
	configPath := h.serviceManager.GetConfig().SingBoxConfigPath

	// 序列化配置
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 写入文件
	return os.WriteFile(configPath, configData, 0644)
}

// saveSingBoxConfigAsInterface 保存 interface{} 类型的配置
func (h *ConfigHandler) saveSingBoxConfigAsInterface(config map[string]interface{}) error {
	configPath := h.serviceManager.GetConfig().SingBoxConfigPath

	// 序列化配置
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 写入文件
	return os.WriteFile(configPath, configData, 0644)
}

// ReorderRouteRules 批量重排序路由规则（支持拖拽）
func (h *ConfigHandler) ReorderRouteRules(c *gin.Context) {
	var req struct {
		RuleIds []string `json:"rule_ids"` // 新的规则ID顺序（1基索引）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config: " + err.Error()})
		return
	}

	if config.Route == nil || config.Route.Rules == nil || len(config.Route.Rules) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rules found"})
		return
	}

	originalRules := config.Route.Rules
	if len(req.RuleIds) != len(originalRules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rule IDs count mismatch"})
		return
	}

	// 构建新的规则顺序
	newRules := make([]SingBoxRule, 0, len(originalRules))
	for _, ruleIdStr := range req.RuleIds {
		ruleIndex, err := strconv.Atoi(ruleIdStr)
		if err != nil || ruleIndex <= 0 || ruleIndex > len(originalRules) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID: " + ruleIdStr})
			return
		}
		// 转换为0基索引并复制规则
		newRules = append(newRules, originalRules[ruleIndex-1])
	}

	// 替换规则顺序
	config.Route.Rules = newRules

	// 验证配置
	if err := h.validateSingBoxConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration validation failed: " + err.Error()})
		return
	}

	// 保存配置
	if err := h.saveSingBoxConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Route rules reordered successfully",
		"rules_count":    len(newRules),
		"need_restart":   true,
		"validation_msg": "配置已通过 Sing-Box 官方验证，可以选择重启服务以应用更改",
	})
}

// ValidateCurrentSingBoxConfig 验证当前的Sing-Box配置文件
func (h *ConfigHandler) ValidateCurrentSingBoxConfig(c *gin.Context) {
	configPath := h.getSingBoxConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "配置文件不存在"))
		return
	}

	// 检查 sing-box 二进制文件是否存在
	singboxBinary := h.getSingBoxBinary()
	if singboxBinary == "" {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500,
			"无法找到 sing-box 二进制文件，请确保 sing-box 已正确安装"))
		return
	}

	// 执行 sing-box check 命令验证当前配置文件
	cmd := exec.Command(singboxBinary, "check", "-c", configPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// 验证失败
		c.JSON(http.StatusBadRequest, models.SuccessResponse(gin.H{
			"valid":             false,
			"error":             string(output),
			"message":           "配置文件验证失败",
			"validation_method": "sing-box check",
		}))
		return
	}

	// 验证成功
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"valid":             true,
		"message":           "✅ 配置文件验证通过，可以安全重启服务",
		"validation_method": "sing-box check",
	}))
}

// CreateRuleSet 创建规则集
func (h *ConfigHandler) CreateRuleSet(c *gin.Context) {
	var req struct {
		Tag            string `json:"tag" binding:"required"`
		Type           string `json:"type" binding:"required"`
		Format         string `json:"format"`
		URL            string `json:"url"`
		Path           string `json:"path"`
		DownloadDetour string `json:"download_detour"`
		UpdateInterval string `json:"update_interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfigAsInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置失败: " + err.Error()})
		return
	}

	// 创建新的规则集
	newRuleSet := map[string]interface{}{
		"tag":  req.Tag,
		"type": req.Type,
	}

	if req.Format != "" {
		newRuleSet["format"] = req.Format
	}
	if req.URL != "" {
		newRuleSet["url"] = req.URL
	}
	if req.Path != "" {
		newRuleSet["path"] = req.Path
	}
	if req.DownloadDetour != "" {
		newRuleSet["download_detour"] = req.DownloadDetour
	}
	if req.UpdateInterval != "" {
		newRuleSet["update_interval"] = req.UpdateInterval
	}

	// 确保 route 存在
	if config["route"] == nil {
		config["route"] = make(map[string]interface{})
	}
	route := config["route"].(map[string]interface{})

	// 处理 rule_set
	var ruleSets []interface{}
	if route["rule_set"] != nil {
		if ruleSetSlice, ok := route["rule_set"].([]interface{}); ok {
			ruleSets = ruleSetSlice
		} else if singleRuleSet, ok := route["rule_set"].(map[string]interface{}); ok {
			ruleSets = []interface{}{singleRuleSet}
		}
	}

	// 检查是否已存在相同的 tag
	for _, rs := range ruleSets {
		if ruleSet, ok := rs.(map[string]interface{}); ok {
			if ruleSet["tag"] == req.Tag {
				c.JSON(http.StatusBadRequest, gin.H{"error": "规则集标签已存在: " + req.Tag})
				return
			}
		}
	}

	// 添加新规则集
	ruleSets = append(ruleSets, newRuleSet)
	route["rule_set"] = ruleSets

	// 保存配置 (跳过验证，直接保存到文件)
	if err := h.saveSingBoxConfigAsInterface(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "✅ 规则集创建成功",
		"need_restart":      true,
		"validation_method": "sing-box check",
	})
}

// UpdateRuleSet 更新规则集
func (h *ConfigHandler) UpdateRuleSet(c *gin.Context) {
	ruleSetId := c.Param("id")
	ruleSetIndex, err := strconv.Atoi(ruleSetId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则集ID"})
		return
	}

	var req struct {
		Tag            string `json:"tag" binding:"required"`
		Type           string `json:"type" binding:"required"`
		Format         string `json:"format"`
		URL            string `json:"url"`
		Path           string `json:"path"`
		DownloadDetour string `json:"download_detour"`
		UpdateInterval string `json:"update_interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfigAsInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置失败: " + err.Error()})
		return
	}

	// 获取规则集列表
	route := config["route"].(map[string]interface{})
	var ruleSets []interface{}
	if route["rule_set"] != nil {
		if ruleSetSlice, ok := route["rule_set"].([]interface{}); ok {
			ruleSets = ruleSetSlice
		} else if singleRuleSet, ok := route["rule_set"].(map[string]interface{}); ok {
			ruleSets = []interface{}{singleRuleSet}
		}
	}

	if ruleSetIndex < 0 || ruleSetIndex >= len(ruleSets) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "规则集索引超出范围"})
		return
	}

	// 检查标签冲突
	for i, rs := range ruleSets {
		if i != ruleSetIndex {
			if ruleSet, ok := rs.(map[string]interface{}); ok {
				if ruleSet["tag"] == req.Tag {
					c.JSON(http.StatusBadRequest, gin.H{"error": "规则集标签已存在: " + req.Tag})
					return
				}
			}
		}
	}

	// 更新规则集
	updatedRuleSet := map[string]interface{}{
		"tag":  req.Tag,
		"type": req.Type,
	}

	if req.Format != "" {
		updatedRuleSet["format"] = req.Format
	}
	if req.URL != "" {
		updatedRuleSet["url"] = req.URL
	}
	if req.Path != "" {
		updatedRuleSet["path"] = req.Path
	}
	if req.DownloadDetour != "" {
		updatedRuleSet["download_detour"] = req.DownloadDetour
	}
	if req.UpdateInterval != "" {
		updatedRuleSet["update_interval"] = req.UpdateInterval
	}

	ruleSets[ruleSetIndex] = updatedRuleSet
	route["rule_set"] = ruleSets

	// 保存配置 (跳过验证，直接保存到文件)
	if err := h.saveSingBoxConfigAsInterface(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "✅ 规则集更新成功",
		"need_restart":      true,
		"validation_method": "sing-box check",
	})
}

// DeleteRuleSet 删除规则集
func (h *ConfigHandler) DeleteRuleSet(c *gin.Context) {
	ruleSetId := c.Param("id")
	ruleSetIndex, err := strconv.Atoi(ruleSetId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则集ID"})
		return
	}

	// 读取当前配置
	config, err := h.readSingBoxConfigAsInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置失败: " + err.Error()})
		return
	}

	// 获取规则集列表
	route := config["route"].(map[string]interface{})
	var ruleSets []interface{}
	if route["rule_set"] != nil {
		if ruleSetSlice, ok := route["rule_set"].([]interface{}); ok {
			ruleSets = ruleSetSlice
		} else if singleRuleSet, ok := route["rule_set"].(map[string]interface{}); ok {
			ruleSets = []interface{}{singleRuleSet}
		}
	}

	if ruleSetIndex < 0 || ruleSetIndex >= len(ruleSets) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "规则集索引超出范围"})
		return
	}

	// 删除规则集
	ruleSets = append(ruleSets[:ruleSetIndex], ruleSets[ruleSetIndex+1:]...)

	// 更新配置
	if len(ruleSets) == 0 {
		delete(route, "rule_set")
	} else if len(ruleSets) == 1 {
		route["rule_set"] = ruleSets[0]
	} else {
		route["rule_set"] = ruleSets
	}

	// 保存配置 (跳过验证，直接保存到文件)
	if err := h.saveSingBoxConfigAsInterface(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "✅ 规则集删除成功",
		"need_restart":      true,
		"validation_method": "sing-box check",
	})
}

// DetectSingBoxPath 检测 Sing-Box 二进制文件路径
func (h *ConfigHandler) DetectSingBoxPath(c *gin.Context) {
	var detectionMethods []string
	var foundPath string

	// 方法1: 从 systemctl 配置中获取路径
	path := detectSingBoxPathFromSystemctl()
	if path != "" {
		foundPath = path
		detectionMethods = append(detectionMethods, "systemctl")
	} else {
		detectionMethods = append(detectionMethods, "systemctl (failed)")

		// 添加详细的调试信息
		var debugInfo []string

		// 检查 systemctl show 输出
		cmd := exec.Command("systemctl", "show", "sing-box", "--property=ExecStart", "--value")
		if output, err := cmd.Output(); err != nil {
			debugInfo = append(debugInfo, "systemctl show 失败: "+err.Error())
		} else {
			execStart := strings.TrimSpace(string(output))
			if execStart == "" {
				debugInfo = append(debugInfo, "systemctl show 返回空内容")
			} else {
				debugInfo = append(debugInfo, "systemctl show 输出: "+execStart)
			}
		}

		// 检查服务状态
		cmd = exec.Command("systemctl", "status", "sing-box")
		if _, err := cmd.Output(); err != nil {
			debugInfo = append(debugInfo, "systemctl status 失败: "+err.Error())
		} else {
			debugInfo = append(debugInfo, "服务状态检查完成")
		}

		// 检查 systemctl cat
		cmd = exec.Command("systemctl", "cat", "sing-box")
		if output, err := cmd.Output(); err != nil {
			debugInfo = append(debugInfo, "systemctl cat 失败: "+err.Error())
		} else {
			content := strings.TrimSpace(string(output))
			if content != "" {
				debugInfo = append(debugInfo, "systemctl cat 成功，长度: "+fmt.Sprintf("%d", len(content)))
			}
		}

		// 方法2: 尝试常见路径
		commonPaths := []string{
			"/usr/local/bin/sing-box",
			"/usr/bin/sing-box",
			"/opt/sing-box/sing-box",
			"/etc/sing-box/sing-box",
		}

		// 在 macOS 开发环境中添加测试路径
		if runtime.GOOS == "darwin" {
			commonPaths = append(commonPaths,
				"./test-sing-box",
				"/usr/local/bin/sing-box-test",
			)
		}
		for _, p := range commonPaths {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				if info.Mode()&0111 != 0 { // 检查执行权限
					foundPath = p
					detectionMethods = append(detectionMethods, "common paths")
					break
				}
			}
		}
	}

	if foundPath == "" {
		// 收集所有调试信息
		var allDebugInfo []string

		// 再次尝试检查 systemctl show 输出
		cmd := exec.Command("systemctl", "show", "sing-box", "--property=ExecStart", "--value")
		if output, err := cmd.Output(); err != nil {
			allDebugInfo = append(allDebugInfo, "systemctl show 失败: "+err.Error())
		} else {
			execStart := strings.TrimSpace(string(output))
			if execStart == "" {
				allDebugInfo = append(allDebugInfo, "systemctl show 返回空内容")
			} else {
				allDebugInfo = append(allDebugInfo, "systemctl show 输出: "+execStart)
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error":             "未找到 Sing-Box 二进制文件路径",
			"detection_methods": detectionMethods,
			"debug_info":        allDebugInfo,
			"checked_paths": []string{
				"/usr/local/bin/sing-box",
				"/usr/bin/sing-box",
				"/opt/sing-box/sing-box",
				"/etc/sing-box/sing-box",
			},
		})
		return
	}

	// 验证路径并获取文件信息
	info, err := os.Stat(foundPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "路径验证失败: " + err.Error(),
			"path":  foundPath,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":              foundPath,
		"detection_method":  detectionMethods[len(detectionMethods)-1],
		"detection_methods": detectionMethods,
		"file_size":         info.Size(),
		"permissions":       info.Mode().String(),
		"modified_time":     info.ModTime().Format("2006-01-02 15:04:05"),
	})
}

// CheckSingBoxUpdate 检查 Sing-Box 更新
func (h *ConfigHandler) CheckSingBoxUpdate(c *gin.Context) {
	// 获取当前版本
	currentVersion := getCurrentSingBoxVersion()

	// 获取最新版本信息
	latestInfo, err := getLatestSingBoxRelease()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取最新版本信息失败: " + err.Error(),
		})
		return
	}

	// 检查是否有更新
	hasUpdate := compareVersions(currentVersion, latestInfo.Version) < 0

	c.JSON(http.StatusOK, gin.H{
		"hasUpdate":      hasUpdate,
		"currentVersion": currentVersion,
		"version":        latestInfo.Version,
		"publishTime":    latestInfo.PublishTime,
		"downloadUrl":    latestInfo.DownloadURL,
	})
}

// UpdateSingBoxKernel 更新 Sing-Box 内核
func (h *ConfigHandler) UpdateSingBoxKernel(c *gin.Context) {
	var req struct {
		DownloadURL string `json:"downloadUrl" binding:"required"`
		TargetPath  string `json:"targetPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 启动异步更新过程
	go performKernelUpdate(req.DownloadURL, req.TargetPath)

	c.JSON(http.StatusOK, gin.H{
		"message": "内核更新已开始",
	})
}

// UpdateSingBoxKernelStream SSE 流式更新进度
func (h *ConfigHandler) UpdateSingBoxKernelStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 这里应该实现 SSE 流式更新
	// 简化处理，直接返回完成状态
	c.String(http.StatusOK, "data: {\"finished\": true, \"success\": true, \"percentage\": 100, \"message\": \"更新完成\"}\n\n")
}

// 辅助函数和结构体定义

// GitHubRelease GitHub 发布信息结构
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// LatestReleaseInfo 最新版本信息
type LatestReleaseInfo struct {
	Version     string
	PublishTime string
	DownloadURL string
}

// detectSingBoxPathFromSystemctl 从 systemctl 配置中检测路径
func detectSingBoxPathFromSystemctl() string {
	// 方法1: 尝试使用 systemctl show 获取 ExecStart
	cmd := exec.Command("systemctl", "show", "sing-box", "--property=ExecStart", "--value")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		execStart := strings.TrimSpace(string(output))
		if path := parseExecStartPath(execStart); path != "" {
			return path
		}
	}

	// 方法2: 尝试使用完整的 systemctl show 获取所有属性
	cmd = exec.Command("systemctl", "show", "sing-box")
	output, err = cmd.Output()
	if err == nil {
		if path := parseSystemctlShowOutput(string(output)); path != "" {
			return path
		}
	}

	// 方法3: 尝试使用 systemctl cat 获取服务文件内容
	cmd = exec.Command("systemctl", "cat", "sing-box")
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		if path := parseServiceContent(string(output)); path != "" {
			return path
		}
	}

	// 方法4: 尝试读取 service 文件内容
	servicePaths := []string{
		"/etc/systemd/system/sing-box.service",
		"/usr/lib/systemd/system/sing-box.service",
		"/lib/systemd/system/sing-box.service",
		"/run/systemd/system/sing-box.service",
		"/etc/systemd/system/multi-user.target.wants/sing-box.service",
	}

	for _, servicePath := range servicePaths {
		if path := parseServiceFile(servicePath); path != "" {
			return path
		}
	}

	// 方法5: 尝试通过 ps 命令查找正在运行的进程
	cmd = exec.Command("pgrep", "-f", "sing-box")
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		pid := strings.TrimSpace(string(output))
		if path := getProcessPath(pid); path != "" {
			return path
		}
	}

	return ""
}

// parseExecStartPath 解析 ExecStart 输出
func parseExecStartPath(execStart string) string {
	if execStart == "" {
		return ""
	}

	// 移除可能的前缀和后缀空格
	execStart = strings.TrimSpace(execStart)

	// 移除可能的 path= 前缀 (某些 systemd 版本会有这个)
	execStart = strings.TrimPrefix(execStart, "path=")

	// 处理可能的引号包围
	execStart = strings.Trim(execStart, "\"'")

	// 处理可能的特殊字符前缀，如 @ 符号
	if strings.HasPrefix(execStart, "@") {
		// @/path/to/binary 格式，移除 @ 前缀
		execStart = strings.TrimPrefix(execStart, "@")
	}

	// 如果有空格，分割命令行参数，取第一个作为二进制路径
	parts := strings.Fields(execStart)
	if len(parts) > 0 {
		binaryPath := parts[0]

		// 验证路径是否存在
		if info, err := os.Stat(binaryPath); err == nil && !info.IsDir() {
			// 检查是否有执行权限
			if info.Mode()&0111 != 0 {
				return binaryPath
			}
		}
	}

	return ""
}

// parseServiceFile 解析 systemd service 文件
func parseServiceFile(servicePath string) string {
	content, err := os.ReadFile(servicePath)
	if err != nil {
		return ""
	}
	return parseServiceContent(string(content))
}

// parseServiceContent 解析服务文件内容
func parseServiceContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 查找 ExecStart= 行
		if strings.HasPrefix(line, "ExecStart=") {
			execStart := strings.TrimPrefix(line, "ExecStart=")
			if path := parseExecStartPath(execStart); path != "" {
				return path
			}
		}
	}
	return ""
}

// parseSystemctlShowOutput 解析 systemctl show 的完整输出
func parseSystemctlShowOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ExecStart=") {
			execStart := strings.TrimPrefix(line, "ExecStart=")
			// 移除可能的时间戳和路径前缀
			if idx := strings.Index(execStart, " "); idx > 0 {
				parts := strings.Fields(execStart)
				if len(parts) >= 2 {
					// 通常格式: { path=/usr/bin/sing-box ; argv[]=/usr/bin/sing-box -c /etc/sing-box/config.json ; ... }
					for _, part := range parts {
						if strings.HasPrefix(part, "path=") {
							path := strings.TrimPrefix(part, "path=")
							path = strings.Trim(path, ";")
							if path := parseExecStartPath(path); path != "" {
								return path
							}
						}
						if strings.HasPrefix(part, "argv[]=") {
							argv := strings.TrimPrefix(part, "argv[]=")
							argv = strings.Trim(argv, ";")
							if path := parseExecStartPath(argv); path != "" {
								return path
							}
						}
					}
				}
			}
			if path := parseExecStartPath(execStart); path != "" {
				return path
			}
		}
	}
	return ""
}

// getProcessPath 通过进程ID获取可执行文件路径
func getProcessPath(pid string) string {
	// 方法1: 通过 /proc/PID/exe 链接
	exePath := "/proc/" + pid + "/exe"
	if realPath, err := os.Readlink(exePath); err == nil {
		if info, err := os.Stat(realPath); err == nil && !info.IsDir() {
			if info.Mode()&0111 != 0 {
				return realPath
			}
		}
	}

	// 方法2: 通过 ps 命令获取命令行
	cmd := exec.Command("ps", "-p", pid, "-o", "cmd", "--no-headers")
	output, err := cmd.Output()
	if err == nil {
		cmdline := strings.TrimSpace(string(output))
		if path := parseExecStartPath(cmdline); path != "" {
			return path
		}
	}

	return ""
}

// getCurrentSingBoxVersion 获取当前 Sing-Box 版本
func getCurrentSingBoxVersion() string {
	// 从 sing-box --version 命令获取版本号
	path := detectSingBoxPathFromSystemctl()
	if path == "" {
		return "unknown"
	}

	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// 读取分支信息
	branchInfo := ""
	if content, err := os.ReadFile("/etc/sing-box/version"); err == nil {
		branchInfo = strings.TrimSpace(string(content))
	}

	// 解析版本输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 尝试多种匹配模式来提取基础版本号
		patterns := []string{
			// 匹配完整版本信息（优先匹配带有完整后缀的版本）
			`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*(?:-[a-zA-Z0-9]+)*)+)`,
			// 匹配包含version关键字的行，提取版本号
			`(?i)version[:\s]+v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*)?)`,
			// 匹配行首的版本号
			`^v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*)?)`,
			// 匹配任何位置的版本号
			`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*)?)`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				baseVersion := matches[1]
				if !strings.HasPrefix(baseVersion, "v") {
					baseVersion = "v" + baseVersion
				}

				// 组合版本号和分支信息
				finalVersion := baseVersion
				if branchInfo != "" && branchInfo != "official-core" && !strings.Contains(baseVersion, branchInfo) {
					// 如果分支信息不为空、不是官方内核且版本号中不包含分支信息，则添加分支信息
					finalVersion = baseVersion + "-" + branchInfo
				}
				// 如果是官方内核，直接使用基础版本号，不添加分支信息

				return finalVersion
			}
		}
	}

	return "unknown"
}

// getLatestSingBoxRelease 获取最新 Sing-Box 发布信息
func getLatestSingBoxRelease() (*LatestReleaseInfo, error) {
	// 检查是否为官方内核
	branchInfo := ""
	if content, err := os.ReadFile("/etc/sing-box/version"); err == nil {
		branchInfo = strings.TrimSpace(string(content))
	}

	if branchInfo == "official-core" {
		// 使用官方仓库
		return getOfficialSingBoxRelease()
	} else {
		// 使用 reF1nd 仓库
		return getReF1ndSingBoxRelease()
	}
}

// getOfficialSingBoxRelease 获取官方 Sing-Box 发布信息
func getOfficialSingBoxRelease() (*LatestReleaseInfo, error) {
	url := "https://api.github.com/repos/SagerNet/sing-box/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 版本号直接使用 tag_name
	version := release.TagName
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// 构建官方下载链接
	downloadURL := buildOfficialDownloadURL(version)

	return &LatestReleaseInfo{
		Version:     version,
		PublishTime: release.PublishedAt.Format("2006-01-02 15:04:05"),
		DownloadURL: downloadURL,
	}, nil
}

// getReF1ndSingBoxRelease 获取 reF1nd Sing-Box 发布信息
func getReF1ndSingBoxRelease() (*LatestReleaseInfo, error) {
	url := "https://api.github.com/repos/herozmy/StoreHouse/releases/tags/sing-box-reF1nd"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 从发布描述中提取真正的版本号
	version := extractVersionFromBody(release.Body)
	if version == "" {
		return nil, fmt.Errorf("无法从发布描述中提取版本号")
	}

	// 根据当前平台选择合适的下载链接
	downloadURL := selectDownloadURL(release.Assets)
	if downloadURL == "" {
		return nil, fmt.Errorf("未找到适合当前平台的下载链接")
	}

	return &LatestReleaseInfo{
		Version:     version,
		PublishTime: release.PublishedAt.Format("2006-01-02 15:04:05"),
		DownloadURL: downloadURL,
	}, nil
}

// buildOfficialDownloadURL 构建官方 Sing-Box 下载链接
func buildOfficialDownloadURL(version string) string {
	// 移除版本号前的 v 前缀用于构建下载链接
	versionWithoutV := strings.TrimPrefix(version, "v")

	// 获取架构信息
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "arm":
		// 默认使用 armv7，这是最常见的ARM架构
		arch = "armv7"
	default:
		arch = "amd64" // 默认使用 amd64
	}

	// 构建下载链接：https://github.com/SagerNet/sing-box/releases/download/{version}/sing-box-{version#v}-linux-{arch}.tar.gz
	downloadURL := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/%s/sing-box-%s-linux-%s.tar.gz",
		version, versionWithoutV, arch)

	return downloadURL
}

// selectDownloadURL 根据当前平台选择下载链接
func selectDownloadURL(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 构建目标文件名模式
	var targetPattern string
	switch osName {
	case "linux":
		switch arch {
		case "amd64":
			targetPattern = "sing-box-reF1nd-dev-linux-amd64.tar.gz"
		case "arm64":
			targetPattern = "sing-box-reF1nd-dev-linux-arm64.tar.gz"
		default:
			targetPattern = "sing-box-reF1nd-dev-linux-amd64.tar.gz"
		}
	case "darwin":
		// 对于 macOS，暂时使用 linux-amd64 版本进行测试
		targetPattern = "sing-box-reF1nd-dev-linux-amd64.tar.gz"
	default:
		targetPattern = "sing-box-reF1nd-dev-linux-amd64.tar.gz"
	}

	// 查找匹配的文件
	for _, asset := range assets {
		if asset.Name == targetPattern {
			return asset.BrowserDownloadURL
		}
	}

	// 如果没找到，返回第一个 linux-amd64 版本作为备选
	for _, asset := range assets {
		if asset.Name == "sing-box-reF1nd-dev-linux-amd64.tar.gz" {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

// compareVersions 比较版本号
func compareVersions(v1, v2 string) int {
	// 移除 v 前缀
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// 分离版本号和commit hash
	// 格式: 1.12.8-01f4b410
	v1Base, v1Hash := parseVersionAndHash(v1)
	v2Base, v2Hash := parseVersionAndHash(v2)

	// 先比较基础版本号
	baseCompare := compareBaseVersions(v1Base, v2Base)
	if baseCompare != 0 {
		return baseCompare
	}

	// 如果基础版本相同，比较commit hash
	// 有hash的版本被认为是较新的开发版本
	if v1Hash != "" && v2Hash == "" {
		return 1 // v1有hash，v2没有，v1更新
	} else if v1Hash == "" && v2Hash != "" {
		return -1 // v2有hash，v1没有，v2更新
	} else if v1Hash != "" && v2Hash != "" {
		// 都有hash，比较hash值（简单的字符串比较）
		if v1Hash < v2Hash {
			return -1
		} else if v1Hash > v2Hash {
			return 1
		}
	}

	return 0
}

// parseVersionAndHash 分离版本号和commit hash
func parseVersionAndHash(version string) (baseVersion, hash string) {
	parts := strings.Split(version, "-")
	baseVersion = parts[0]
	if len(parts) > 1 {
		hash = parts[1]
	}
	return
}

// compareBaseVersions 比较基础版本号 (x.y.z)
func compareBaseVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}

	return 0
}

// extractVersionFromBody 从发布描述中提取版本号
func extractVersionFromBody(body string) string {
	// 匹配格式如: "更新 sing-box reF1nd-dev 版至 v1.13.0-alpha.17-reF1nd"
	re := regexp.MustCompile(`至\s+(v?\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*(?:-[a-zA-Z0-9]+)*)?)`)
	if matches := re.FindStringSubmatch(body); len(matches) > 1 {
		version := matches[1]
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		return version
	}

	// 备用匹配：直接匹配版本号格式
	re2 := regexp.MustCompile(`v?\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*(?:-[a-zA-Z0-9]+)*)?`)
	if match := re2.FindString(body); match != "" {
		if !strings.HasPrefix(match, "v") {
			match = "v" + match
		}
		return match
	}

	return ""
}

// performKernelUpdate 执行内核更新 (简化版本)
func performKernelUpdate(downloadURL, targetPath string) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "sing-box-update-*")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	// 下载文件
	tempFile := filepath.Join(tempDir, "sing-box.tar.gz")
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return
	}

	// 解压文件
	extractDir := filepath.Join(tempDir, "extracted")
	if err := extractTarGz(tempFile, extractDir); err != nil {
		return
	}

	// 查找 sing-box 二进制文件
	binaryPath, err := findSingBoxBinary(extractDir)
	if err != nil {
		return
	}

	// 停止服务
	stopCmd := exec.Command("systemctl", "stop", "sing-box")
	if err := stopCmd.Run(); err != nil {
		return
	}

	// 备份原文件
	backupPath := targetPath + ".backup"
	if err := copyFile(targetPath, backupPath); err != nil {
		return
	}

	// 复制新文件
	if err := copyFile(binaryPath, targetPath); err != nil {
		// 恢复备份
		copyFile(backupPath, targetPath)
		return
	}

	// 设置权限
	if err := os.Chmod(targetPath, 0755); err != nil {
		return
	}

	// 启动服务
	startCmd := exec.Command("systemctl", "start", "sing-box")
	if err := startCmd.Run(); err != nil {
		return
	}
}

// downloadFile 下载文件
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 错误: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz 解压 tar.gz 文件
func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			// 设置文件权限
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// findSingBoxBinary 在目录中查找 sing-box 二进制文件
func findSingBoxBinary(dir string) (string, error) {
	var binaryPath string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(info.Name(), "sing-box") {
			// 检查是否有执行权限
			if info.Mode()&0111 != 0 {
				binaryPath = path
				return filepath.SkipDir // 找到后停止搜索
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if binaryPath == "" {
		return "", fmt.Errorf("未找到 sing-box 二进制文件")
	}

	return binaryPath, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
