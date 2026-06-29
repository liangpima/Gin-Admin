package middleware

import (
	"github.com/gin-gonic/gin"
)

// Tenant 中间件不再从 Header/Query 读取 tenant_id
// 租户 ID 仅从 JWT claims 获取（在 Auth 中间件中设置）
func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
