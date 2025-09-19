package service

import (
	"fmt"
	"herobox/internal/config"
	"herobox/internal/models"
	"herobox/internal/utils"
	"os"
	"strings"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	config *config.Config
}

// NewServiceManager 创建服务管理器
func NewServiceManager(cfg *config.Config) *ServiceManager {
	return &ServiceManager{
		config: cfg,
	}
}

// GetConfig 获取配置
func (sm *ServiceManager) GetConfig() *config.Config {
	return sm.config
}

// GetServiceInfo 获取服务信息
func (sm *ServiceManager) GetServiceInfo(serviceName string) *models.ServiceInfo {
	var realServiceName string
	var binaryPath string
	
	switch serviceName {
	case "mosdns":
		realServiceName = sm.config.MosdnsServiceName
		binaryPath = sm.config.MosdnsBinaryPath
	case "sing-box":
		realServiceName = sm.config.SingBoxServiceName
		binaryPath = sm.config.SingBoxBinaryPath
	default:
		return &models.ServiceInfo{
			Name:   serviceName,
			Status: models.StatusUnknown,
		}
	}

	return utils.GetServiceStatusWithBinary(realServiceName, binaryPath)
}

// GetAllServicesInfo 获取所有服务信息
func (sm *ServiceManager) GetAllServicesInfo() map[string]*models.ServiceInfo {
	services := make(map[string]*models.ServiceInfo)
	
	services["mosdns"] = sm.GetServiceInfo("mosdns")
	services["sing-box"] = sm.GetServiceInfo("sing-box")
	
	return services
}

// ControlService 控制服务
func (sm *ServiceManager) ControlService(serviceName string, action models.ServiceAction) error {
	var realServiceName string
	switch serviceName {
	case "mosdns":
		realServiceName = sm.config.MosdnsServiceName
	case "sing-box":
		realServiceName = sm.config.SingBoxServiceName
	default:
		return fmt.Errorf("未知的服务: %s", serviceName)
	}

	return utils.ControlService(realServiceName, action)
}

// GetDashboardData 获取仪表板数据
func (sm *ServiceManager) GetDashboardData() *models.DashboardData {
	return &models.DashboardData{
		SystemInfo: utils.GetSystemInfo(),
		Services:   sm.GetAllServicesInfo(),
		RecentLogs: sm.getRecentLogs(),
	}
}

// getRecentLogs 获取最近的日志
func (sm *ServiceManager) getRecentLogs() map[string]string {
	logs := make(map[string]string)
	
	// 获取mosdns日志
	if content, err := sm.GetLogContent("mosdns", 10, ""); err == nil {
		logs["mosdns"] = content
	}
	
	// 获取sing-box日志
	if content, err := sm.GetLogContent("sing-box", 10, ""); err == nil {
		logs["sing-box"] = content
	}
	
	return logs
}

// GetLogContent 获取日志内容
func (sm *ServiceManager) GetLogContent(serviceName string, lines int, filter string) (string, error) {
	var logPath string
	switch serviceName {
	case "mosdns":
		logPath = sm.config.MosdnsLogPath
	case "sing-box":
		logPath = sm.config.SingBoxLogPath
	default:
		return "", fmt.Errorf("未知的服务: %s", serviceName)
	}

	// 检查日志文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Sprintf("日志文件不存在: %s\n\n提示：\n1. 请确保 %s 服务已启动\n2. 检查服务配置中的日志路径设置\n3. 如果是首次运行，日志文件可能还未创建", logPath, serviceName), nil
	}

	// 使用tail命令获取日志
	args := []string{"-n", fmt.Sprintf("%d", lines)}
	if filter != "" {
		// 使用grep过滤
		output, err := utils.RunCommand("tail", append(args, logPath)...)
		if err != nil {
			return fmt.Sprintf("读取日志失败: %v\n日志路径: %s", err, logPath), nil
		}
		
		// 过滤内容
		filteredOutput, err := utils.RunCommand("grep", filter)
		if err != nil {
			return output, nil // 如果grep失败，返回原始内容
		}
		return filteredOutput, nil
	}
	
	output, err := utils.RunCommand("tail", append(args, logPath)...)
	if err != nil {
		return fmt.Sprintf("读取日志失败: %v\n日志路径: %s", err, logPath), nil
	}
	return output, nil
}

// GetConfigFile 获取配置文件
func (sm *ServiceManager) GetConfigFile(serviceName string) (*models.ConfigFile, error) {
	var configPath string
	switch serviceName {
	case "mosdns":
		configPath = sm.config.MosdnsConfigPath
	case "sing-box":
		configPath = sm.config.SingBoxConfigPath
	default:
		return nil, fmt.Errorf("未知的服务: %s", serviceName)
	}

	return utils.GetFileInfo(configPath)
}

// UpdateConfigFile 更新配置文件
func (sm *ServiceManager) UpdateConfigFile(serviceName, content string, backup bool) (string, error) {
	var configPath string
	switch serviceName {
	case "mosdns":
		configPath = sm.config.MosdnsConfigPath
	case "sing-box":
		configPath = sm.config.SingBoxConfigPath
	default:
		return "", fmt.Errorf("未知的服务: %s", serviceName)
	}

	var backupPath string
	var err error

	// 创建备份
	if backup {
		backupPath, err = utils.BackupFile(configPath, sm.config.BackupDir)
		if err != nil {
			return "", fmt.Errorf("备份文件失败: %v", err)
		}
	}

	// 写入新内容
	if err := utils.WriteFile(configPath, content); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %v", err)
	}

	return backupPath, nil
}

// ValidateConfig 验证配置文件
func (sm *ServiceManager) ValidateConfig(serviceName, content string) error {
	// 这里可以添加配置文件格式验证逻辑
	switch serviceName {
	case "mosdns":
		// 可以添加YAML格式验证
		if !strings.Contains(content, "plugins:") {
			return fmt.Errorf("mosdns配置文件格式可能不正确，缺少plugins配置")
		}
	case "sing-box":
		// 可以添加JSON格式验证
		if !strings.Contains(content, "{") || !strings.Contains(content, "}") {
			return fmt.Errorf("sing-box配置文件格式可能不正确，不是有效的JSON格式")
		}
	}
	return nil
}
