package utils

import (
	"sync"
	"testing"
)

// TestGenerateIDConcurrentUnique 并发生成 ID 必须唯一且无数据竞争。
//
// 用 `go test -race` 运行可检出懒初始化的数据竞争：
// 早前的 `if sf == nil { InitSnowflake(1) }` 在并发首调时会同时写全局变量。
func TestGenerateIDConcurrentUnique(t *testing.T) {
	InitSnowflake(1)

	const goroutines = 16
	const perGoroutine = 500

	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := make(map[int64]struct{}, goroutines*perGoroutine)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ids := make([]int64, 0, perGoroutine)
			for j := 0; j < perGoroutine; j++ {
				ids = append(ids, GenerateID())
			}
			mu.Lock()
			for _, id := range ids {
				if _, dup := seen[id]; dup {
					t.Errorf("生成了重复 ID: %d", id)
				}
				seen[id] = struct{}{}
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(seen) != goroutines*perGoroutine {
		t.Errorf("唯一 ID 数量为 %d，期望 %d", len(seen), goroutines*perGoroutine)
	}
}

// TestInitSnowflakeMachineIDBound 机器号越界时回退为 1，不 panic
func TestInitSnowflakeMachineIDBound(t *testing.T) {
	sf = nil
	sfOnce = sync.Once{}
	InitSnowflake(-1)
	if sf == nil || sf.machineID != 1 {
		t.Errorf("越界机器号应回退为 1，实际 %+v", sf)
	}
}
