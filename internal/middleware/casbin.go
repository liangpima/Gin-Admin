package middleware

import (
	"net/http"

	"go-admin/internal/common"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

var enforcer *casbin.Enforcer

func InitCasbin(modelPath string) error {
	var err error
	enforcer, err = casbin.NewEnforcer(modelPath)
	if err != nil {
		return err
	}
	return nil
}

func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			c.Next()
			return
		}

		// 如果没有加载任何策略，跳过权限检查（开发阶段兼容）
		policies, _ := enforcer.GetPolicy()
		if len(policies) == 0 {
			c.Next()
			return
		}

		// 从 context 获取用户角色（由 Auth 中间件设置）
		// 注意：当前 JWT 不包含角色信息，需要扩展 JWT claims 或从 DB 查询
		// 临时方案：使用 username 作为 subject 进行匹配
		username := common.GetCurrentUsername(c)
		if username == "" {
			common.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		object := c.Request.URL.Path
		action := c.Request.Method

		// 检查是否有该用户的策略
		ok, _ := enforcer.Enforce(username, object, action)
		if !ok {
			// 检查是否为 admin 角色（超级管理员跳过检查）
			adminOk, _ := enforcer.Enforce("admin", object, action)
			if !adminOk {
				c.JSON(http.StatusForbidden, common.Response{
					Code:    common.CodeForbidden,
					Message: "没有权限访问",
					Data:    nil,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
