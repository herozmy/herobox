package api

import (
	"herobox/internal/config"
	"herobox/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 添加CORS中间件
	router.Use(CORSMiddleware())

	// 创建服务管理器
	serviceManager := service.NewServiceManager(cfg)

	// 创建处理器
	serviceHandler := NewServiceHandler(serviceManager)
	configHandler := NewConfigHandler(serviceManager)
	logHandler := NewLogHandler(serviceManager)

	// 静态文件服务 - 修复路径配置
	router.Static("/assets", cfg.WebDir+"/assets")
	router.StaticFile("/favicon.ico", cfg.WebDir+"/favicon.ico")

	// 处理前端路由 - 所有非API请求都返回index.html
	router.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if gin.Mode() == gin.DebugMode {
			println("NoRoute hit for:", c.Request.URL.Path)
		}

		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// 其他请求返回index.html (用于前端路由)
		c.File(cfg.WebDir + "/index.html")
	})

	// API路由组 - 移除认证，内网直接使用
	api := router.Group("/api")
	{
		// 仪表板
		api.GET("/dashboard", serviceHandler.GetDashboard)

		// 服务管理
		services := api.Group("/services")
		{
			services.GET("/", serviceHandler.GetAllServices)
			services.GET("/:name", serviceHandler.GetService)
			services.POST("/:name/action", serviceHandler.ControlService)
		}

		// 配置管理
		configs := api.Group("/config")
		{
			configs.GET("/:service", configHandler.GetConfig)
			configs.PUT("/:service", configHandler.UpdateConfig)
		}

		// 日志查看
		logs := api.Group("/logs")
		{
			logs.GET("/:service", logHandler.GetLogs)
		}

		// Sing-Box 配置管理
		singbox := api.Group("/singbox")
		{
			singbox.GET("/config", configHandler.GetSingBoxConfig)
			singbox.PUT("/config", configHandler.UpdateSingBoxConfig)
			singbox.POST("/config/validate", configHandler.ValidateSingBoxConfig)
			singbox.POST("/config/validate-current", configHandler.ValidateCurrentSingBoxConfig)

			// 分模块获取配置
			singbox.GET("/inbounds", configHandler.GetSingBoxInbounds)
			singbox.GET("/outbounds", configHandler.GetSingBoxOutbounds)
			singbox.POST("/outbounds", configHandler.CreateSingBoxOutbound)
			singbox.PUT("/outbounds/:id", configHandler.UpdateSingBoxOutbound)
			singbox.DELETE("/outbounds/:id", configHandler.DeleteSingBoxOutbound)
			singbox.POST("/outbounds/validate", configHandler.ValidateOutboundsChanges)
			singbox.POST("/outbounds/batch-save", configHandler.BatchSaveOutbounds)
			singbox.POST("/restart", configHandler.RestartSingBoxService)
			singbox.GET("/rules", configHandler.GetSingBoxRules)
			singbox.POST("/rules/route", configHandler.CreateRouteRule)
			singbox.PUT("/rules/route/:id", configHandler.UpdateRouteRule)
			singbox.DELETE("/rules/route/:id", configHandler.DeleteRouteRule)
			// 路由规则排序
			singbox.POST("/rules/route/:id/move-up", configHandler.MoveRouteRuleUp)
			singbox.POST("/rules/route/:id/move-down", configHandler.MoveRouteRuleDown)
			singbox.POST("/rules/route/reorder", configHandler.ReorderRouteRules)

			// 规则集管理
			singbox.POST("/rulesets", configHandler.CreateRuleSet)
			singbox.PUT("/rulesets/:id", configHandler.UpdateRuleSet)
			singbox.DELETE("/rulesets/:id", configHandler.DeleteRuleSet)

			// 内核更新管理
			kernel := singbox.Group("/kernel")
			{
				kernel.GET("/detect-path", configHandler.DetectSingBoxPath)
				kernel.GET("/check-update", configHandler.CheckSingBoxUpdate)
				kernel.POST("/update", configHandler.UpdateSingBoxKernel)
				kernel.GET("/update-stream", configHandler.UpdateSingBoxKernelStream)
			}
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "HeroBox运行正常",
		})
	})

	return router
}
