package middleware

import (
	"time"

	"go-admin/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	cfg := config.Cfg.CORS

	// 生产环境必须配置白名单，开发环境允许所有来源
	allowAllOrigins := len(cfg.AllowOrigins) == 0 && !config.IsProduction()

	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-Id"}
	}
	if len(cfg.ExposeHeaders) == 0 {
		cfg.ExposeHeaders = []string{"Content-Length", "Content-Disposition"}
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
