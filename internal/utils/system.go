package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"herobox/internal/models"
)

// RunCommand 执行系统命令
func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// checkBinaryExists 检查二进制文件是否存在
func checkBinaryExists(serviceName string) bool {
	// 常见的二进制文件路径
	var commonPaths []string

	switch serviceName {
	case "mosdns":
		commonPaths = []string{
			"/usr/local/bin/mosdns",
			"/usr/bin/mosdns",
			"/bin/mosdns",
			"/opt/mosdns/mosdns",
		}
	case "sing-box":
		commonPaths = []string{
			"/usr/local/bin/sing-box",
			"/usr/bin/sing-box",
			"/bin/sing-box",
			"/opt/sing-box/sing-box",
		}
	default:
		commonPaths = []string{
			"/usr/local/bin/" + serviceName,
			"/usr/bin/" + serviceName,
			"/bin/" + serviceName,
		}
	}

	// 检查常见路径
	for _, path := range commonPaths {
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			// 检查文件是否可执行
			if stat.Mode()&0111 != 0 {
				return true
			}
		}
	}

	// 尝试使用which命令查找（更可靠的方法）
	output, err := RunCommand("which", serviceName)
	if err == nil && strings.TrimSpace(output) != "" {
		return true
	}

	// 最后尝试直接执行命令看是否存在（带--help参数避免启动）
	_, err = RunCommand(serviceName, "--help")
	// 如果命令存在，即使--help失败也会返回特定的错误，而不是"command not found"
	if err != nil {
		errStr := strings.ToLower(err.Error())
		// 如果包含"not found"或"no such file"，说明命令不存在
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such file") {
			return false
		}
		// 其他错误可能说明命令存在但参数问题
		return true
	}

	return true
}

// checkSystemdServiceExists 检查systemd服务是否存在
func checkSystemdServiceExists(serviceName string) bool {
	// 首先检查systemctl命令是否存在
	_, err := RunCommand("which", "systemctl")
	if err != nil {
		// systemctl不存在，可能不是systemd系统，返回false
		return false
	}

	// 方法1: 使用 systemctl list-unit-files 检查
	listOutput, err := RunCommand("systemctl", "list-unit-files", serviceName+".service", "--no-legend")
	if err == nil && strings.TrimSpace(listOutput) != "" {
		// 如果有输出，说明服务文件存在
		return true
	}

	// 方法2: 使用 systemctl list-units 检查（包括运行时加载的服务）
	listUnitsOutput, err := RunCommand("systemctl", "list-units", serviceName+".service", "--all", "--no-legend")
	if err == nil && strings.TrimSpace(listUnitsOutput) != "" {
		// 如果有输出，说明服务存在（可能是运行时加载的）
		return true
	}

	// 方法3: 直接尝试获取服务状态，如果服务不存在会返回特定错误
	output, err := RunCommand("systemctl", "status", serviceName)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		outputStr := strings.ToLower(output)

		// 检查各种"服务不存在"的错误信息
		notFoundPatterns := []string{
			"not found",
			"could not be found",
			"unit " + serviceName + ".service could not be found",
			"failed to get properties",
			"no such file or directory",
		}

		for _, pattern := range notFoundPatterns {
			if strings.Contains(errStr, pattern) || strings.Contains(outputStr, pattern) {
				return false
			}
		}

		// 其他错误可能说明服务存在但有问题
		return true
	}

	// 如果 systemctl status 成功执行，说明服务存在
	return true
}

