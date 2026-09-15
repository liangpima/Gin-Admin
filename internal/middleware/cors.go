package middleware

import (
	"time"

	"go-admin/config"
	"go-admin/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	cfg := config.Cfg.CORS

	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-Id"}
	}
	if len(cfg.ExposeHeaders) == 0 {
		cfg.ExposeHeaders = []string{"Content-Length", "Content-Disposition"}
	}

	// 未配置来源白名单时的降级策略。
	// 注意：AllowAllOrigins 与 AllowCredentials 同时开启时，浏览器会按请求
	// Origin 回显，等价于对任意站点开放「带凭证」访问，因此必须避免该组合。
	allowAllOrigins := false
	if len(cfg.AllowOrigins) == 0 {
		switch {
		case config.IsProduction():
			// 生产环境未配置白名单：拒绝全部跨域请求，宁可不可用也不误开放
			logger.Log.Errorf("[cors] 生产环境未配置 cors.allow_origins，将拒绝全部跨域请求")
		case cfg.AllowCredentials:
			logger.Log.Warnf("[cors] 未配置 cors.allow_origins 且启用了 allow_credentials，" +
				"已按同源策略收紧；如需跨域请显式配置来源白名单")
		default:
			allowAllOrigins = true
		}
	}

	return cors.New(cors.Config{
		AllowAllOrigins:  allowAllOrigins,
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
