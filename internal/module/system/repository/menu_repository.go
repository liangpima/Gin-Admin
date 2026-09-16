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
	// FindParentID 返回菜单的父节点 ID，ok=false 表示菜单不存在（用于父级成环校验）
	FindParentID(id uint) (uint, bool, error)
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

// Delete 软删除菜单，并清理角色-菜单关联，避免留下孤儿记录
func (r *menuRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", id).Delete(&model.SysRoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.SysMenu{}, id).Error
	})
}

// FindParentID 查询菜单的父节点 ID。
// 用 Pluck 而非 First：记录不存在时返回空切片，无需额外处理 ErrRecordNotFound。
func (r *menuRepository) FindParentID(id uint) (uint, bool, error) {
	var parents []uint
	if err := r.db.Model(&model.SysMenu{}).Where("id = ?", id).Pluck("parent_id", &parents).Error; err != nil {
		return 0, false, err
	}
	if len(parents) == 0 {
		return 0, false, nil
	}
	return parents[0], true, nil
}
