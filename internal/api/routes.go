package api

import (
	"database/sql"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/utils"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置API路由
func SetupRoutes(r *gin.RouterGroup, db *sql.DB, logger utils.Logger, cfg *config.Config) {
	// 创建控制器
	siteController := NewSiteController(db, logger, cfg)
	taskController := NewTaskController(db, logger, cfg)
	dataController := NewDataController(db, logger)
	systemController := NewSystemController(db, logger, cfg)

	// 站点管理路由
	sites := r.Group("/sites")
	{
		sites.GET("", siteController.ListSites)
		sites.POST("", siteController.CreateSite)
		sites.GET("/:id", siteController.GetSite)
		sites.PUT("/:id", siteController.UpdateSite)
		sites.DELETE("/:id", siteController.DeleteSite)
		sites.POST("/:id/test", siteController.TestSite)
		sites.PUT("/:id/toggle", siteController.ToggleSite)
		sites.POST("/:id/run", siteController.RunSiteTask) // 添加运行任务的路由
	}

	// 任务管理路由
	tasks := r.Group("/tasks")
	{
		tasks.GET("", taskController.ListTasks)
		tasks.POST("", taskController.CreateTask)
		tasks.GET("/:id", taskController.GetTask)
		tasks.PUT("/:id", taskController.UpdateTask)
		tasks.DELETE("/:id", taskController.DeleteTask)
		tasks.POST("/:id/start", taskController.StartTask)
		tasks.POST("/:id/stop", taskController.StopTask)
		tasks.GET("/:id/logs", taskController.GetTaskLogs)
		tasks.GET("/:id/status", taskController.GetTaskStatus)
	}

	// 数据管理路由
	data := r.Group("/data")
	{
		data.GET("/items", dataController.ListItems)
		data.GET("/items/:id", dataController.GetItem)
		data.DELETE("/items/:id", dataController.DeleteItem)
		data.GET("/items/export", dataController.ExportItems)
		data.GET("/statistics", dataController.GetStatistics)
		data.POST("/items/search", dataController.SearchItems)
	}

	// 系统管理路由
	system := r.Group("/system")
	{
		system.GET("/status", systemController.GetSystemStatus)
		system.GET("/config", systemController.GetConfig)
		system.PUT("/config", systemController.UpdateConfig)
		system.GET("/logs", systemController.GetLogs)
		system.POST("/backup", systemController.CreateBackup)
		system.GET("/backups", systemController.ListBackups)
		system.POST("/restore", systemController.RestoreBackup)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "服务正常运行",
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":    "1.0.0",
			"build_time": "2023-01-01",
			"go_version": "1.21",
		})
	})
}
