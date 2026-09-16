package database

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysqlDuplicateEntry 唯一约束冲突的错误码（ER_DUP_ENTRY）
const mysqlDuplicateEntry = 1062

// IsDuplicateKey 判断是否为唯一约束冲突。
//
// 这里用错误码类型断言，而不是匹配 "Duplicate entry" 字符串：
// 原始错误文本会随 MySQL 版本、sql_mode、甚至驱动版本变化，
// 字符串匹配一旦失效是**静默**的 —— 本该返回 400 的友好提示会退回 500，
// 而且没有任何迹象表明匹配逻辑坏了。
//
// 注意：gorm.Config 未开启 TranslateError，所以 GORM 不会把它转成
// gorm.ErrDuplicatedKey，拿到的是驱动原始的 *mysql.MySQLError。
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return myErr.Number == mysqlDuplicateEntry
	}
	return false
}
