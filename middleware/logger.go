package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerWithWriter 返回一个将日志写入指定 io.Writer 的中间件
func LoggerWithWriter(out io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fmt.Fprintf(out, "[CustomGin LOG] %s %s -> %d (%v)\n", method, path, statusCode, latency)
	}
}
