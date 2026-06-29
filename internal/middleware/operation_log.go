package middleware

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"github.com/gin-gonic/gin"
)

var (
_operationLogRepo     repository.LogRepository
_operationLogRepoOnce sync.Once
)

func getOperationLogRepo() repository.LogRepository {
	_operationLogRepoOnce.Do(func() {
		_operationLogRepo = repository.NewLogRepository()
	})
	return _operationLogRepo
}

var skipPaths = []string{
	"/api/v1/auth/userInfo",
	"/api/v1/system/log",
	"/api/v1/captcha",
	"/uploads/",
}

// 敏感 GET 路径也需要记录日志
var sensitiveGetPaths = []string{
	"/api/v1/system/user/",
	"/api/v1/system/role/",
	"/api/v1/system/config/",
	"/api/v1/system/file/",
	"/api/v1/member/",
	"/api/v1/system/pay/",
}

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 非敏感 GET 请求跳过
		if c.Request.Method == "GET" {
			isSensitive := false
			for _, sp := range sensitiveGetPaths {
				if strings.HasPrefix(path, sp) {
					isSensitive = true
					break
				}
			}
			if !isSensitive {
				c.Next()
				return
			}
		}

		for _, skip := range skipPaths {
			if strings.HasPrefix(path, skip) {
				c.Next()
				return
			}
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		start := time.Now()
		c.Next()
		latency := time.Since(start).Milliseconds()

		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := common.NormalizeIP(c.ClientIP())
		userAgent := c.Request.UserAgent()

		operatorID := common.GetCurrentUserID(c)
		operatorName := common.GetCurrentUsername(c)

		title := resolveTitle(path)

		log := &model.SysOperationLog{
			TenantID:      common.GetTenantID(c),
			Title:         title,
			Action:        method,
			RequestMethod: method,
			RequestURL:    path,
			RequestParam:  string(bodyBytes),
			Status:        1,
			IP:            clientIP,
			UserAgent:     userAgent,
			OperatorID:    operatorID,
			OperatorName:  operatorName,
			CostTime:      latency,
		}

		if statusCode >= 400 {
			log.Status = 0
			log.ErrorMsg = "HTTP " + strings.TrimSpace(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}

		_ = getOperationLogRepo().CreateOperationLog(log)
	}
}

func resolveTitle(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		module := parts[3]
		resource := ""
		if len(parts) >= 5 {
			resource = parts[4]
		}
		if resource != "" && resource != "list" {
			return module + "-" + resource
		}
		return module
	}
	return "unknown"
}
