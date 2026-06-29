package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"go-admin/internal/common"
	"go-admin/internal/logger"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				sanitizedRequest := sanitizeRequest(string(httpRequest))

				if brokenPipe {
					logger.Log.Errorw("panic",
						"error", err,
						"request", sanitizedRequest,
					)
					c.Abort()
					return
				}

				logger.Log.Errorw("panic",
					"error", err,
					"request", sanitizedRequest,
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, common.Response{
					Code:    common.CodeInternalError,
					Message: "服务器内部错误",
					Data:    nil,
				})
			}
		}()

		c.Next()
	}
}

// sanitizeRequest 脱敏请求中的敏感 Header
func sanitizeRequest(req string) string {
	lines := strings.Split(req, "\n")
	for i, line := range lines {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "authorization:") {
			lines[i] = "Authorization: Bearer [REDACTED]"
		}
		if strings.HasPrefix(lower, "cookie:") {
			lines[i] = "Cookie: [REDACTED]"
		}
	}
	return strings.Join(lines, "\n")
}
