package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-admin/config"
	_ "go-admin/docs"
	"go-admin/internal/cache"
	"go-admin/internal/database"
	"go-admin/internal/logger"
	"go-admin/internal/middleware"
	"go-admin/internal/module/system/service"
	"go-admin/pkg/task"
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

	// 生产环境若仍使用默认密钥则拒绝启动（默认密钥可导致 JWT 被伪造）
	if err := config.ValidateSecurity(); err != nil {
		fmt.Printf("%v\n", err)
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

	// 初始化 Casbin 权限模型并同步策略。
	// 鉴权属于安全控制，初始化失败时拒绝启动，避免在"无鉴权"状态下对外提供服务。
	if err := middleware.InitCasbin(config.Cfg.Casbin.ModelPath); err != nil {
		logger.Log.Fatalf("初始化Casbin失败: %v", err)
	}

	upload.Init(service.LoadOSSConfig())

	r := router.Setup(config.Cfg.Server.Mode)

	// 注册日志清理定时任务：每天 03:00 清理超过保留期的操作日志与登录日志，
	// 避免日志表无限增长（保留天数见 log.db_retention_days，<=0 表示不清理）。
	// 多实例部署时每个实例都会执行，删除操作幂等，影响仅为重复执行。
	logService := service.NewLogService()
	retentionDays := config.Cfg.Log.DBRetentionDays
	if retentionDays <= 0 {
		logger.Log.Infof("日志保留天数配置为 %d，已跳过日志清理任务", retentionDays)
	} else if _, err := task.AddJob("0 3 * * *", func() {
		deleted, err := logService.CleanExpiredLogs(retentionDays)
		if err != nil {
			logger.Log.Errorf("清理超期日志失败: %v", err)
			return
		}
		if deleted > 0 {
			logger.Log.Infof("已清理 %d 条超期日志", deleted)
		}
	}); err != nil {
		logger.Log.Warnf("注册日志清理任务失败: %v", err)
	} else {
		logger.Log.Infof("已注册日志清理任务：每天 03:00 清理 %d 天前的操作/登录日志", retentionDays)
	}
	task.Start()

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(config.Cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 在独立协程中启动 HTTP 服务，主协程负责监听退出信号
	go func() {
		logger.Log.Infof("服务启动在 %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号，执行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Log.Info("正在关闭服务...")

	// 停止定时任务调度
	task.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Errorf("HTTP 服务关闭异常: %v", err)
	}

	// 关闭数据库连接池
	if sqlDB, err := database.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Errorf("关闭数据库连接失败: %v", err)
		}
	}

	logger.Log.Info("服务已退出")
}
