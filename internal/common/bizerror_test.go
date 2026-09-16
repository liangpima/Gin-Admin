package common

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestBizErrorCode(t *testing.T) {
	if e, ok := AsBizError(NewBizError("用户名已存在")); !ok || e.Code != CodeBadRequest {
		t.Errorf("NewBizError 应为 400，实际 %+v ok=%v", e, ok)
	}
	if e, ok := AsBizError(NewNotFoundError("用户不存在")); !ok || e.Code != CodeNotFound {
		t.Errorf("NewNotFoundError 应为 404，实际 %+v ok=%v", e, ok)
	}
	if e, ok := AsBizError(NewBizErrorf("不支持的渠道: %s", "foo")); !ok || e.Code != CodeBadRequest || e.Msg != "不支持的渠道: foo" {
		t.Errorf("NewBizErrorf 不正确: %+v", e)
	}
}

// TestNotFoundOrErr 校验「记录不存在」被归一为 404，而系统错误原样透出。
//
// 这条是本次梳理的关键：GORM 查不到记录时返回 gorm.ErrRecordNotFound，
// 若不转换就会变成 500「服务器内部错误」，用户完全不知道是自己查的 ID 不对。
func TestNotFoundOrErr(t *testing.T) {
	if err := NotFoundOrErr(nil, "用户不存在"); err != nil {
		t.Errorf("nil 错误应原样返回 nil，实际 %v", err)
	}

	e, ok := AsBizError(NotFoundOrErr(gorm.ErrRecordNotFound, "用户不存在"))
	if !ok || e.Code != CodeNotFound || e.Msg != "用户不存在" {
		t.Errorf("ErrRecordNotFound 应转为 404 业务错误，实际 %+v ok=%v", e, ok)
	}

	dbErr := errors.New("invalid connection")
	if got := NotFoundOrErr(dbErr, "用户不存在"); got != dbErr {
		t.Errorf("系统错误应原样透出，实际 %v", got)
	}
	if IsBizError(dbErr) {
		t.Error("系统错误不应被判定为业务错误")
	}
}

// TestAsBizErrorUnwrap 校验错误被包装后仍能识别（service 常用 fmt.Errorf %w 包装）
func TestAsBizErrorUnwrap(t *testing.T) {
	wrapped := fmt.Errorf("保存失败: %w", NewBizError("用户名已存在"))
	if !IsBizError(wrapped) {
		t.Error("包装后的业务错误应仍能被识别")
	}
	e, _ := AsBizError(wrapped)
	if e.Code != CodeBadRequest {
		t.Errorf("包装后应保留业务码 400，实际 %d", e.Code)
	}
}

// TestFailWith 校验统一出口的状态码选择：
// 业务错误按其 Code 输出且原文案可见；系统错误只给通用文案，不泄漏细节。
func TestFailWith(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name    string
		err     error
		wantCode int
		wantMsg  string
	}{
		{"业务错误→400 且原文可见", NewBizError("用户名已存在"), CodeBadRequest, "用户名已存在"},
		{"资源不存在→404", NewNotFoundError("用户不存在"), CodeNotFound, "用户不存在"},
		{"系统错误→500 且不泄漏细节", errors.New("Error 1146: Table 'gin.sys_user' doesn't exist"), CodeInternalError, "服务器内部错误"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

			FailWith(ctx, c.err)

			if w.Code != http.StatusOK {
				t.Errorf("HTTP 状态码应为 200（业务码在 body 中），实际 %d", w.Code)
			}
			body := w.Body.String()
			if !contains(body, fmt.Sprintf(`"code":%d`, c.wantCode)) {
				t.Errorf("业务码应为 %d，实际响应 %s", c.wantCode, body)
			}
			if !contains(body, c.wantMsg) {
				t.Errorf("响应应包含 %q，实际 %s", c.wantMsg, body)
			}
			if c.wantCode == CodeInternalError && contains(body, "Table") {
				t.Errorf("系统错误泄漏了内部细节: %s", body)
			}
		})
	}
}

// TestFailWithNil 防御：nil 不应写出任何响应
func TestFailWithNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	FailWith(ctx, nil)

	if w.Body.Len() != 0 {
		t.Errorf("nil 错误不应产生响应，实际 %s", w.Body.String())
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
