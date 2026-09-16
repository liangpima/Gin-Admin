package middleware

import (
	"context"
	"fmt"
	"strings"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/logger"
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

		// 检查 Token 是否已被吊销。
		// 查询失败时**拒绝**而不是放行：Redis 抖动期间放行等于让已登出的
		// token 重新生效（fail-open），而这恰恰是吊销机制最该起作用的时刻。
		revoked, err := cache.IsTokenRevoked(context.Background(), tokenString)
		if err != nil {
			logger.Log.Errorf("[auth] 查询 token 吊销状态失败: %v", err)
			common.Error(c, common.CodeInternalError, "鉴权服务暂时不可用，请稍后重试")
			c.Abort()
			return
		}
		if revoked {
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

		// 检查用户级别 Token 吊销（密码修改/禁用）。同样 fail-closed：
		// 早前用 `userRevoked, _ :=` 吞掉错误，Redis 异常时会当成「未吊销」，
		// 已停用账号仍可继续访问。
		userRevoked, existsErr := cache.Exists(context.Background(),
			fmt.Sprintf("user:token_revoked:%d", claims.UserID))
		if existsErr != nil {
			logger.Log.Errorf("[auth] 查询用户级 token 吊销标记失败: %v", existsErr)
			common.Error(c, common.CodeInternalError, "鉴权服务暂时不可用，请稍后重试")
			c.Abort()
			return
		}
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