// GetServiceStatusWithBinary 获取systemd服务状态（带二进制路径检查）
func GetServiceStatusWithBinary(serviceName, binaryPath string) *models.ServiceInfo {
	info := &models.ServiceInfo{
		Name:   serviceName,
		Status: models.StatusUnknown,
	}

	// 首先检查指定的二进制文件路径
	if binaryPath != "" {
		if stat, err := os.Stat(binaryPath); err == nil && !stat.IsDir() {
			// 指定路径的二进制文件存在
			if stat.Mode()&0111 == 0 {
				// 文件存在但不可执行
				info.Status = models.StatusNotInstalled
				return info
			}
		} else {
			// 指定路径的二进制文件不存在，检查其他常见路径
			if !checkBinaryExists(serviceName) {
				info.Status = models.StatusNotInstalled
				return info
			}
		}
	} else {
		// 没有指定路径，使用通用检查
		if !checkBinaryExists(serviceName) {
			info.Status = models.StatusNotInstalled
			return info
		}
	}

	// 使用多种方式检查systemd服务是否存在
	serviceExists := checkSystemdServiceExists(serviceName)
	if !serviceExists {
		// systemd服务不存在，算作未安装
		info.Status = models.StatusNotInstalled
		return info
	}

	// 服务存在，获取服务状态
	output, err := RunCommand("systemctl", "is-active", serviceName)
	if err != nil {
		// 无法获取状态，但服务存在，可能是服务未启用或有其他问题
		// 进一步检查服务是否已启用
		enableOutput, enableErr := RunCommand("systemctl", "is-enabled", serviceName)
		if enableErr != nil {
			// 无法检查启用状态，可能是服务配置有问题
			info.Status = models.StatusFailed
		} else {
			enableStatus := strings.TrimSpace(enableOutput)
			if enableStatus == "disabled" || enableStatus == "masked" {
				info.Status = models.StatusStopped
			} else {
				info.Status = models.StatusFailed
			}
		}
		return info
	}

	switch strings.TrimSpace(output) {
	case "active":
		info.Status = models.StatusRunning
	case "inactive":
		info.Status = models.StatusStopped
	case "failed":
		info.Status = models.StatusFailed
	case "activating":
		info.Status = models.StatusRunning // 正在启动，算作运行中
	default:
		info.Status = models.StatusUnknown
	}

	// 如果服务正在运行，获取更多信息
	if info.Status == models.StatusRunning {
		// 获取PID
		if pidOutput, err := RunCommand("systemctl", "show", "-p", "MainPID", serviceName); err == nil {
			if strings.Contains(pidOutput, "MainPID=") {
				pidStr := strings.Split(pidOutput, "MainPID=")[1]
				if pid, err := strconv.Atoi(strings.TrimSpace(pidStr)); err == nil && pid > 0 {
					info.PID = pid
				}
			}
		}

		// 获取运行时间
		if statusOutput, err := RunCommand("systemctl", "show", "-p", "ActiveEnterTimestamp", serviceName); err == nil {
			if strings.Contains(statusOutput, "ActiveEnterTimestamp=") {
				timeStr := strings.Split(statusOutput, "ActiveEnterTimestamp=")[1]
				if startTime, err := time.Parse("Mon 2006-01-02 15:04:05 MST", strings.TrimSpace(timeStr)); err == nil {
					uptime := time.Since(startTime)
					info.Uptime = FormatDuration(uptime)
				}
			}
		}
	}

	// 获取版本和平台信息
	getVersionAndPlatform(info, serviceName, binaryPath)

	return info
}

// getVersionAndPlatform 获取服务的版本和平台信息
func getVersionAndPlatform(info *models.ServiceInfo, serviceName, binaryPath string) {
	// 获取二进制文件路径
	var binaryToCheck string
	if binaryPath != "" {
		binaryToCheck = binaryPath
	} else {
		// 查找二进制文件
		binaryToCheck = findBinaryPath(serviceName)
	}

	if binaryToCheck == "" {
		return
	}

	// 尝试获取版本信息
	if version := getVersionFromBinary(binaryToCheck, serviceName); version != "" {
		info.Version = version
	}

	// 获取平台信息
	if platform := getPlatformInfo(); platform != "" {
		info.Platform = platform
	}

	// 获取分支信息（仅针对 sing-box）
	if serviceName == "sing-box" {
		if branch := getBranchInfo(); branch != "" {
			info.Branch = branch
		}
	}
}

// findBinaryPath 查找二进制文件路径
func findBinaryPath(serviceName string) string {
	commonPaths := []string{
		"/usr/local/bin/" + serviceName,
		"/usr/bin/" + serviceName,
		"/opt/" + serviceName + "/" + serviceName,
		"./" + serviceName,
	}

	for _, path := range commonPaths {
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() && stat.Mode()&0111 != 0 {
			return path
		}
	}

	// 尝试使用which命令查找
	if output, err := RunCommand("which", serviceName); err == nil {
		path := strings.TrimSpace(output)
		if path != "" {
			return path
		}
	}

	return ""
}

// getVersionFromBinary 从二进制文件获取版本信息
func getVersionFromBinary(binaryPath, serviceName string) string {
	// 尝试不同的版本命令
	versionCommands := [][]string{
		{binaryPath, "version"},
		{binaryPath, "--version"},
		{binaryPath, "-version"},
		{binaryPath, "-v"},
	}

	for _, cmd := range versionCommands {
		if output, err := RunCommand(cmd[0], cmd[1:]...); err == nil {
			version := parseVersionOutput(output, serviceName)
			if version != "" {
				return version
			}
		}
	}

	return ""
}

