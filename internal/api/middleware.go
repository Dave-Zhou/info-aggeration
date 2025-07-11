package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"example.com/m/v2/internal/utils"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	
	return cors.New(config)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		// 获取请求信息
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 记录日志
		logger.Info("API请求",
			"method", reqMethod,
			"uri", reqUri,
			"status", statusCode,
			"latency", latencyTime,
			"ip", clientIP,
		)
	}
}

// AuthMiddleware 认证中间件（预留）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现JWT认证逻辑
		c.Next()
	}
}

// RateLimitMiddleware 限流中间件（预留）
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现限流逻辑
		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(500, gin.H{
				"error":   "Internal Server Error",
				"message": err,
			})
		}
		c.AbortWithStatus(500)
	})
} 