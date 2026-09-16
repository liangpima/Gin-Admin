package service

import (
	"context"
	"errors"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/logger"
	"go-admin/internal/middleware"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type RoleService interface {
	Create(req *dto.CreateRoleRequest, operatorID, tenantID uint) error
	Update(req *dto.UpdateRoleRequest, operatorID, tenantID uint) error
	Delete(tenantID, id uint) error
	FindByID(tenantID, id uint) (interface{}, error)
	FindByIDs(tenantID uint, ids []uint) ([]model.SysRole, error)
	FindList(tenantID uint, req *dto.RoleListRequest) ([]interface{}, int64, error)
	UpdateStatus(tenantID uint, req *dto.StatusRequest) error
	FindAll(tenantID uint) ([]model.SysRole, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService() RoleService {
	return &roleService{
		roleRepo: repository.NewRoleRepository(),
	}
}

func (s *roleService) Create(req *dto.CreateRoleRequest, operatorID, tenantID uint) error {
	// 角色 code 必须**全局**唯一：Casbin 策略的主体是角色 code，域是常量 "default"，
	// 两个租户用同一 code 会导致策略合并、互相继承权限（跨租户泄漏）。
	// 所以重名校验也必须按全局来，否则会出现「校验通过、插入撞唯一索引」→ 500
	if count, err := s.roleRepo.CountByCode(req.Code, 0); err != nil {
		return err
	} else if count > 0 {
		return common.NewBizError("角色编码已存在")
	}

	role := &model.SysRole{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
			TenantID: tenantID,
		},
		Name:      req.Name,
		Code:      req.Code,
		Sort:      req.Sort,
		Status:    req.Status,
		DataScope: req.DataScope,
	}
	role.Remark = req.Remark

	if err := s.roleRepo.Create(role); err != nil {
		// 上面 Count 校验有时间窗口，并发下靠唯一索引兜底
		if errors.Is(err, common.ErrDuplicateKey) {
			return common.NewBizError("角色编码已存在")
		}
		return err
	}

	if len(req.MenuIds) > 0 {
		if err := s.roleRepo.ReplaceMenus(tenantID, role.ID, req.MenuIds); err != nil {
			return err
		}
		s.syncPolicies()
	}

	return nil
}

func (s *roleService) Update(req *dto.UpdateRoleRequest, operatorID, tenantID uint) error {
	// 同 Create：角色 code 全局唯一，改编码时也按全局校验
	if count, err := s.roleRepo.CountByCode(req.Code, req.ID); err != nil {
		return err
	} else if count > 0 {
		return common.NewBizError("角色编码已存在")
	}

	role, err := s.roleRepo.FindByID(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewNotFoundError("角色不存在")
		}
		return err
	}

	role.Name = req.Name
	role.Code = req.Code
	role.Sort = req.Sort
	role.Status = req.Status
	role.DataScope = req.DataScope
	role.Remark = req.Remark
	role.UpdateBy = operatorID

	if err := s.roleRepo.Update(tenantID, role); err != nil {
		if errors.Is(err, common.ErrDuplicateKey) {
			return common.NewBizError("角色编码已存在")
		}
		return err
	}

	if req.MenuIds != nil {
		if err := s.roleRepo.ReplaceMenus(tenantID, role.ID, req.MenuIds); err != nil {
			return err
		}
	}

	// 角色编码/状态/授权变化都会影响策略，必须同步（不能只在菜单变更时同步）
	s.syncPolicies()

	return nil
}

func (s *roleService) Delete(tenantID, id uint) error {
	if err := s.roleRepo.Delete(tenantID, id); err != nil {
		return err
	}
	s.syncPolicies()
	return nil
}

// syncPolicies 角色授权变更后重建 Casbin 策略并清理角色缓存，使权限立即生效。
// 失败仅记录日志：授权数据已落库，下次启动会重新同步。
func (s *roleService) syncPolicies() {
	if err := middleware.SyncPoliciesFromRoleMenus(); err != nil {
		logger.Log.Errorf("同步权限策略失败: %v", err)
	}
	// 角色编码/状态变更后，用户角色缓存需失效，否则最长 60s 内仍按旧角色鉴权
	if err := cache.DelByPrefix(context.Background(), "rbac:roles:"); err != nil {
		logger.Log.Warnf("清理角色缓存失败: %v", err)
	}
}

func (s *roleService) FindByID(tenantID, id uint) (interface{}, error) {
	role, err := s.roleRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "角色不存在")
	}
	return role, nil
}

func (s *roleService) FindByIDs(tenantID uint, ids []uint) ([]model.SysRole, error) {
	return s.roleRepo.FindByIDs(tenantID, ids)
}

func (s *roleService) FindList(tenantID uint, req *dto.RoleListRequest) ([]interface{}, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	req.PageSize = common.NormalizePageSize(req.PageSize)

	roles, total, err := s.roleRepo.FindList(tenantID, req.Name, req.Code, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]interface{}, len(roles))
	for i, r := range roles {
		result[i] = r
	}
	return result, total, nil
}

func (s *roleService) UpdateStatus(tenantID uint, req *dto.StatusRequest) error {
	return s.roleRepo.UpdateStatus(tenantID, req.ID, req.Status)
}

func (s *roleService) FindAll(tenantID uint) ([]model.SysRole, error) {
	roles, _, err := s.roleRepo.FindList(tenantID, "", "", nil, 1, 1000)
	return roles, err
}
