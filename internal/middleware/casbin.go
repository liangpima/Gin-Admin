package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/logger"
	"go-admin/internal/module/system/repository"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

var enforcer *casbin.Enforcer

// rbacDomain 与 config/casbin/model.conf 中的 dom 对应。
// 策略记录（p）中的 dom 字段必须与此值一致，否则匹配会失败。
//
// 这里使用固定域：租户数据隔离已在 Repository 层解决，
// 且 sys_role.code 为全局唯一索引，不同租户的角色 code 不会冲突。
const rbacDomain = "default"

// rbacRoleCacheTTL 用户角色的缓存时长。
// 角色变更后，鉴权结果最多需要这么久才生效。
const rbacRoleCacheTTL = 60 * time.Second

// adminRoleCode 超级管理员角色 code，始终放行全部权限
const adminRoleCode = "admin"

func InitCasbin(modelPath string) error {
	// 确保策略表存在（幂等），避免部署时因漏执行建表语句导致鉴权静默失效
	if err := database.DB.AutoMigrate(&CasbinRule{}); err != nil {
		return fmt.Errorf("创建 casbin_rule 表失败: %w", err)
	}

	var err error
	enforcer, err = casbin.NewEnforcer(modelPath, newGormAdapter(database.DB))
	if err != nil {
		return err
	}

	return SyncPoliciesFromRoleMenus()
}

// SyncPoliciesFromRoleMenus 依据「角色-菜单」授权关系重建权限策略。
//
// 菜单上配置的 permission 即权限码，路由通过 middleware.Perm(code) 声明所需权限码，
// 因此给角色分配菜单就等同于分配权限，不需要手工维护 casbin_rule。
// 在启动时以及角色/菜单发生变更后调用。
func SyncPoliciesFromRoleMenus() error {
	if enforcer == nil {
		return errors.New("casbin 未初始化")
	}

	// 读取「角色 code -> 权限码」映射（菜单 permission 为空表示仅作导航，不产生权限）
	type permRow struct {
		RoleCode   string
		Permission string
	}
	var rows []permRow
	if err := database.DB.Table("sys_role_menu AS rm").
		Select("r.code AS role_code, m.permission AS permission").
		Joins("JOIN sys_role AS r ON r.id = rm.role_id AND r.deleted_at IS NULL").
		Joins("JOIN sys_menu AS m ON m.id = rm.menu_id AND m.deleted_at IS NULL").
		Where("m.permission <> ''").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("读取角色菜单权限失败: %w", err)
	}

	// 清空旧策略：先清库表，再让 enforcer 重新加载（此时为空），避免内存与库不一致
	if err := database.DB.Where("1 = 1").Delete(&CasbinRule{}).Error; err != nil {
		return fmt.Errorf("清空权限策略失败: %w", err)
	}
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("重新加载权限策略失败: %w", err)
	}

	// 超级管理员始终放行全部权限
	rules := [][]string{{adminRoleCode, rbacDomain, "*", "*"}}
	seen := map[string]bool{adminRoleCode + "|*": true}

	for _, r := range rows {
		if r.RoleCode == "" || r.Permission == "" {
			continue
		}
		key := r.RoleCode + "|" + r.Permission
		if seen[key] {
			continue
		}
		seen[key] = true
		rules = append(rules, []string{r.RoleCode, rbacDomain, r.Permission, "*"})
	}

	if _, err := enforcer.AddPolicies(rules); err != nil {
		return fmt.Errorf("写入权限策略失败: %w", err)
	}

	logger.Log.Infof("[casbin] 已根据角色-菜单关系同步 %d 条权限策略", len(rules))
	return nil
}

func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 初始化失败时 enforcer 为 nil，此时无法鉴权，放行并由启动日志告警
		if enforcer == nil {
			c.Next()
			return
		}

		// 依据路由权限登记表确定所需权限码
		code, registered := routePermission(c.Request.Method, c.FullPath())
		if !registered {
			// 未登记的路由默认拒绝，避免新增接口漏配权限后被放行
			logger.Log.Warnf("[casbin] 路由未登记权限码，已拒绝: %s %s", c.Request.Method, c.FullPath())
			common.Forbidden(c, "该接口未配置访问权限")
			c.Abort()
			return
		}
		// 空权限码表示仅要求登录态（自助接口）
		if code == "" {
			c.Next()
			return
		}

		roles := resolveRoleCodes(c)
		if len(roles) == 0 {
			common.Forbidden(c, "当前账号未分配角色，无法访问")
			c.Abort()
			return
		}

		// model.conf 的 request_definition 为四元组 (sub, dom, obj, act)。
		// 用户可能拥有多个角色，任一角色命中该权限码即放行。
		for _, role := range roles {
			ok, err := enforcer.Enforce(role, rbacDomain, code, c.Request.Method)
			if err != nil {
				logger.Log.Errorf("[casbin] 权限校验出错: %v", err)
				common.Error(c, common.CodeInternalError, "权限校验失败")
				c.Abort()
				return
			}
			if ok {
				c.Next()
				return
			}
		}

		common.Forbidden(c, fmt.Sprintf("没有权限访问（需要 %s）", code))
		c.Abort()
	}
}

// resolveRoleCodes 解析当前用户的角色 code 列表，作为 RBAC 的匹配主体。
// 结果短时缓存于 Redis，避免每个请求都查库。
func resolveRoleCodes(c *gin.Context) []string {
	userID := common.GetCurrentUserID(c)
	if userID == 0 {
		return nil
	}
	tenantID := common.GetTenantID(c)

	ctx := context.Background()
	cacheKey := fmt.Sprintf("rbac:roles:%d:%d", tenantID, userID)

	if v, err := cache.Get(ctx, cacheKey); err == nil {
		if codes := splitRoleCodes(v); len(codes) > 0 {
			return codes
		}
	}

	roleIDs, err := repository.NewUserRepository().FindRoleIDsByUserID(userID)
	if err != nil || len(roleIDs) == 0 {
		return nil
	}

	roles, err := repository.NewRoleRepository().FindByIDs(tenantID, roleIDs)
	if err != nil {
		return nil
	}

	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		if r.Code != "" {
			codes = append(codes, r.Code)
		}
	}

	if len(codes) > 0 {
		_ = cache.Set(ctx, cacheKey, strings.Join(codes, ","), rbacRoleCacheTTL)
	}
	return codes
}

func splitRoleCodes(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	codes := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			codes = append(codes, p)
		}
	}
	return codes
}
