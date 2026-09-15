package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"github.com/gin-gonic/gin"
)

const (
	// maskedValue 敏感字段在日志中的占位符
	maskedValue = "******"
	// maxBodyLogLength 请求体落库的最大长度（字符），防止大 body 撑爆日志表
	maxBodyLogLength = 2000
)

// sensitiveBodyFields 命中即脱敏的字段名（按小写比较）。
// 直接记录原始请求体会把明文密码写入日志表——创建用户、重置密码、
// 修改密码这几个接口的 body 里都带密码，因此必须先脱敏再落库。
var sensitiveBodyFields = map[string]bool{
	"password":        true,
	"oldpassword":     true,
	"newpassword":     true,
	"confirmpassword": true,
	"secret":          true,
	"secretkey":       true,
	"accesskey":       true,
	"privatekey":      true,
	"apiv3key":        true,
	"token":           true,
	"accesstoken":     true,
	"refreshtoken":    true,
}

// maskSensitiveFields 递归替换 JSON 结构中的敏感字段值
func maskSensitiveFields(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, item := range val {
			if sensitiveBodyFields[strings.ToLower(k)] {
				val[k] = maskedValue
				continue
			}
			maskSensitiveFields(item)
		}
	case []interface{}:
		for _, item := range val {
			maskSensitiveFields(item)
		}
	}
}

// sanitizeRequestBody 对请求体脱敏并按长度截断后再落库
func sanitizeRequestBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		// 非 JSON（如文件上传），不记录具体内容
		return "[non-json body omitted]"
	}

	maskSensitiveFields(parsed)

	out, err := json.Marshal(parsed)
	if err != nil {
		return "[unparsable body omitted]"
	}

	// 按字符截断，避免截断多字节字符
	runes := []rune(string(out))
	if len(runes) > maxBodyLogLength {
		return string(runes[:maxBodyLogLength]) + "...[truncated]"
	}
	return string(runes)
}

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
			RequestParam:  sanitizeRequestBody(bodyBytes),
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
