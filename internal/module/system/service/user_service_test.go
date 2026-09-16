package service

import (
	"errors"
	"testing"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
)

// stubRoleService 只实现 normalizeRoleIDs 会触达的 FindByIDs，
// 其余方法不会被调用，返回零值即可。
type stubRoleService struct {
	// roles 模拟「按 tenantID 过滤后」查到的角色
	roles []model.SysRole
	err   error
	// lastIDs 记录实际收到的 ID 列表，用于校验去重/过滤行为
	lastIDs []uint
}

func (s *stubRoleService) Create(req *dto.CreateRoleRequest, operatorID, tenantID uint) error {
	return nil
}
func (s *stubRoleService) Update(req *dto.UpdateRoleRequest, operatorID, tenantID uint) error {
	return nil
}
func (s *stubRoleService) Delete(tenantID, id uint) error { return nil }
func (s *stubRoleService) FindByID(tenantID, id uint) (interface{}, error) {
	return nil, nil
}
func (s *stubRoleService) FindByIDs(tenantID uint, ids []uint) ([]model.SysRole, error) {
	s.lastIDs = ids
	return s.roles, s.err
}
func (s *stubRoleService) FindList(tenantID uint, req *dto.RoleListRequest) ([]interface{}, int64, error) {
	return nil, 0, nil
}
func (s *stubRoleService) UpdateStatus(tenantID uint, req *dto.StatusRequest) error { return nil }
func (s *stubRoleService) FindAll(tenantID uint) ([]model.SysRole, error)          { return nil, nil }

// TestNormalizeRoleIDsRejectsForeignRole 回归保护：跨租户角色分配必须被拒绝。
//
// sys_user_role 是纯关联表（只有 user_id / role_id，没有 tenant_id 列），
// 租户隔离无法靠 TenantScope 完成。不校验的后果是租户 A 的管理员可以把用户绑到
// 租户 B 的角色 ID 上（ID 可枚举），而 Casbin 策略是按角色 code 生成的，
// 绑上即继承对方的菜单与权限码 —— 跨租户权限提升。
func TestNormalizeRoleIDsRejectsForeignRole(t *testing.T) {
	// 请求绑 2 个角色，但按本租户过滤后只查到 1 个 → 另一个属于其他租户
	stub := &stubRoleService{roles: []model.SysRole{{Code: "editor"}}}
	svc := &userService{roleService: stub}

	_, err := svc.normalizeRoleIDs(2, []uint{1, 99})
	if err == nil {
		t.Fatal("包含非本租户的角色 ID 时必须拒绝")
	}
	if !common.IsBizError(err) {
		t.Errorf("应返回业务错误(400)，实际: %T", err)
	}
}

