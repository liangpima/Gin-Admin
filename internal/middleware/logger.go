package middleware

import (
	"time"

	"go-admin/internal/common"
	"go-admin/internal/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 这里**刻意不读取请求体**。
		//
		// 早前这里 io.ReadAll 了整个 body 再塞回 c.Request.Body，但读取结果
		// 从未被使用 —— 访问日志只记录 status/method/path/ip/latency。
		// 代价是每个请求都在内存里多复制一份 body：上传接口按 upload.max_size
		// 最大 10MB，N 个并发上传就是 10N MB 的额外占用，而需要落库的请求体
		// 已由 OperationLog 中间件按需读取。
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := common.NormalizeIP(c.ClientIP())
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		logger.Log.Infow("request",
			"status", statusCode,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency", latency.String(),
			"errors", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}
