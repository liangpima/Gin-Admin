package middleware

import (
	"sync"
	"testing"
)

// TestRoutePermissionDefaultDeny 未登记的路由必须返回 ok=false。
//
// 这是整个鉴权体系的「默认拒绝」基石：CasbinAuth 对未登记的 (method, path)
// 一律 403，而不是放行。一旦有人把这里改成「查不到就当作空权限码（仅登录）」，
// 所有漏配权限的新接口都会静默变成「登录即可访问」，且没有任何报错。
func TestRoutePermissionDefaultDeny(t *testing.T) {
	if _, ok := routePermission("GET", "/api/v1/definitely-not-registered"); ok {
		t.Fatal("未登记的路由必须返回 ok=false（默认拒绝）")
	}
}

func TestRoutePermissionRegistered(t *testing.T) {
	RegisterPermission("GET", "/api/v1/_test_resource", "test:resource:list")

	code, ok := routePermission("GET", "/api/v1/_test_resource")
	if !ok {
		t.Fatal("已登记路由应返回 ok=true")
	}
	if code != "test:resource:list" {
		t.Errorf("权限码应为 test:resource:list，got %q", code)
	}

	// method 是 key 的一部分：同一路径的不同方法互不影响，
	// 否则只登记 GET 会让同路径的 POST 也被视作已授权
	if _, ok := routePermission("POST", "/api/v1/_test_resource"); ok {
		t.Error("仅登记了 GET，POST 不应被视作已登记")
	}
}

// TestRoutePermissionEmptyCodeMeansLoginOnly 空权限码表示「仅要求登录态」，
// 必须与「未登记」区分开：前者放行（走 c.Next），后者 403。
// 两者在 map 里都「没有有效权限码」，靠 ok 区分，容易被改错。
func TestRoutePermissionEmptyCodeMeansLoginOnly(t *testing.T) {
	RegisterPermission("GET", "/api/v1/_test_self_service", "")

	code, ok := routePermission("GET", "/api/v1/_test_self_service")
	if !ok {
		t.Fatal("已登记（空权限码）的路由应返回 ok=true，与未登记区分开")
	}
	if code != "" {
		t.Errorf("权限码应为空字符串，got %q", code)
	}
}

func TestRegisterPermissionOverwrites(t *testing.T) {
	RegisterPermission("GET", "/api/v1/_test_overwrite", "old:code")
	RegisterPermission("GET", "/api/v1/_test_overwrite", "new:code")

	code, _ := routePermission("GET", "/api/v1/_test_overwrite")
	if code != "new:code" {
		t.Errorf("重复登记应覆盖旧值，got %q", code)
	}
}

// TestRoutePermissionConcurrent 登记发生在路由注册阶段，查询发生在每个请求上。
// 两者并发访问同一个 map —— 没有 RWMutex 会触发
// fatal error: concurrent map read and map write，
// 这类错误 Recovery 中间件抓不住，进程会直接退出。
//
// 配合 -race 运行时这条用例才有完整意义（普通运行只能碰运气撞出 panic）。
func TestRoutePermissionConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			RegisterPermission("GET", "/api/v1/_concurrent", "code:concurrent")
		}()
		go func() {
			defer wg.Done()
			routePermission("GET", "/api/v1/_concurrent")
		}()
	}
	wg.Wait()
}