// TestNormalizeRoleIDsAcceptsOwnRoles 本租户角色应正常通过。
func TestNormalizeRoleIDsAcceptsOwnRoles(t *testing.T) {
	stub := &stubRoleService{roles: []model.SysRole{{Code: "editor"}, {Code: "viewer"}}}
	svc := &userService{roleService: stub}

	got, err := svc.normalizeRoleIDs(2, []uint{1, 2})
	if err != nil {
		t.Fatalf("本租户角色不应报错: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("期望 2 个角色，实际 %d", len(got))
	}
}

// TestNormalizeRoleIDsDedupesAndDropsZero 重复项与 0 应被剔除后再查询。
func TestNormalizeRoleIDsDedupesAndDropsZero(t *testing.T) {
	stub := &stubRoleService{roles: []model.SysRole{{Code: "editor"}}}
	svc := &userService{roleService: stub}

	got, err := svc.normalizeRoleIDs(1, []uint{1, 1, 0, 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("重复项与 0 应被剔除，实际返回 %v", got)
	}
	if len(stub.lastIDs) != 1 || stub.lastIDs[0] != 1 {
		t.Errorf("传给 FindByIDs 的应为去重后的 [1]，实际 %v", stub.lastIDs)
	}
}

// TestNormalizeRoleIDsEmptySkipsQuery 空输入不应触发查询。
func TestNormalizeRoleIDsEmptySkipsQuery(t *testing.T) {
	stub := &stubRoleService{err: errors.New("FindByIDs 不应被调用")}
	svc := &userService{roleService: stub}

	for _, in := range [][]uint{nil, {}, {0}} {
		got, err := svc.normalizeRoleIDs(1, in)
		if err != nil {
			t.Errorf("空输入不应报错, in=%v, err=%v", in, err)
		}
		if got != nil {
			t.Errorf("空输入应返回 nil, in=%v, got=%v", in, got)
		}
	}

	// 传空数组表示「清空角色」，调用方会走到 ReplaceRoles(userID, nil)
	got, err := svc.normalizeRoleIDs(1, []uint{})
	if err != nil || got != nil {
		t.Errorf("空数组应返回 (nil, nil) 以便清空角色, got=%v err=%v", got, err)
	}
}

// TestNormalizeRoleIDsPropagatesSystemError DB 故障必须原样上抛，
// 不能被包装成 400 业务错误 —— 否则数据库挂了会显示成「参数不合法」。
func TestNormalizeRoleIDsPropagatesSystemError(t *testing.T) {
	stub := &stubRoleService{err: errors.New("db down")}
	svc := &userService{roleService: stub}

	_, err := svc.normalizeRoleIDs(1, []uint{1})
	if err == nil {
		t.Fatal("系统错误必须上抛")
	}
	if common.IsBizError(err) {
		t.Error("系统错误不应被包装成业务错误")
	}
}

// stubPostService 与 stubRoleService 同构，只实现 normalizePostIDs 会触达的 FindByIDs。
type stubPostService struct {
	posts   []model.SysPost
	err     error
	lastIDs []uint
}

func (s *stubPostService) Create(tenantID uint, name, code string, sort int, status int8, operatorID uint) error {
	return nil
}
func (s *stubPostService) Update(tenantID, id uint, name, code string, sort int, status int8, operatorID uint) error {
	return nil
}
func (s *stubPostService) Delete(tenantID, id uint) error { return nil }
func (s *stubPostService) FindByID(tenantID, id uint) (interface{}, error) {
	return nil, nil
}
func (s *stubPostService) FindByIDs(tenantID uint, ids []uint) ([]model.SysPost, error) {
	s.lastIDs = ids
	return s.posts, s.err
}
func (s *stubPostService) FindAll(tenantID uint) ([]model.SysPost, error) { return nil, nil }
func (s *stubPostService) FindList(tenantID uint, name string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	return nil, 0, nil
}
func (s *stubPostService) UpdateStatus(tenantID, id uint, status int8) error { return nil }

// TestNormalizePostIDsRejectsForeignPost 岗位已改为租户内数据，
// 因此与角色同理：sys_user_post 是纯关联表（无 tenant_id），
// 必须确认岗位属于当前租户后再绑定，否则可拿其他租户的岗位 ID 建立跨租户绑定。
func TestNormalizePostIDsRejectsForeignPost(t *testing.T) {
	stub := &stubPostService{posts: []model.SysPost{{Code: "dev"}}}
	svc := &userService{postService: stub}

	_, err := svc.normalizePostIDs(2, []uint{1, 99})
	if err == nil {
		t.Fatal("包含非本租户的岗位 ID 时必须拒绝")
	}
	if !common.IsBizError(err) {
		t.Errorf("应返回业务错误(400)，实际: %T", err)
	}
}

func TestNormalizePostIDsAcceptsOwnPosts(t *testing.T) {
	stub := &stubPostService{posts: []model.SysPost{{Code: "dev"}, {Code: "ops"}}}
	svc := &userService{postService: stub}

	got, err := svc.normalizePostIDs(2, []uint{1, 2})
	if err != nil {
		t.Fatalf("本租户岗位不应报错: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("期望 2 个岗位，实际 %d", len(got))
	}
}

func TestNormalizePostIDsDedupesAndDropsZero(t *testing.T) {
	stub := &stubPostService{posts: []model.SysPost{{Code: "dev"}}}
	svc := &userService{postService: stub}

	got, err := svc.normalizePostIDs(1, []uint{1, 1, 0, 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("重复项与 0 应被剔除，实际返回 %v", got)
	}
	if len(stub.lastIDs) != 1 || stub.lastIDs[0] != 1 {
		t.Errorf("传给 FindByIDs 的应为去重后的 [1]，实际 %v", stub.lastIDs)
	}
}

func TestNormalizePostIDsEmptySkipsQuery(t *testing.T) {
	stub := &stubPostService{err: errors.New("FindByIDs 不应被调用")}
	svc := &userService{postService: stub}

	for _, in := range [][]uint{nil, {}, {0}} {
		got, err := svc.normalizePostIDs(1, in)
		if err != nil {
			t.Errorf("空输入不应报错, in=%v, err=%v", in, err)
		}
		if got != nil {
			t.Errorf("空输入应返回 nil, in=%v, got=%v", in, got)
		}
	}
}

func TestNormalizePostIDsPropagatesSystemError(t *testing.T) {
	stub := &stubPostService{err: errors.New("db down")}
	svc := &userService{postService: stub}

	_, err := svc.normalizePostIDs(1, []uint{1})
	if err == nil {
		t.Fatal("系统错误必须上抛")
	}
	if common.IsBizError(err) {
		t.Error("系统错误不应被包装成业务错误")
	}
}
