package main

import (
	"fmt"
	"os"

	"go-admin/config"
	_ "go-admin/docs"
	"go-admin/internal/cache"
	"go-admin/internal/database"
	"go-admin/internal/logger"
	"go-admin/pkg/upload"
	"go-admin/router"
)

// @title Gin-Admin API
// @version 1.0.0
// @description Gin-Admin 后台管理系统 API 文档
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}

func main() {
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := config.Init(configPath); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Log.Sync()

	if err := database.Init(); err != nil {
		logger.Log.Fatalf("初始化数据库失败: %v", err)
	}

	if err := cache.Init(); err != nil {
		logger.Log.Warnf("初始化Redis失败(可选): %v", err)
	}

	upload.Init()

	r := router.Setup(config.Cfg.Server.Mode)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	logger.Log.Infof("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Log.Fatalf("服务启动失败: %v", err)
	}
}
