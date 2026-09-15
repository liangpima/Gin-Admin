package middleware

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestSanitizeRequestBodyMasksPasswords 验证请求体中的密码等敏感字段被脱敏。
//
// 这是安全要求：操作日志曾把创建用户 / 重置密码 / 修改密码接口的明文密码写入数据库。
func TestSanitizeRequestBodyMasksPasswords(t *testing.T) {
	body := []byte(`{"username":"alice","password":"Secret123","profile":{"oldPassword":"Old123","nickname":"A"}}`)

	got := sanitizeRequestBody(body)

	if strings.Contains(got, "Secret123") || strings.Contains(got, "Old123") {
		t.Fatalf("敏感值未被脱敏: %s", got)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("脱敏结果不是合法 JSON: %v (结果=%s)", err, got)
	}
	if parsed["password"] != maskedValue {
		t.Errorf("password 应被替换为 %q，实际 %v", maskedValue, parsed["password"])
	}
	if parsed["username"] != "alice" {
		t.Errorf("非敏感字段 username 不应被改动，实际 %v", parsed["username"])
	}

	profile, ok := parsed["profile"].(map[string]interface{})
	if !ok {
		t.Fatalf("profile 结构异常: %v", parsed["profile"])
	}
	if profile["oldPassword"] != maskedValue {
		t.Errorf("嵌套的 oldPassword 应被脱敏，实际 %v", profile["oldPassword"])
	}
	if profile["nickname"] != "A" {
		t.Errorf("嵌套的非敏感字段不应被改动，实际 %v", profile["nickname"])
	}
}

// TestSanitizeRequestBodyCaseInsensitive 验证字段名匹配不区分大小写
func TestSanitizeRequestBodyCaseInsensitive(t *testing.T) {
	body := []byte(`{"Password":"X1","NEWPASSWORD":"X2","AccessKey":"X3","token":"X4","privateKey":"X5"}`)

	got := sanitizeRequestBody(body)

	for _, secret := range []string{"X1", "X2", "X3", "X4", "X5"} {
		if strings.Contains(got, secret) {
			t.Errorf("敏感值 %s 未被脱敏: %s", secret, got)
		}
	}
}

// TestSanitizeRequestBodyArray 验证数组元素中的敏感字段也会被处理
func TestSanitizeRequestBodyArray(t *testing.T) {
	body := []byte(`{"items":[{"password":"P1"},{"nested":{"secret":"P2"}}]}`)

	got := sanitizeRequestBody(body)

	if strings.Contains(got, "P1") || strings.Contains(got, "P2") {
		t.Errorf("数组内的敏感值未被脱敏: %s", got)
	}
}

// TestSanitizeRequestBodyNonJSON 验证非 JSON 请求体不记录内容
func TestSanitizeRequestBodyNonJSON(t *testing.T) {
	body := []byte("------WebKitFormBoundary\r\nContent-Disposition: form-data; name=\"file\"\r\n\r\nsecret-content")

	got := sanitizeRequestBody(body)

	if strings.Contains(got, "secret-content") {
		t.Errorf("非 JSON 请求体不应记录内容，实际: %s", got)
	}
}

// TestSanitizeRequestBodyEmpty 验证空请求体返回空串
func TestSanitizeRequestBodyEmpty(t *testing.T) {
	if got := sanitizeRequestBody(nil); got != "" {
		t.Errorf("nil 请求体应返回空串，实际 %q", got)
	}
	if got := sanitizeRequestBody([]byte{}); got != "" {
		t.Errorf("空请求体应返回空串，实际 %q", got)
	}
}

// TestSanitizeRequestBodyTruncates 验证超长请求体被截断
func TestSanitizeRequestBodyTruncates(t *testing.T) {
	long := strings.Repeat("a", maxBodyLogLength*2)
	body := []byte(`{"nickname":"` + long + `"}`)

	got := sanitizeRequestBody(body)

	if !strings.HasSuffix(got, "...[truncated]") {
		t.Errorf("超长请求体应带截断标记，实际结尾: %q", tail(got, 30))
	}
	if n := len([]rune(got)); n > maxBodyLogLength+len("...[truncated]") {
		t.Errorf("截断后长度 %d 超出上限 %d", n, maxBodyLogLength)
	}
}

// TestSanitizeRequestBodyTruncateRuneSafe 验证截断不会切断多字节字符
func TestSanitizeRequestBodyTruncateRuneSafe(t *testing.T) {
	long := strings.Repeat("中", maxBodyLogLength)
	body := []byte(`{"nickname":"` + long + `"}`)

	got := sanitizeRequestBody(body)

	if !utf8.ValidString(got) {
		t.Error("截断产生了非法 UTF-8，说明切断了多字节字符")
	}
}

// TestSanitizeRequestBodyKeepsShortBodyIntact 验证短请求体不被改动结构
func TestSanitizeRequestBodyKeepsShortBodyIntact(t *testing.T) {
	body := []byte(`{"a":1,"b":"x"}`)

	got := sanitizeRequestBody(body)

	if strings.Contains(got, "truncated") {
		t.Errorf("短请求体不应被截断: %s", got)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("结果应为合法 JSON: %v", err)
	}
	if parsed["b"] != "x" {
		t.Errorf("普通字段应原样保留，实际 %v", parsed["b"])
	}
}

func tail(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}
