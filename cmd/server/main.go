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

	// 开发模式下 Swagger 与 gin 调试输出是开启的（见 router.Setup）。
	// 这里显式告警：生产部署若忘记改 mode，这些调试入口会直接暴露。
	if !config.IsProduction() {
		logger.Log.Warnf("当前为开发模式(mode=%s)：Swagger 文档与调试信息已开启，生产环境请设置 mode=release",
			config.Cfg.Server.Mode)
	}

	if err := database.Init(); err != nil {
		logger.Log.Fatalf("初始化数据库失败: %v", err)
	}

	// Redis 是**必需**依赖，不可降级：它承载 refresh token 存储、token 黑名单、
	// 登录失败限频、验证码与角色缓存。登录流程本身就会写 refresh token，
	// Redis 不可用时登录直接失败；若这里只告警而继续启动，
	// cache.RDB 为 nil 会让后续每次调用空指针 panic，被 Recovery 兜成 500，
	// 相当于全站不可用且错误信息毫无指向性。失败即退出。
	if err := cache.Init(); err != nil {
		logger.Log.Fatalf("初始化Redis失败: %v", err)
	}

	// 初始化 Casbin 权限模型并同步策略。
	// 鉴权属于安全控制，初始化失败时拒绝启动，避免在"无鉴权"状态下对外提供服务。
	if err := middleware.InitCasbin(config.Cfg.Casbin.ModelPath); err != nil {
		logger.Log.Fatalf("初始化Casbin失败: %v", err)
	}

	// 扩展名白名单来自配置（未配置则沿用内置默认值）
	upload.SetAllowedExts(config.Cfg.Upload.AllowExts)
	upload.Init(service.LoadOSSConfig())

	r := router.Setup(config.Cfg.Server.Mode)

	// 注册日志清理定时任务：每天 03:00 清理超过保留期的操作日志与登录日志，
	// 避免日志表无限增长（保留天数见 log.db_retention_days，<=0 表示不清理）。
	// 多实例部署时每个实例都会执行，删除操作幂等，影响仅为重复执行。
	// 把定时任务的 panic 处理接到项目日志上（pkg/task 不依赖 internal，故由这里注入）
	task.SetPanicHandler(func(spec string, r interface{}, stack []byte) {
		logger.Log.Errorf("[cron] 任务 %s panic 已捕获，进程继续运行: %v\n%s", spec, r, stack)
	})

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
