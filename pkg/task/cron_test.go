package task

import (
	"strings"
	"testing"
)

// TestWrapWithRecoverCatchesPanic 验证任务 panic 不会外溢。
//
// 这是关键回归测试：robfig/cron 默认不 recover，
// 任务 panic 会直接终止整个进程（Recovery 中间件只管 HTTP，救不了后台任务）。
// 若这里失败，说明恢复能力被破坏，凌晨的清理任务一旦 panic 会导致服务无人值守时挂掉。
func TestWrapWithRecoverCatchesPanic(t *testing.T) {
	var gotSpec string
	var gotPanic interface{}
	var gotStack []byte

	orig := PanicHandler
	defer func() { PanicHandler = orig }()

	PanicHandler = func(spec string, r interface{}, stack []byte) {
		gotSpec, gotPanic, gotStack = spec, r, stack
	}

	wrapped := wrapWithRecover("0 3 * * *", func() {
		panic("模拟任务内部 panic")
	})

	// 关键断言：调用过程本身不能 panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic 未被拦截，直接外溢了: %v", r)
		}
	}()
	wrapped()

	if gotSpec != "0 3 * * *" {
		t.Errorf("PanicHandler 收到的 spec 不正确: %q", gotSpec)
	}
	if gotPanic != "模拟任务内部 panic" {
		t.Errorf("PanicHandler 收到的 panic 值不正确: %v", gotPanic)
	}
	if len(gotStack) == 0 || !strings.Contains(string(gotStack), "goroutine") {
		t.Error("应记录堆栈信息以便定位，实际为空或不含 goroutine")
	}
}

// TestWrapWithRecoverPassesThroughNormalRun 确保正常任务不受包装影响
func TestWrapWithRecoverPassesThroughNormalRun(t *testing.T) {
	called := false
	PanicHandlerCalled := false

	orig := PanicHandler
	defer func() { PanicHandler = orig }()
	PanicHandler = func(string, interface{}, []byte) { PanicHandlerCalled = true }

	wrapWithRecover("test", func() { called = true })()

	if !called {
		t.Error("任务体未被执行")
	}
	if PanicHandlerCalled {
		t.Error("正常执行不应触发 PanicHandler")
	}
}

// TestDefaultPanicHandlerNotNil 默认实现必须存在，避免未注入时静默吞掉 panic
func TestDefaultPanicHandlerNotNil(t *testing.T) {
	if PanicHandler == nil {
		t.Fatal("PanicHandler 默认实现不应为 nil")
	}
}
