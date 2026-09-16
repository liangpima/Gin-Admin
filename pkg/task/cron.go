package task

import (
	"log"
	"runtime/debug"
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	c   *cron.Cron
	once sync.Once
)

// PanicHandler 任务 panic 时的回调，由调用方注入（见 main.go 注入 zap 日志）。
//
// 之所以做成可注入而不是直接 import internal/logger：
// pkg/ 层保持「不依赖 internal/」的约定，依赖方向才不会被倒置。
// 默认实现写标准日志，保证未注入时也不会静默丢失。
//
// 必须在 Start() 之前设置，运行期不再变更，因此无需加锁。
var PanicHandler = func(spec string, r interface{}, stack []byte) {
	log.Printf("[cron] 任务 %s panic: %v\n%s", spec, r, stack)
}

// SetPanicHandler 注入 panic 处理函数，应在 Start() 之前调用
func SetPanicHandler(h func(spec string, r interface{}, stack []byte)) {
	if h != nil {
		PanicHandler = h
	}
}

func Init() {
	once.Do(func() {
		c = cron.New()
	})
}

// AddJob 注册定时任务。
//
// 任务体内统一包一层 recover：robfig/cron 会把每个任务放进独立 goroutine 执行，
// 而它**默认不 recover panic** —— 任务里一旦 panic，整个进程直接崩溃。
// middleware.Recovery 只覆盖 HTTP 请求，救不了后台任务，
// 因此凌晨执行的清理任务若 panic，服务会在无人值守时挂掉。
func AddJob(spec string, cmd func()) (cron.EntryID, error) {
	Init()
	return c.AddFunc(spec, wrapWithRecover(spec, cmd))
}

// wrapWithRecover 把任务体包成「panic 不会外溢」的函数。
// 抽成独立函数是为了可单测 —— cron 按时间触发，不方便直接驱动。
func wrapWithRecover(spec string, cmd func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				PanicHandler(spec, r, debug.Stack())
			}
		}()
		cmd()
	}
}

func Start() {
	Init()
	c.Start()
}

func Stop() {
	if c != nil {
		c.Stop()
	}
}
