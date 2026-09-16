package common

import (
	"database/sql"
	"strings"
	"testing"

	"go-admin/internal/testsupport"

	_ "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// tenantScopedThing 仅用于断言 SQL 生成，不映射真实表
type tenantScopedThing struct {
	ID       uint
	TenantID uint
	Name     string
}

func (tenantScopedThing) TableName() string { return "t_tenant_thing" }

// newDryRunDB 构造一个不会真正连接数据库的 GORM 实例，用于断言生成的 SQL。
//
// 三处设置缺一不可，否则测试会真的去连库：
//   - sql.Open 是**惰性**的：只校验 DSN 格式，不建立连接
//   - SkipInitializeWithVersion —— 跳过"查询 MySQL 版本"，否则 gorm.Open 立刻连库
//   - DisableAutomaticPing —— **gorm.Open 默认会 Ping 一次**，只加前两项仍会连库
//     （实测：漏掉它时报 "Error 1045 Access denied for user 'user'@'localhost'"，
//      说明确实发起了真实连接）
//
// DryRun 模式下 GORM 只构建 SQL 不执行，因此可纯离线验证条件拼接。
func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("mysql", "user:pass@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		t.Fatalf("sql.Open 失败: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(
		gormmysql.New(gormmysql.Config{
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true,
		}),
		&gorm.Config{
			DryRun:               true,
			DisableAutomaticPing: true,
		},
	)
	if err != nil {
		t.Fatalf("gorm.Open 失败: %v", err)
	}
	return db
}

// buildScopedSQL 返回 TenantScope 作用后生成的 SQL
func buildScopedSQL(t *testing.T, tenantID uint) string {
	t.Helper()

	db := newDryRunDB(t)
	var out []tenantScopedThing
	tx := TenantScope(db.Model(&tenantScopedThing{}), tenantID).Find(&out)
	if tx.Error != nil {
		t.Fatalf("构建查询失败: %v", tx.Error)
	}
	return tx.Statement.SQL.String()
}

// TestTenantScopeAddsFilterForRealTenant 租户 ID 非 0 时必须附加 tenant_id 条件。
func TestTenantScopeAddsFilterForRealTenant(t *testing.T) {
	sqlStr := buildScopedSQL(t, 42)
	if !strings.Contains(sqlStr, "tenant_id") {
		t.Fatalf("tenantID=42 应生成 tenant_id 过滤条件，实际 SQL: %s", sqlStr)
	}
}

// TestTenantScopeSkippedForZeroTenant 把「tenantID=0 表示不过滤」这一语义
// **固定在测试里**。
//
// 这是本项目最危险的失效模式：租户表的 Repository 方法一旦漏传 tenantID，
// TenantScope 不会报错，而是静默退化成全表查询 —— 接口照常返回 200，
// 数据却跨租户泄漏了，从响应上看不出任何异常。
//
// 因此"0 = 平台级、不过滤"这个行为必须有测试兜住：
//   - 若有人把 TenantScope 改成"0 也过滤"，平台级账号会突然查不到任何数据，
//     这个用例会立刻失败并指出原因；
//   - 若有人改成"0 报错/panic"，同样会在这里暴露，而不是等到线上。
//
// 另需知悉：真正防止漏传的手段是**让 tenantID 贯穿 Controller → Service →
// Repository 的签名**（见 AGENTS.md 规则 7），本用例只是最后的语义护栏。
func TestTenantScopeSkippedForZeroTenant(t *testing.T) {
	sqlStr := buildScopedSQL(t, 0)
	if strings.Contains(sqlStr, "tenant_id") {
		t.Fatalf("tenantID=0 表示平台级、不应过滤，但生成的 SQL 含 tenant_id 条件: %s", sqlStr)
	}
}

// TestTenantScopeIsNotMutatingSourceDB 确认 TenantScope 不会污染传入的 db 实例。
//
// GORM 的链式调用会共享同一个 Statement，若实现写成原地修改，
// 同一个 db 上后续的无租户查询会被"传染"上租户条件，
// 导致平台级查询莫名其妙查不到数据。GORM 的 Where 返回新实例，
// 这里把这个假设固定下来。
func TestTenantScopeIsNotMutatingSourceDB(t *testing.T) {
	db := newDryRunDB(t)
	base := db.Model(&tenantScopedThing{})

	var scoped []tenantScopedThing
	scopedTx := TenantScope(base, 7).Find(&scoped)
	if scopedTx.Error != nil {
		t.Fatalf("构建带租户查询失败: %v", scopedTx.Error)
	}
	if !strings.Contains(scopedTx.Statement.SQL.String(), "tenant_id") {
		t.Fatalf("前置条件不成立：带租户查询未生成 tenant_id 条件")
	}

	// 复用同一个 base 再查一次，不应带 tenant_id
	var plain []tenantScopedThing
	plainTx := base.Session(&gorm.Session{NewDB: true}).Find(&plain)
	if plainTx.Error != nil {
		t.Fatalf("构建无租户查询失败: %v", plainTx.Error)
	}
	if strings.Contains(plainTx.Statement.SQL.String(), "tenant_id") {
		t.Errorf("TenantScope 污染了源 db 实例，后续查询被附加了 tenant_id 条件: %s",
			plainTx.Statement.SQL.String())
	}
}

// tenantProbe 用于端到端行隔离测试，会真正建表并写入数据
type tenantProbe struct {
	ID       uint `gorm:"primarykey"`
	TenantID uint `gorm:"index"`
	Name     string
}

func (tenantProbe) TableName() string { return "tenant_probe" }

// TestTenantScopeIsolatesRows 端到端验证：不同租户的数据真的互相看不见。
//
// 上面三个用例都只断言「SQL 里有没有 tenant_id」，DryRun 不执行查询，
// 因此它们无法发现「条件拼对了但没生效」这类问题。本用例用内存库真正建表、
// 写入三个租户的数据，再按租户查询，断言返回的行集合。
//
// 这是整套测试里最重要的一条 —— 租户隔离失效属于「静默越权」，
// 接口照常返回 200，只是把别的租户数据也带出来了，靠人工点界面很难发现。
func TestTenantScopeIsolatesRows(t *testing.T) {
	db := testsupport.NewDB(t, &tenantProbe{})

	seed := []tenantProbe{
		{TenantID: 1, Name: "t1-a"},
		{TenantID: 1, Name: "t1-b"},
		{TenantID: 2, Name: "t2-a"},
		{TenantID: 0, Name: "platform"},
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("准备数据失败: %v", err)
	}

	t.Run("租户1 只能看到自己的数据", func(t *testing.T) {
		var got []tenantProbe
		if err := TenantScope(db, 1).Find(&got).Error; err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("应返回 2 条，实际 %d 条: %+v", len(got), got)
		}
		for _, r := range got {
			if r.TenantID != 1 {
				t.Errorf("泄漏了其他租户的数据: %+v", r)
			}
		}
	})

	t.Run("租户2 只能看到自己的数据", func(t *testing.T) {
		var got []tenantProbe
		if err := TenantScope(db, 2).Find(&got).Error; err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if len(got) != 1 || got[0].Name != "t2-a" {
			t.Errorf("应只返回 t2-a，实际 %+v", got)
		}
	})

	t.Run("tenantID=0 为平台级查询，可看到全部", func(t *testing.T) {
		var got []tenantProbe
		if err := TenantScope(db, 0).Find(&got).Error; err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if len(got) != 4 {
			t.Errorf("应返回全部 4 条，实际 %d", len(got))
		}
	})
}
