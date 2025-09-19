package api

import (
	"herobox/internal/models"
	"herobox/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志处理器
type LogHandler struct {
	serviceManager *service.ServiceManager
}

// NewLogHandler 创建日志处理器
func NewLogHandler(sm *service.ServiceManager) *LogHandler {
	return &LogHandler{
		serviceManager: sm,
	}
}

// GetLogs 获取日志
func (h *LogHandler) GetLogs(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "服务名称不能为空"))
		return
	}

	if serviceName != "mosdns" && serviceName != "sing-box" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "不支持的服务"))
		return
	}

	// 解析查询参数
	linesStr := c.DefaultQuery("lines", "100")
	lines, err := strconv.Atoi(linesStr)
	if err != nil || lines <= 0 {
		lines = 100
	}
	if lines > 10000 {
		lines = 10000 // 限制最大行数
	}

	filterKeyword := c.Query("filter")

	// 获取日志内容
	content, err := h.serviceManager.GetLogContent(serviceName, lines, filterKeyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	// 计算行数
	totalLines := len([]rune(content))
	filteredLines := totalLines

	response := models.LogResponse{
		Content:       content,
		TotalLines:    totalLines,
		FilteredLines: filteredLines,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}
