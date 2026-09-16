// Package testsupport 提供仅用于测试的辅助设施（内存数据库等）。
//
// 只被 *_test.go 引用，不会进入生产二进制。
//
// 为什么需要它：仓储层直接使用全局 database.DB，而租户隔离、软删除
// 这类最容易回归的行为恰恰在仓储层。此前 system / member 模块零测试，
// 主要障碍就是「跑测试需要真实 MySQL」。
// 这里用纯 Go 的 SQLite 驱动（glebarez/sqlite）替代 —— 不需要 cgo，
// 不依赖本机装了什么，也不碰开发库的数据。
package testsupport

import (
	"fmt"
	"strings"
	"testing"

	"go-admin/internal/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDB 创建独立的内存数据库并注入 database.DB，返回句柄。
//
// models 会依次 AutoMigrate；传空表示只建库不建表
// （用于只关心生成 SQL 的场景，例如租户过滤断言）。
//
// ⚠️ database.DB 是包级变量，本函数会临时改写它，
// 因此**使用它的测试不能并行**（不要调用 t.Parallel）。
// 每个测试用独立的库名，测试之间互不污染。
func NewDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_", "\\", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层连接失败: %v", err)
	}
	// 内存库必须限制为单连接：多连接会各自看到不同的库
	sqlDB.SetMaxOpenConns(1)

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("建表失败: %v", err)
		}
	}

	prev := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = prev
		_ = sqlDB.Close()
	})

	return db
}

// DryRun 返回一个只构建 SQL、不真正执行的会话。
// 用于断言「查询是否带上了租户条件」而不必造数据。
func DryRun(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{DryRun: true})
}
