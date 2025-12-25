package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 请求日志中间件
// 记录每个请求的基本信息，便于调试和监控
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)

		// 获取响应状态
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 格式化日志
		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[API] %3d | %13v | %15s | %-7s %s",
			status,
			latency,
			clientIP,
			method,
			path,
		)

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[ERROR] %s", e.Error())
			}
		}
	}
}

// RecoveryMiddleware panic 恢复中间件
// 捕获 panic，防止服务崩溃
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}
