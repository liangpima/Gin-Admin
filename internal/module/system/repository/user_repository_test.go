package repository

import (
	"strings"
	"testing"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/testsupport"
)

const testTenant uint = 1

func newUserRepoWithDB(t *testing.T) UserRepository {
	t.Helper()
	// 先建库（会注入 database.DB），再构造仓储 —— 仓储在构造时捕获 database.DB
	testsupport.NewDB(t, &model.SysUser{}, &model.SysUserRole{}, &model.SysUserPost{})
	return NewUserRepository()
}

func seedUser(t *testing.T, repo UserRepository, username string, tenantID uint) *model.SysUser {
	t.Helper()
	u := &model.SysUser{
		TenantBaseModel: common.TenantBaseModel{TenantID: tenantID},
		Username:        username,
		Password:        "x",
		Nickname:        "测试",
		Status:          1,
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return u
}

// TestUserRepositoryDeleteReleasesUsername 软删除必须释放唯一索引占用的用户名。
//
// 这是本项目真实踩过的坑：唯一索引不区分记录是否已软删除，
// 删除时若不改写 username，同名用户就再也建不出来（Duplicate entry）。
// 而且报错发生在 INSERT 阶段，前端只会看到一句数据库错误，很难定位。
func TestUserRepositoryDeleteReleasesUsername(t *testing.T) {
	repo := newUserRepoWithDB(t)

	u := seedUser(t, repo, "admin", testTenant)
	if err := repo.Delete(testTenant, u.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	t.Run("删除后用户名被改写以释放唯一值", func(t *testing.T) {
		got, err := repo.FindByUsername(testTenant, "admin")
		if err == nil && got.ID != 0 {
			t.Error("已删除的用户不应还能按原名查到")
		}
	})

	t.Run("同名用户应能重新创建", func(t *testing.T) {
		again := &model.SysUser{
			TenantBaseModel: common.TenantBaseModel{TenantID: testTenant},
			Username:        "admin",
			Password:        "x",
			Status:          1,
		}
		if err := repo.Create(again); err != nil {
			t.Fatalf("同名用户无法重建（软删除未释放唯一值）: %v", err)
		}
		if again.ID == u.ID {
			t.Error("应创建出一条新记录")
		}
	})
}

// TestUserRepositoryDeleteCleansRelations 删除用户要顺带清掉关联表，避免孤儿数据
func TestUserRepositoryDeleteCleansRelations(t *testing.T) {
	repo := newUserRepoWithDB(t)
	u := seedUser(t, repo, "temp", testTenant)

	if err := repo.ReplaceRoles(u.ID, []uint{1, 2}); err != nil {
		t.Fatalf("写入角色关联失败: %v", err)
	}
	if err := repo.ReplacePosts(u.ID, []uint{5}); err != nil {
		t.Fatalf("写入岗位关联失败: %v", err)
	}

	if err := repo.Delete(testTenant, u.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	roleIDs, err := repo.FindRoleIDsByUserID(u.ID)
	if err != nil {
		t.Fatalf("查询角色关联失败: %v", err)
	}
	if len(roleIDs) != 0 {
		t.Errorf("角色关联应被清理，实际残留 %v", roleIDs)
	}
}

// TestUserRepositoryTenantIsolation 仓储层的租户隔离。
//
// 这条对应「漏传 tenantID 会静默退化为全表查询」这一最典型的越权成因：
// 两个租户各有一条同名用户，按租户查询只能看到自己那条。
func TestUserRepositoryTenantIsolation(t *testing.T) {
	repo := newUserRepoWithDB(t)

	// 注意：username 的唯一索引是全局的，所以两个租户要用不同用户名，
	// 这里考察的是「查询是否带租户过滤」，不是唯一约束语义。
	seedUser(t, repo, "user-a", 1)
	seedUser(t, repo, "user-b", 2)

	t.Run("按租户过滤列表", func(t *testing.T) {
		users, total, err := repo.FindList(1, "", "", nil, 0, 1, 100)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if total != 1 || len(users) != 1 {
			t.Fatalf("租户1 应只看到 1 条，实际 total=%d len=%d", total, len(users))
		}
		if users[0].Username != "user-a" {
			t.Errorf("看到了其他租户的数据: %s", users[0].Username)
		}
	})

	t.Run("跨租户按 ID 查不到", func(t *testing.T) {
		u, _, _ := repo.FindList(2, "", "", nil, 0, 1, 100)
		if len(u) != 1 {
			t.Fatalf("准备数据异常: %d", len(u))
		}

		// 用租户 1 的身份去查租户 2 的用户
		if got, err := repo.FindByID(1, u[0].ID); err == nil {
			t.Errorf("跨租户查询应失败，实际查到了: %+v", got)
		}
	})
}

// TestUserRepositoryCountByUsername 重名校验是**全局**的，不按租户隔离。
//
// 这不是漏加租户过滤，而是与 uk_username 这个全局唯一索引保持一致：
// 用户名是登录标识，登录时只凭 username 定位账号（FindByUsernameForAuth 无租户参数），
// 若允许两个租户各有一个同名用户，登录时无法判断该进哪个租户。
//
// 之所以要写这条用例：签名里没有 tenantID 参数，很容易被后来者当成 bug
// 「顺手补上」租户过滤 —— 那样校验会放行、INSERT 却撞唯一索引，
// 对外表现为 500，用户完全看不出是重名。
func TestUserRepositoryCountByUsername(t *testing.T) {
	repo := newUserRepoWithDB(t)
	seeded := seedUser(t, repo, "dup", testTenant)

	// 换一个租户建同名用户：唯一索引会拒绝，说明唯一性确实跨租户
	dupOtherTenant := &model.SysUser{
		TenantBaseModel: common.TenantBaseModel{TenantID: 2},
		Username:        "dup",
		Password:        "x",
		Nickname:        "另一个租户",
		Status:          1,
	}
	if err := repo.Create(dupOtherTenant); err == nil {
		t.Errorf("跨租户同名用户应被唯一索引拒绝，实际创建成功")
	}

	n, err := repo.CountByUsername("dup", 0)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if n != 1 {
		t.Errorf("同名用户应计到 1 条（全局，含其他租户视角），实际 %d", n)
	}

	n, err = repo.CountByUsername("dup", seeded.ID)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if n != 0 {
		t.Errorf("排除自身后应计到 0 条，实际 %d", n)
	}

	n, err = repo.CountByUsername("not-exist", 0)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if n != 0 {
		t.Errorf("不存在的用户名应计到 0，实际 %d", n)
	}
}

// TestUserRepositorySoftDeletedNotCounted 已软删除的记录不应出现在正常查询里
func TestUserRepositorySoftDeletedNotCounted(t *testing.T) {
	repo := newUserRepoWithDB(t)
	u := seedUser(t, repo, "gone", testTenant)

	if err := repo.Delete(testTenant, u.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	users, total, err := repo.FindList(testTenant, "", "", nil, 0, 1, 100)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if total != 0 || len(users) != 0 {
		t.Errorf("软删除记录不应出现在列表中，实际 total=%d", total)
	}
}

// TestFreedUsernameFormat 顺带确认改写后的用户名形态，避免有人误改成明文覆盖
func TestFreedUsernameFormat(t *testing.T) {
	// Delete 会一并清理关联表，因此这里也要建出来
	db := testsupport.NewDB(t, &model.SysUser{}, &model.SysUserRole{}, &model.SysUserPost{})
	repo := NewUserRepository()

	u := seedUser(t, repo, "alice", testTenant)
	if err := repo.Delete(testTenant, u.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	var raw model.SysUser
	if err := db.Unscoped().First(&raw, u.ID).Error; err != nil {
		t.Fatalf("读取原始记录失败: %v", err)
	}
	if !strings.HasPrefix(raw.Username, "alice_del_") {
		t.Errorf("用户名应被改写为 alice_del_<id>，实际 %q", raw.Username)
	}
}
