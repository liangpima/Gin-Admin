package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type RoleService interface {
	Create(req *dto.CreateRoleRequest, operatorID uint) error
	Update(req *dto.UpdateRoleRequest, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (interface{}, error)
	FindList(req *dto.RoleListRequest) ([]interface{}, int64, error)
	UpdateStatus(req *dto.StatusRequest) error
	FindAll() ([]model.SysRole, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService() RoleService {
	return &roleService{
		roleRepo: repository.NewRoleRepository(),
	}
}

func (s *roleService) Create(req *dto.CreateRoleRequest, operatorID uint) error {
	if s.roleRepo.CountByCode(req.Code, 0) > 0 {
		return errors.New("角色编码已存在")
	}

	role := &model.SysRole{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
		},
		Name:      req.Name,
		Code:      req.Code,
		Sort:      req.Sort,
		Status:    req.Status,
		DataScope: req.DataScope,
	}
	role.Remark = req.Remark

	if err := s.roleRepo.Create(role); err != nil {
		return err
	}

	if len(req.MenuIds) > 0 {
		_ = s.roleRepo.ReplaceMenus(role.ID, req.MenuIds)
	}

	return nil
}

func (s *roleService) Update(req *dto.UpdateRoleRequest, operatorID uint) error {
	if s.roleRepo.CountByCode(req.Code, req.ID) > 0 {
		return errors.New("角色编码已存在")
	}

	role, err := s.roleRepo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
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

	if err := s.roleRepo.Update(role); err != nil {
		return err
	}

	if req.MenuIds != nil {
		_ = s.roleRepo.ReplaceMenus(role.ID, req.MenuIds)
	}

	return nil
}

func (s *roleService) Delete(id uint) error {
	return s.roleRepo.Delete(id)
}

func (s *roleService) FindByID(id uint) (interface{}, error) {
	return s.roleRepo.FindByID(id)
}

func (s *roleService) FindList(req *dto.RoleListRequest) ([]interface{}, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	roles, total, err := s.roleRepo.FindList(req.Name, req.Code, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]interface{}, len(roles))
	for i, r := range roles {
		result[i] = r
	}
	return result, total, nil
}

func (s *roleService) UpdateStatus(req *dto.StatusRequest) error {
	return s.roleRepo.UpdateStatus(req.ID, req.Status)
}

func (s *roleService) FindAll() ([]model.SysRole, error) {
	roles, _, err := s.roleRepo.FindList("", "", nil, 1, 1000)
	return roles, err
}
