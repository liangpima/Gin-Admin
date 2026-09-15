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

type MenuService interface {
	Create(req *dto.CreateMenuRequest, operatorID uint) error
	Update(req *dto.UpdateMenuRequest, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (interface{}, error)
	FindAll() ([]model.SysMenu, error)
	FindTree() ([]model.SysMenu, error)
	FindTreeForManage() ([]model.SysMenu, error)
	FindMenusByRoleIDs(roleIDs []uint) ([]model.SysMenu, error)
}

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService() MenuService {
	return &menuService{
		menuRepo: repository.NewMenuRepository(),
	}
}

func (s *menuService) Create(req *dto.CreateMenuRequest, operatorID uint) error {
	menu := &model.SysMenu{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		ParentID:   req.ParentID,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Redirect:   req.Redirect,
		Icon:       req.Icon,
		Title:      req.Title,
		Type:       req.Type,
		Permission: req.Permission,
		Sort:       req.Sort,
		Visible:    req.Visible,
		Status:     req.Status,
		IsExternal: req.IsExternal,
		IsCache:    req.IsCache,
	}

	if err := s.menuRepo.Create(menu); err != nil {
		return err
	}
	s.syncPolicies()
	return nil
}

func (s *menuService) Update(req *dto.UpdateMenuRequest, operatorID uint) error {
	menu, err := s.menuRepo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return err
	}

	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Redirect = req.Redirect
	menu.Icon = req.Icon
	menu.Title = req.Title
	menu.Type = req.Type
	menu.Permission = req.Permission
	menu.Sort = req.Sort
	menu.Visible = req.Visible
	menu.Status = req.Status
	menu.IsExternal = req.IsExternal
	menu.IsCache = req.IsCache
	menu.UpdateBy = operatorID

	if err := s.menuRepo.Update(menu); err != nil {
		return err
	}
	s.syncPolicies()
	return nil
}

func (s *menuService) Delete(id uint) error {
	if err := s.menuRepo.Delete(id); err != nil {
		return err
	}
	s.syncPolicies()
	return nil
}

// syncPolicies 菜单的权限码增删改会影响策略，需重建并清理角色缓存。
func (s *menuService) syncPolicies() {
	if err := middleware.SyncPoliciesFromRoleMenus(); err != nil {
		logger.Log.Errorf("同步权限策略失败: %v", err)
	}
	if err := cache.DelByPrefix(context.Background(), "rbac:roles:"); err != nil {
		logger.Log.Warnf("清理角色缓存失败: %v", err)
	}
}

func (s *menuService) FindByID(id uint) (interface{}, error) {
	return s.menuRepo.FindByID(id)
}

func (s *menuService) FindAll() ([]model.SysMenu, error) {
	return s.menuRepo.FindAll()
}

func (s *menuService) FindTree() ([]model.SysMenu, error) {
	menus, err := s.menuRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

func (s *menuService) FindTreeForManage() ([]model.SysMenu, error) {
	menus, err := s.menuRepo.FindAllForManage()
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

func (s *menuService) FindMenusByRoleIDs(roleIDs []uint) ([]model.SysMenu, error) {
	return s.menuRepo.FindMenusByRoleIDs(roleIDs)
}
