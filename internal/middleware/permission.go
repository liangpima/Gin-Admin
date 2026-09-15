package middleware

import (
	"sync"
)

// routePerms 路由权限表：key 为 "METHOD 完整路径模板"，value 为所需权限码。
//
// 空字符串表示该路由仅要求登录态（自助接口）。
// 未登记的路由会被默认拒绝，避免新增接口时漏配权限而被放行。
//
// 之所以用「登记表 + 组级中间件查表」而不是「每路由挂一个设置 context 的中间件」：
// gin 的组中间件在路由处理器之前执行，若把权限码写在路由级中间件里，
// 组级鉴权中间件运行时该值尚未写入，会导致校验被整体跳过。
var (
	routePermsMu sync.RWMutex
	routePerms   = make(map[string]string)
)

// RegisterPermission 登记路由所需权限码，在路由注册阶段调用。
// code 为空表示仅要求登录态。
func RegisterPermission(method, fullPath, code string) {
	routePermsMu.Lock()
	defer routePermsMu.Unlock()
	routePerms[method+" "+fullPath] = code
}

// routePermission 查询路由所需权限码；ok 为 false 表示该路由未登记
func routePermission(method, fullPath string) (code string, ok bool) {
	routePermsMu.RLock()
	defer routePermsMu.RUnlock()
	code, ok = routePerms[method+" "+fullPath]
	return code, ok
}
