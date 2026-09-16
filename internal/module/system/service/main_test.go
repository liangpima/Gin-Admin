package service

import (
	"fmt"
	"os"
	"testing"

	"go-admin/config"
	"go-admin/internal/logger"
)

// TestMain 初始化日志。
//
// service 层在错误路径上会写日志（`logger.Log.Warnf` 等），
// 而 `logger.Log` 只有在 logger.Init() 之后才非 nil ——
// 不初始化的话，任何走到错误分支的用例都会以 nil 解引用 panic 告终，
// 等于错误路径完全无法测试。
//
// 把级别压到 error，避免测试输出被 Info/Warn 淹没。
func TestMain(m *testing.M) {
	config.Cfg.Log.Level = "error"
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
