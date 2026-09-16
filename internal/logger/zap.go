package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go-admin/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 全局日志器。
//
// 初值设为 no-op 而不是 nil：这样它**永远不为 nil**，所有调用点都可以直接写
// logger.Log.Infof(...)，不必逐个加空值判断。
//
// 早前声明为 `var Log *zap.SugaredLogger`（nil），后果是任何在 Init() 之前
// 触发的日志都会空指针 panic —— 而最需要日志的恰恰是启动失败与错误处理路径；
// 单元测试不调用 Init()，因此模块内一旦用 logger.Log 就会把测试打成 panic。
// 这正是 common.FailWith 里那句 `if logger.Log != nil` 保护存在的原因。
var Log = zap.NewNop().Sugar()

func Init() error {
	cfg := config.Cfg.Log

	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
	}

	if cfg.Filename != "" {
		// 日志目录**必须显式创建**：os.OpenFile 不会自动建父目录，
		// 而早前这里把失败静默跳过了（`if err == nil`），
		// 结果是「配置里写着写文件，实际只输出到 stdout」，运维毫不知情。
		if dir := filepath.Dir(cfg.Filename); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("创建日志目录失败 %s: %w", dir, err)
			}
		}

		// 启动时探一次可写性，避免「目录在但没权限」这类问题被拖到运行期才发现
		probe, err := os.OpenFile(cfg.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("日志文件不可写 %s: %w", cfg.Filename, err)
		}
		_ = probe.Close()

		// 按大小/份数/天数轮转，配置项与 lumberjack 字段一一对应。
		// 早前这几个配置项虽已声明却从未被读取，日志实际会无限增长。
		writer := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,    // 单个文件上限（MB）
			MaxBackups: cfg.MaxBackups, // 保留的历史文件数
			MaxAge:     cfg.MaxAge,     // 保留天数
			Compress:   cfg.Compress,   // 是否压缩历史文件
			LocalTime:  true,           // 文件名用本地时间，便于人工排查
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(writer), level))
	}

	core := zapcore.NewTee(cores...)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Log = l.Sugar()

	return nil
}
