package middleware

import (
	"context"
	"fmt"
	"strings"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/pkg/auth"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.Unauthorized(c, "Token格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查 Token 是否已被吊销
		if cache.IsTokenRevoked(context.Background(), tokenString) {
			common.Unauthorized(c, "Token已失效")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			common.Unauthorized(c, "Token无效或已过期")
			c.Abort()
			return
		}

		// 检查用户级别 Token 吊销（密码修改/禁用）
		userRevoked, _ := cache.Exists(context.Background(),
			fmt.Sprintf("user:token_revoked:%d", claims.UserID))
		if userRevoked {
			common.Unauthorized(c, "Token已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(common.ContextKeyUserID, claims.UserID)
		c.Set(common.ContextKeyUsername, claims.Username)
		c.Set(common.ContextKeyTenantID, claims.TenantID)
		c.Set(common.ContextKeyDeptID, claims.DeptID)

		c.Next()
	}
}
