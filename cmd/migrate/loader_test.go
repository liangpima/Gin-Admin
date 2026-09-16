package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMigration(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}
}

// TestLoadMigrationsSortsByVersion 按版本升序返回，与文件在磁盘上的顺序无关。
//
// 这是执行器的正确性基础：迁移之间有依赖（先建表、后加列），
// 顺序错了会在中途失败，而 DDL 无法回滚。
func TestLoadMigrationsSortsByVersion(t *testing.T) {
	dir := t.TempDir()
	// 刻意乱序写入
	writeMigration(t, dir, "2026-09-16-post-tenant.sql", "SELECT 1;")
	writeMigration(t, dir, "2026-09-15-schema.sql", "SELECT 1;")
	writeMigration(t, dir, "2026-09-15-settings-menu-permission.sql", "SELECT 1;")

	got, err := loadMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("应加载 3 个迁移，got %d", len(got))
	}

	want := []string{
		"2026-09-15-schema",
		"2026-09-15-settings-menu-permission",
		"2026-09-16-post-tenant",
	}
	for i, w := range want {
		if got[i].version != w {
			t.Errorf("第 %d 个版本应为 %s，got %s", i, w, got[i].version)
		}
	}
}

// TestLoadMigrationsIgnoresNonSQL 非 .sql 文件必须被忽略。
// 迁移目录里常会放 README、.bak 备份，误当迁移执行会出事故。
func TestLoadMigrationsIgnoresNonSQL(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, dir, "2026-09-15-schema.sql", "SELECT 1;")
	writeMigration(t, dir, "README.md", "# 迁移说明")
	writeMigration(t, dir, "2026-09-15-schema.sql.bak", "旧版本内容")

	got, err := loadMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("应只加载 1 个迁移，got %d", len(got))
	}
	if got[0].version != "2026-09-15-schema" {
		t.Errorf("版本应为 2026-09-15-schema，got %s", got[0].version)
	}
}

// TestLoadMigrationsRejectsBadName 命名不合规必须报错，而不是"尽力猜顺序"。
//
// 版本号同时承担排序与幂等标识两个职责：格式错了会安静地跑乱顺序，
// 而这类错误在 DDL 无法回滚的前提下代价很高。宁可在启动阶段拒绝。
func TestLoadMigrationsRejectsBadName(t *testing.T) {
	cases := []string{
		"schema.sql",         // 无日期前缀
		"2026-13-45-bad.sql", // 月份日期越界
		"abcd-ef-gh-x.sql",   // 前缀不是日期
		"x.sql",              // 太短
	}

	for _, file := range cases {
		t.Run(file, func(t *testing.T) {
			dir := t.TempDir()
			writeMigration(t, dir, file, "SELECT 1;")

			if _, err := loadMigrations(dir); err == nil {
				t.Errorf("命名 %s 应被拒绝", file)
			}
		})
	}
}

// TestLoadMigrationsMissingDir 目录不存在时必须报错。
//
// 若返回空列表，「迁移目录被误删 / -config 配错导致工作目录不对」
// 就会表现成「没有待执行的迁移」，从而安静地跳过整个升级过程 ——
// 这正是加执行器要解决的问题，不能自己再制造一个。
func TestLoadMigrationsMissingDir(t *testing.T) {
	if _, err := loadMigrations(filepath.Join(t.TempDir(), "not-exist")); err == nil {
		t.Fatal("目录不存在时应返回错误")
	}
}

func TestLoadMigrationsEmptyDir(t *testing.T) {
	got, err := loadMigrations(t.TempDir())
	if err != nil {
		t.Fatalf("空目录不应报错: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("空目录应返回空列表，got %d", len(got))
	}
}

// TestValidateVersionAcceptsRealNames 用仓库里真实的迁移文件名做回归，
// 避免「规则改了但既有文件反而不合规」这种自伤。
func TestValidateVersionAcceptsRealNames(t *testing.T) {
	names := []string{
		"2026-09-15-schema",
		"2026-09-15-settings-menu-permission",
		"2026-09-16-post-tenant",
		"2026-09-16", // 只有日期也应接受：规则只约束前缀
	}
	for _, n := range names {
		if err := validateVersion(n); err != nil {
			t.Errorf("%s 应被接受，却报错: %v", n, err)
		}
	}
}
