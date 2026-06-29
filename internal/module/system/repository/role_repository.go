package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *model.SysRole) error
	FindByID(id uint) (*model.SysRole, error)
	FindByCode(code string) (*model.SysRole, error)
	FindList(name, code string, status *int8, page, pageSize int) ([]model.SysRole, int64, error)
	Update(role *model.SysRole) error
	Delete(id uint) error
	UpdateStatus(id uint, status int8) error
	ReplaceMenus(roleID uint, menuIDs []uint) error
	FindMenusByRoleID(roleID uint) ([]model.SysMenu, error)
	FindMenuIDsByRoleID(roleID uint) ([]uint, error)
	CountByCode(code string, excludeID uint) int64
	FindByIDs(ids []uint) ([]model.SysRole, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository() RoleRepository {
	return &roleRepository{db: database.DB}
}

func (r *roleRepository) Create(role *model.SysRole) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindByID(id uint) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *roleRepository) FindByCode(code string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.Where("code = ?", code).First(&role).Error
	return &role, err
}

func (r *roleRepository) FindList(name, code string, status *int8, page, pageSize int) ([]model.SysRole, int64, error) {
	var roles []model.SysRole
	var total int64

	query := r.db.Model(&model.SysRole{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&roles).Error
	return roles, total, err
}

func (r *roleRepository) Update(role *model.SysRole) error {
	return r.db.Model(role).Select("Name", "Code", "Sort", "Status", "DataScope", "Remark", "UpdateBy").Updates(role).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysRole{}, id).Error
}

func (r *roleRepository) UpdateStatus(id uint, status int8) error {
	return r.db.Model(&model.SysRole{}).Where("id = ?", id).Update("status", status).Error
}

func (r *roleRepository) ReplaceMenus(roleID uint, menuIDs []uint) error {
	r.db.Where("role_id = ?", roleID).Delete(&model.SysRoleMenu{})
	for _, menuID := range menuIDs {
		rm := model.SysRoleMenu{RoleID: roleID, MenuID: menuID}
		r.db.Create(&rm)
	}
	return nil
}

func (r *roleRepository) FindMenusByRoleID(roleID uint) ([]model.SysMenu, error) {
	var menus []model.SysMenu
	err := r.db.Joins("JOIN sys_role_menu ON sys_role_menu.menu_id = sys_menu.id").
		Where("sys_role_menu.role_id = ?", roleID).
		Order("sys_menu.sort ASC, sys_menu.id ASC").
		Find(&menus).Error
	return menus, err
}

func (r *roleRepository) FindMenuIDsByRoleID(roleID uint) ([]uint, error) {
	var menuIDs []uint
	err := r.db.Model(&model.SysRoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *roleRepository) CountByCode(code string, excludeID uint) int64 {
	var count int64
	query := r.db.Model(&model.SysRole{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	query.Count(&count)
	return count
}

func (r *roleRepository) FindByIDs(ids []uint) ([]model.SysRole, error) {
	var roles []model.SysRole
	err := r.db.Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}
