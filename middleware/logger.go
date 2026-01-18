package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 是一个简单的请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 执行 handler
		c.Next()

		// 记录日志
		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

		// 使用 println 输出到终端（不影响 HTTP 响应）
		println("[GIN LOG]", method, path, "->", statusCode, "(", latency, ")")
	}
}
