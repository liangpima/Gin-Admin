package service

import (
	"errors"

	"go-admin/internal/common"
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

	return s.menuRepo.Create(menu)
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

	return s.menuRepo.Update(menu)
}

func (s *menuService) Delete(id uint) error {
	return s.menuRepo.Delete(id)
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
