package common

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestFreedUniqueValue 软删除时释放唯一索引占用的值。
//
// 背景：唯一索引不区分记录是否已软删除。删除时不改写唯一字段，
// 被删记录会一直占用该值，导致同名记录再也建不出来（Duplicate entry）。
// 这是本项目真实踩过的坑（删掉角色 viewer 后无法用同一 code 重建）。
func TestFreedUniqueValue(t *testing.T) {
	t.Run("追加主键后缀", func(t *testing.T) {
		got := FreedUniqueValue("admin", 5, 64)
		if got != "admin_del_5" {
			t.Errorf("got %q, want %q", got, "admin_del_5")
		}
	})

	t.Run("不同主键产生不同结果", func(t *testing.T) {
		a := FreedUniqueValue("admin", 1, 64)
		b := FreedUniqueValue("admin", 2, 64)
		if a == b {
			t.Errorf("不同 id 应得到不同结果，均得到 %q", a)
		}
	})

	t.Run("按字符数截断，不破坏多字节字符", func(t *testing.T) {
		// 后缀 "_del_123" 共 8 字符，上限 20 → 原值最多保留 12 个字符
		name := strings.Repeat("测", 30) // 每个汉字 3 字节
		got := FreedUniqueValue(name, 123, 20)

		if utf8.RuneCountInString(got) > 20 {
			t.Errorf("结果应不超过 20 个字符，实际 %d 个: %q",
				utf8.RuneCountInString(got), got)
		}
		if !utf8.ValidString(got) {
			t.Errorf("截断破坏了 UTF-8 编码: %q", got)
		}
		if !strings.HasSuffix(got, "_del_123") {
			t.Errorf("后缀应保留，实际 %q", got)
		}
		if strings.ContainsRune(got, '\uFFFD') {
			t.Errorf("出现了替换字符，说明按字节截断了: %q", got)
		}
	})

	t.Run("列太短时只保留后缀也不 panic", func(t *testing.T) {
		got := FreedUniqueValue("verylongname", 9, 5)
		if got != "_del_9" {
			t.Errorf("上限不足以容纳后缀时应只保留后缀，实际 %q", got)
		}
	})

	t.Run("空原值只留后缀", func(t *testing.T) {
		if got := FreedUniqueValue("", 7, 64); got != "_del_7" {
			t.Errorf("got %q, want %q", got, "_del_7")
		}
	})
}
