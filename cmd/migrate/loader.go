package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// migration 一个待应用的迁移文件。
type migration struct {
	// version 取自文件名（去掉 .sql），既是排序依据也是幂等标识。
	// 约定形如 2026-09-16-post-tenant —— 日期前缀让字典序等于时间序。
	version string
	path    string
}

// loadMigrations 读取目录下的 .sql 文件，按版本升序返回。
//
// 为什么必须保证顺序：迁移之间存在依赖（先建表、后加列、再建索引），
// 顺序错了会在中途失败，而 DDL 无法回滚。用「日期前缀 + 字典序」
// 是实现稳定排序最简单可靠的方式，因此文件名规范是硬约定 ——
// 不合规的文件直接报错，而不是"尽力猜一个顺序"。
func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// 目录不存在时**不能**返回空列表：那会把「迁移目录被误删或路径配错」
		// 表现成「没有待执行的迁移」，从而安静地跳过整个升级过程。
		return nil, fmt.Errorf("读取迁移目录 %s 失败: %w", dir, err)
	}

	var out []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			// 非 .sql 一律忽略：目录里可能放 README、.bak 备份等
			continue
		}

		version := strings.TrimSuffix(e.Name(), ".sql")
		if err := validateVersion(version); err != nil {
			return nil, fmt.Errorf("迁移文件 %s 命名不合规: %w", e.Name(), err)
		}

		out = append(out, migration{
			version: version,
			path:    filepath.Join(dir, e.Name()),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// validateVersion 校验版本号以合法的 YYYY-MM-DD 开头。
//
// 只要求日期前缀，不强制完整的「日期-描述」格式 ——
// 目的是保证排序正确，而不是限制命名风格。
func validateVersion(v string) error {
	if len(v) < 10 {
		return fmt.Errorf("应以 YYYY-MM-DD 开头（如 2026-09-16-post-tenant）")
	}
	prefix := v[:10]
	if _, err := time.Parse("2006-01-02", prefix); err != nil {
		return fmt.Errorf("日期前缀 %q 不是合法的 YYYY-MM-DD", prefix)
	}
	return nil
}
