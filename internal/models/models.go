package models

import "time"

// ServiceStatus 服务状态
type ServiceStatus string

const (
	StatusRunning     ServiceStatus = "running"
	StatusStopped     ServiceStatus = "stopped"
	StatusFailed      ServiceStatus = "failed"
	StatusNotInstalled ServiceStatus = "not_installed"
	StatusUnknown     ServiceStatus = "unknown"
)

// ServiceAction 服务操作
type ServiceAction string

const (
	ActionStart   ServiceAction = "start"
	ActionStop    ServiceAction = "stop"
	ActionRestart ServiceAction = "restart"
	ActionReload  ServiceAction = "reload"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name   string        `json:"name"`
	Status ServiceStatus `json:"status"`
	PID    int           `json:"pid,omitempty"`
	Uptime string        `json:"uptime,omitempty"`
}

// ServiceActionRequest 服务操作请求
type ServiceActionRequest struct {
	Action ServiceAction `json:"action" binding:"required"`
}

// ServiceActionResponse 服务操作响应
type ServiceActionResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	ServiceInfo *ServiceInfo `json:"service_info,omitempty"`
}

// ConfigFile 配置文件
type ConfigFile struct {
	Path         string     `json:"path"`
	Content      string     `json:"content"`
	BackupPath   string     `json:"backup_path,omitempty"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Size         int64      `json:"size,omitempty"`
}

// ConfigUpdateRequest 配置更新请求
type ConfigUpdateRequest struct {
	Content string `json:"content" binding:"required"`
	Backup  bool   `json:"backup"`
}

// ConfigUpdateResponse 配置更新响应
type ConfigUpdateResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	BackupPath string `json:"backup_path,omitempty"`
}

// LogRequest 日志查询请求
type LogRequest struct {
	Lines         int    `json:"lines" form:"lines"`
	FilterKeyword string `json:"filter_keyword" form:"filter_keyword"`
}

// LogResponse 日志查询响应
type LogResponse struct {
	Content       string `json:"content"`
	TotalLines    int    `json:"total_lines"`
	FilteredLines int    `json:"filtered_lines"`
}


// SystemInfo 系统信息
type SystemInfo struct {
	Hostname    string                 `json:"hostname"`
	OSInfo      string                 `json:"os_info"`
	CPUCount    int                    `json:"cpu_count"`
	MemoryTotal float64                `json:"memory_total"`
	DiskUsage   map[string]interface{} `json:"disk_usage"`
	Uptime      string                 `json:"uptime"`
}

// DashboardData 仪表板数据
type DashboardData struct {
	SystemInfo *SystemInfo             `json:"system_info"`
	Services   map[string]*ServiceInfo `json:"services"`
	RecentLogs map[string]string       `json:"recent_logs"`
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error 错误响应
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}
