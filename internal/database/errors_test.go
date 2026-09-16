package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

// TestIsDuplicateKey 回归保护：唯一约束冲突的识别。
//
// 识别错/识别不到都是静默失效 —— 本该返回 400「XX已存在」的提示会退回
// 500「服务器内部错误」，用户完全不知道是自己重名了，运维也只看到一堆 500。
// 因此这里既验证正例，也验证**不能误判**的反例（连接错误、其他错误码）。
func TestIsDuplicateKey(t *testing.T) {
	dup := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'admin' for key 'uk_username'"}

	t.Run("识别裸的 1062", func(t *testing.T) {
		if !IsDuplicateKey(dup) {
			t.Error("1062 应被识别为唯一冲突")
		}
	})

	t.Run("穿透包装后的 1062", func(t *testing.T) {
		// Repository 包装时必须用 `%w: %w`（两个 %w）：第一个保留
		// common.ErrDuplicateKey 供 Service 用 errors.Is 判定，第二个保留驱动
		// 原始错误，使 IsDuplicateKey 对已包装的错误依然可用。
		// 若把第二个写成 %v，原始错误会被字符串化，errors.As 从此找不到
		// *mysql.MySQLError —— 本用例守护的就是这一点。
		wrapped := fmt.Errorf("%w: %w", errors.New("duplicate key"), dup)
		if !IsDuplicateKey(wrapped) {
			t.Error("应能穿透包装识别出 1062")
		}
	})

	t.Run("其他 MySQL 错误码不误判", func(t *testing.T) {
		for _, code := range []uint16{1064, 1146, 1045, 1213, 0} {
			if IsDuplicateKey(&mysql.MySQLError{Number: code, Message: "x"}) {
				t.Errorf("错误码 %d 不应被识别为唯一冲突", code)
			}
		}
	})

	t.Run("非 MySQL 错误不误判", func(t *testing.T) {
		if IsDuplicateKey(errors.New("Duplicate entry 'admin' for key 'uk_username'")) {
			t.Error("纯文本错误不应被识别（这正是选用错误码而非字符串匹配的原因）")
		}
		if IsDuplicateKey(fmt.Errorf("dial tcp 127.0.0.1:3306: connect: connection refused")) {
			t.Error("连接错误不应被识别为唯一冲突")
		}
	})

	t.Run("nil 不 panic 且返回 false", func(t *testing.T) {
		if IsDuplicateKey(nil) {
			t.Error("nil 应返回 false")
		}
	})
}
