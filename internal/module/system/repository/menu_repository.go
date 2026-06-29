package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type MenuRepository interface {
	Create(menu *model.SysMenu) error
	FindByID(id uint) (*model.SysMenu, error)
	FindAll() ([]model.SysMenu, error)
	FindAllForManage() ([]model.SysMenu, error)
	FindMenusByRoleIDs(roleIDs []uint) ([]model.SysMenu, error)
	Update(menu *model.SysMenu) error
	Delete(id uint) error
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository() MenuRepository {
	return &menuRepository{db: database.DB}
}

func (r *menuRepository) Create(menu *model.SysMenu) error {
	return r.db.Create(menu).Error
}

func (r *menuRepository) FindByID(id uint) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *menuRepository) FindAll() ([]model.SysMenu, error) {
	var menus []model.SysMenu
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *menuRepository) FindAllForManage() ([]model.SysMenu, error) {
	var menus []model.SysMenu
	err := r.db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *menuRepository) FindMenusByRoleIDs(roleIDs []uint) ([]model.SysMenu, error) {
	var menus []model.SysMenu
	err := r.db.Joins("JOIN sys_role_menu ON sys_role_menu.menu_id = sys_menu.id").
		Where("sys_role_menu.role_id IN ? AND sys_menu.status = ?", roleIDs, 1).
		Order("sys_menu.sort ASC, sys_menu.id ASC").
		Distinct().Find(&menus).Error
	return menus, err
}

func (r *menuRepository) Update(menu *model.SysMenu) error {
	return r.db.Model(menu).Select("ParentID", "Name", "Path", "Component", "Redirect", "Icon", "Title", "Type", "Permission", "Sort", "Visible", "Status", "IsExternal", "IsCache", "UpdateBy", "Remark").Updates(menu).Error
}

func (r *menuRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysMenu{}, id).Error
}
