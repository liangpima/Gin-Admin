package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/logger"

	"github.com/gin-gonic/gin"
)

const (
	// maskedValue 敏感字段在日志中的占位符
	maskedValue = "******"
	// maxBodyLogLength 请求体落库的最大长度（字符），防止大 body 撑爆日志表
	maxBodyLogLength = 2000
)

// sensitiveNameFragments 判定「字段名 / 配置项名」是否敏感所用的片段（子串匹配）。
//
// 用子串而非全名精确匹配，是为了覆盖 clientSecret、userPassword、apiKey 这类
// 驼峰或带前缀的命名。刻意不包含裸 "key"：它在 {"key":"site.name","value":"..."}
// 这类结构里只是配置项名，本身不是密文。
var sensitiveNameFragments = []string{
	"password", "passwd", "secret", "token",
	"credential", "privatekey", "accesskey", "apiv3key", "secretkey", "pem",
}

// isSensitiveName 判断字段名或配置项名是否敏感。
//
// 先归一化（去下划线/连字符）再匹配，使 access_key、access-key、accessKey、
// accesskey 这几种写法都能命中同一个片段。
func isSensitiveName(name string) bool {
	normalized := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(name))
	for _, frag := range sensitiveNameFragments {
		if strings.Contains(normalized, frag) {
			return true
		}
	}
	return false
}

// isNameField 判断该字段是否为「配置项名称」字段
func isNameField(key string) bool {
	switch strings.ToLower(key) {
	case "key", "configkey", "config_key", "name", "config_name":
		return true
	}
	return false
}

// isValueField 判断该字段是否为「配置项取值」字段
func isValueField(key string) bool {
	switch strings.ToLower(key) {
	case "value", "configvalue", "config_value", "val":
		return true
	}
	return false
}

// maskSensitiveFields 递归替换 JSON 结构中的敏感字段值，返回处理后的值。
//
// 除按字段名脱敏外，还处理两类容易漏掉的情况：
//  1. 「名称 + 取值」分离：配置批量保存的 body 形如
//     {"items":[{"key":"secret_key","value":"真实密钥"}]}，
//     密钥在通用的 value 字段里，必须结合同级的 key 字段判断。
//  2. 嵌套 JSON 字符串：{"data":"{\"password\":\"x\"}"} 这类把 JSON 当字符串传的写法，
//     需要递归解析后再脱敏，否则会被整体跳过。
func maskSensitiveFields(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// 先看同级是否存在敏感的「名称」字段，若有则其对应的「取值」也要脱敏
		sensitiveValue := false
		for k, item := range val {
			if !isNameField(k) {
				continue
			}
			if name, ok := item.(string); ok && isSensitiveName(name) {
				sensitiveValue = true
				break
			}
		}

		for k, item := range val {
			if isSensitiveName(k) || (sensitiveValue && isValueField(k)) {
				val[k] = maskedValue
				continue
			}
			val[k] = maskSensitiveFields(item)
		}
		return val

	case []interface{}:
		for i, item := range val {
			val[i] = maskSensitiveFields(item)
		}
		return val

	case string:
		trimmed := strings.TrimSpace(val)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			var inner interface{}
			if err := json.Unmarshal([]byte(trimmed), &inner); err == nil {
				if out, err := json.Marshal(maskSensitiveFields(inner)); err == nil {
					return string(out)
				}
			}
		}
		return val
	}

	return v
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

	parsed = maskSensitiveFields(parsed)

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

// OperationLogEntry 审计日志条目 —— middleware 侧的最小数据契约。
//
// 为什么不直接用 system 模块的 model.SysOperationLog：
// 中间件属于横切关注点，直接依赖业务模型意味着「业务表加一个字段」
// 会牵动中间件的编译与测试，依赖方向也被倒置。
// 这里只描述审计所需的信息，落库细节（表名、字段映射）交给业务侧适配器。
type OperationLogEntry struct {
	TenantID      uint
	Title         string
	Action        string
	RequestMethod string
	RequestURL    string
	RequestParam  string
	Status        int8
	IP            string
	UserAgent     string
	OperatorID    uint
	OperatorName  string
	CostTime      int64
	ErrorMsg      string
}

// OperationLogWriter 把审计条目落库的能力。
//
// 实现见 internal/module/system/service/operation_log_writer.go，
// 由 cmd/server/main.go 在启动时注入。
type OperationLogWriter interface {
	WriteOperationLog(entry *OperationLogEntry) error
}

// operationLogWriter 已注入的实现。启动阶段写入一次，运行期只读。
var operationLogWriter OperationLogWriter

// SetOperationLogWriter 注入审计日志写入实现，应在开始处理请求之前调用。
func SetOperationLogWriter(w OperationLogWriter) {
	operationLogWriter = w
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

// isJSONContentType 判断请求体是否为 JSON。
//
// 只有 JSON 才需要读取并脱敏落库 —— sanitizeRequestBody 对非 JSON 本来就返回
// "[non-json body omitted]"。若不加这道判断，每个上传请求都会把完整 body
// （按 upload.max_size 最大 10MB）读进内存再丢弃，纯属浪费。
func isJSONContentType(ct string) bool {
	if ct == "" {
		return false
	}
	mediaType := ct
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		mediaType = ct[:i]
	}
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
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
		if c.Request.Body != nil && isJSONContentType(c.GetHeader("Content-Type")) {
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

		entry := &OperationLogEntry{
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
			entry.Status = 0
			entry.ErrorMsg = "HTTP " + strings.TrimSpace(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}

		// 审计写入是旁路能力：实现未注入或写库失败都不应把正常请求变成 500。
		// 但必须留下明确日志 —— 否则「审计静默失效」会长期无人察觉，
		// 等到需要追溯操作记录时才发现一片空白。
		if operationLogWriter == nil {
			logger.Log.Errorf("[operation-log] OperationLogWriter 未注入，审计日志未记录: %s %s", method, path)
			return
		}
		if err := operationLogWriter.WriteOperationLog(entry); err != nil {
			logger.Log.Errorf("[operation-log] 审计日志写入失败: %s %s: %v", method, path, err)
		}
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
