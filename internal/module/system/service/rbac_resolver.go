package service

import (
	"go-admin/internal/middleware"
	"go-admin/internal/module/system/repository"
)

// RBACRoleResolver 实现 middleware.RoleResolver，把「角色 code 解析」这一
// 能力从中间件里搬回业务侧。
//
// 为什么适配器放在 system 模块而不是 middleware：
// 中间件只声明它需要什么（一个返回角色 code 列表的方法），
// 具体查哪些表、怎么过滤停用角色，属于系统模块的知识。
// 这样依赖方向保持为「业务模块 → 横切工具」，中间件不再反向依赖业务仓储。
//
// 依赖 middleware 仅用于编译期断言接口实现，不调用其任何函数。
type RBACRoleResolver struct{}

// 编译期确认实现满足中间件侧契约。
// 若接口签名变更，这里会直接编译失败，而不是等到运行期注入处才发现。
var _ middleware.RoleResolver = (*RBACRoleResolver)(nil)

func NewRBACRoleResolver() *RBACRoleResolver {
	return &RBACRoleResolver{}
}

// RoleCodesOf 返回用户在当前租户下生效的角色 code 列表。
//
// 被引用的角色必须同时满足：属于当前租户（FindByIDs 内做了租户过滤）、
// 且处于启用状态 —— 停用的角色不应继续参与鉴权，否则「停用角色」
// 这个操作对已登录用户不生效。
func (r *RBACRoleResolver) RoleCodesOf(tenantID, userID uint) ([]string, error) {
	roleIDs, err := repository.NewUserRepository().FindRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}

	roles, err := repository.NewRoleRepository().FindByIDs(tenantID, roleIDs)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Code != "" && role.Status == 1 {
			codes = append(codes, role.Code)
		}
	}
	return codes, nil
}
