package api

import (
	"herobox/internal/models"
	"herobox/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServiceHandler 服务处理器
type ServiceHandler struct {
	serviceManager *service.ServiceManager
}

// NewServiceHandler 创建服务处理器
func NewServiceHandler(sm *service.ServiceManager) *ServiceHandler {
	return &ServiceHandler{
		serviceManager: sm,
	}
}

// GetDashboard 获取仪表板数据
func (h *ServiceHandler) GetDashboard(c *gin.Context) {
	data := h.serviceManager.GetDashboardData()
	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// GetAllServices 获取所有服务信息
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	services := h.serviceManager.GetAllServicesInfo()
	c.JSON(http.StatusOK, models.SuccessResponse(services))
}

// GetService 获取单个服务信息
func (h *ServiceHandler) GetService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "服务名称不能为空"))
		return
	}

	serviceInfo := h.serviceManager.GetServiceInfo(serviceName)
	if serviceInfo.Status == models.StatusUnknown && serviceName != "mosdns" && serviceName != "sing-box" {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "服务不存在"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(serviceInfo))
}

// ControlService 控制服务
func (h *ServiceHandler) ControlService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "服务名称不能为空"))
		return
	}

	var req models.ServiceActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "请求参数错误"))
		return
	}

	// 执行服务操作
	err := h.serviceManager.ControlService(serviceName, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	// 获取操作后的服务信息
	serviceInfo := h.serviceManager.GetServiceInfo(serviceName)

	response := models.ServiceActionResponse{
		Success:     true,
		Message:     "操作执行成功",
		ServiceInfo: serviceInfo,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}
