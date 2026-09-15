package service

import (
	"fmt"
	"testing"
)

// TestYuanToFen 校验「元 → 分」的转换不使用浮点。
//
// 用 int64(amt*100) 会因浮点表示误差少 1 分（19.99 → 1998），
// 而回调会用该值与订单金额比对，少 1 分即判为金额不匹配，
// 结果是用户已付款但订单不入账。
func TestYuanToFen(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"0.01", 1},
		{"19.99", 1999}, // 浮点法会算成 1998
		{"0.29", 29},    // 浮点法会算成 28
		{"1.10", 110},
		{"100", 10000},
		{"8", 800},
		{"0.1", 10},
		{"12.05", 1205},
		{"1.999", 199}, // 超出分的部分直接截断
		{" 7.50 ", 750},
		{"", 0},
		{"abc", 0},
		{"-3.50", -350},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := yuanToFen(c.in); got != c.want {
				t.Errorf("yuanToFen(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

// TestYuanToFenNoFloatError 逐一对齐浮点写法会出错的金额，
// 覆盖 0.01~99.99 全量，确保没有任何一分钱被吞掉。
func TestYuanToFenNoFloatError(t *testing.T) {
	for i := 1; i < 10000; i++ {
		s := fmtAmount(i)
		want := int64(i)

		var f float64
		_, _ = fmt.Sscanf(s, "%f", &f)
		if floatFen := int64(f * 100); floatFen != want {
			// 这个金额正是浮点法会算错的用例，验证新实现算对了
			if got := yuanToFen(s); got != want {
				t.Errorf("yuanToFen(%q) = %d, want %d（浮点法得 %d）", s, got, want, floatFen)
			}
		} else if got := yuanToFen(s); got != want {
			t.Errorf("yuanToFen(%q) = %d, want %d", s, got, want)
		}
	}
}

func fmtAmount(fen int) string {
	return fmtInt(fen/100) + "." + pad2(fen%100)
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + fmtInt(n)
	}
	return fmtInt(n)
}