// parseVersionOutput 解析版本输出
func parseVersionOutput(output, serviceName string) string {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 对于 sing-box，优先匹配完整版本号格式
		if serviceName == "sing-box" {
			// 首先尝试匹配完整版本号格式: v1.13.0-alpha.17-reF1nd
			re := regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*(?:-[a-zA-Z0-9]+)*)?)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				version := matches[1]
				if !strings.HasPrefix(version, "v") {
					version = "v" + version
				}

				// 读取分支信息并组合
				branchInfo := getBranchInfo()
				if branchInfo != "" && branchInfo != "official-core" && !strings.Contains(version, branchInfo) {
					version = version + "-" + branchInfo
				}

				return version
			}

			// 如果是简单的版本行 (第一行是纯版本号)
			if len(lines) > 0 && regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(line) {
				parts := strings.Fields(line)
				baseVersion := "v" + parts[0]

				// 读取分支信息并组合
				branchInfo := getBranchInfo()
				if branchInfo != "" && branchInfo != "official-core" {
					baseVersion = baseVersion + "-" + branchInfo
				}

				return baseVersion
			}
		}

		// 对于 mosdns
		if serviceName == "mosdns" {
			if strings.Contains(strings.ToLower(line), "mosdns") {
				// 提取 mosdns 版本号
				re := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
				if matches := re.FindStringSubmatch(line); len(matches) > 1 {
					return "v" + matches[1]
				}
			}
		}

		// 通用版本号提取 - 匹配标准版本号格式
		re := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return "v" + matches[1]
		}
	}

	return ""
}

// getPlatformInfo 获取平台信息
func getPlatformInfo() string {
	// 获取操作系统和架构信息
	osInfo := runtime.GOOS
	archInfo := runtime.GOARCH

	// 格式化平台信息
	switch osInfo {
	case "linux":
		osInfo = "Linux"
	case "darwin":
		osInfo = "macOS"
	case "windows":
		osInfo = "Windows"
	}

	switch archInfo {
	case "amd64":
		archInfo = "AMD64"
	case "arm64":
		archInfo = "ARM64"
	case "386":
		archInfo = "x86"
	}

	return fmt.Sprintf("%s/%s", osInfo, archInfo)
}

// getBranchInfo 获取 sing-box 分支信息
func getBranchInfo() string {
	// 读取 /etc/sing-box/version 文件
	versionFilePath := "/etc/sing-box/version"

	content, err := os.ReadFile(versionFilePath)
	if err != nil {
		// 如果文件不存在或读取失败，返回空字符串
		return ""
	}

	// 清理内容，移除空白字符
	branchInfo := strings.TrimSpace(string(content))

	// 如果内容为空，返回空字符串
	if branchInfo == "" {
		return ""
	}

	return branchInfo
}

// GetServiceStatus 获取systemd服务状态（兼容性函数）
func GetServiceStatus(serviceName string) *models.ServiceInfo {
	return GetServiceStatusWithBinary(serviceName, "")
}

// ControlService 控制systemd服务
func ControlService(serviceName string, action models.ServiceAction) error {
	switch action {
	case models.ActionStart, models.ActionStop, models.ActionRestart, models.ActionReload:
		_, err := RunCommand("systemctl", string(action), serviceName)
		return err
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() *models.SystemInfo {
	info := &models.SystemInfo{
		CPUCount: runtime.NumCPU(),
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// 获取操作系统信息
	info.OSInfo = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

	// 获取系统运行时间
	if output, err := RunCommand("uptime", "-p"); err == nil {
		info.Uptime = output
	}

	// 获取内存信息
	if output, err := RunCommand("free", "-m"); err == nil {
		lines := strings.Split(output, "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 1 {
				if total, err := strconv.ParseFloat(fields[1], 64); err == nil {
					info.MemoryTotal = total / 1024 // 转换为GB
				}
			}
		}
	}

	// 获取磁盘使用情况
	info.DiskUsage = make(map[string]interface{})
	if output, err := RunCommand("df", "-h"); err == nil {
		lines := strings.Split(output, "\n")
		for i, line := range lines {
			if i == 0 { // 跳过标题行
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				info.DiskUsage[fields[5]] = map[string]string{
					"size":  fields[1],
					"used":  fields[2],
					"avail": fields[3],
					"use%":  fields[4],
				}
			}
		}
	}

	return info
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天 %d小时 %d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

// BackupFile 备份文件
func BackupFile(filePath, backupDir string) (string, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s/%s_%s", backupDir, timestamp, strings.Replace(filePath, "/", "_", -1))

	// 复制文件
	_, err := RunCommand("cp", filePath, backupPath)
	if err != nil {
		return "", err
	}

	return backupPath, nil
}

// ReadFile 读取文件内容
func ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件内容
func WriteFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

// GetFileInfo 获取文件信息
func GetFileInfo(filePath string) (*models.ConfigFile, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	content, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	modTime := stat.ModTime()
	return &models.ConfigFile{
		Path:         filePath,
		Content:      content,
		LastModified: &modTime,
		Size:         stat.Size(),
	}, nil
}
