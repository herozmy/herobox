package config

import (
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	// 基础配置
	Port        string `json:"port"`
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`


	// 服务配置
	MosdnsServiceName  string `json:"mosdns_service_name"`
	SingBoxServiceName string `json:"sing_box_service_name"`

	// 二进制文件路径
	MosdnsBinaryPath  string `json:"mosdns_binary_path"`
	SingBoxBinaryPath string `json:"sing_box_binary_path"`

	// 配置文件路径
	MosdnsConfigPath  string `json:"mosdns_config_path"`
	SingBoxConfigPath string `json:"sing_box_config_path"`

	// 日志文件路径
	MosdnsLogPath  string `json:"mosdns_log_path"`
	SingBoxLogPath string `json:"sing_box_log_path"`

	// 高级配置
	MockMode       bool `json:"mock_mode"`
	ServiceTimeout int  `json:"service_timeout"`
	MaxLogLines    int  `json:"max_log_lines"`

	// 备份目录
	BackupDir string `json:"backup_dir"`

	// Web静态文件目录
	WebDir string `json:"web_dir"`
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		MosdnsServiceName:  getEnv("MOSDNS_SERVICE_NAME", "mosdns"),
		SingBoxServiceName: getEnv("SING_BOX_SERVICE_NAME", "sing-box"),

		MosdnsBinaryPath:  getEnv("MOSDNS_BINARY_PATH", "/usr/local/bin/mosdns"),
		SingBoxBinaryPath: getEnv("SING_BOX_BINARY_PATH", "/usr/local/bin/sing-box"),

		MosdnsConfigPath:  getEnv("MOSDNS_CONFIG_PATH", "/etc/mosdns/config.yaml"),
		SingBoxConfigPath: getEnv("SING_BOX_CONFIG_PATH", "/etc/sing-box/config.json"),

		MosdnsLogPath:  getEnv("MOSDNS_LOG_PATH", "/var/log/mosdns.log"),
		SingBoxLogPath: getEnv("SING_BOX_LOG_PATH", "/var/log/sing-box.log"),

		MockMode:       getEnvBool("MOCK_MODE", false),
		ServiceTimeout: getEnvInt("SERVICE_TIMEOUT", 30),
		MaxLogLines:    getEnvInt("MAX_LOG_LINES", 10000),

		BackupDir: getEnv("BACKUP_DIR", "./backups"),
		WebDir:    getEnv("WEB_DIR", "./web/dist"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数类型的环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool 获取布尔类型的环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
