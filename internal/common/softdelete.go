package common

import "fmt"

// FreedUniqueValue 生成软删除时用于释放唯一索引占用的值。
//
// 背景：唯一索引不区分记录是否已软删除。若删除时不改写唯一字段，
// 被删除的记录会一直占用该值，导致同名记录无法再次创建（Duplicate entry）。
//
// 做法：追加主键 id 作为后缀（id 唯一，故结果唯一），并按列长度上限
// 截断原值（按字符计，避免截断多字节字符）。
// 后缀保持紧凑（形如 _del_123），以免在 varchar(20) 这类短列上
// 把原值截断过多。
func FreedUniqueValue(value string, id uint, maxLen int) string {
	suffix := fmt.Sprintf("_del_%d", id)

	limit := maxLen - len([]rune(suffix))
	if limit < 0 {
		limit = 0
	}

	runes := []rune(value)
	if len(runes) > limit {
		runes = runes[:limit]
	}

	return string(runes) + suffix
}
