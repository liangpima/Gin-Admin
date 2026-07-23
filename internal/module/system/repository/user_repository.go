package repository

import (
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.SysUser) error
	FindByID(tenantID, id uint) (*model.SysUser, error)
	FindByUsername(tenantID uint, username string) (*model.SysUser, error)
	FindByUsernameForAuth(username string) (*model.SysUser, error)
	FindList(tenantID uint, username, phone string, status *int8, deptID uint, page, pageSize int) ([]model.SysUser, int64, error)
	Update(user *model.SysUser) error
	Delete(tenantID, id uint) error
	UpdateStatus(tenantID, id uint, status int8) error
	ResetPassword(tenantID, id uint, password string) error
	ReplaceRoles(userID uint, roleIDs []uint) error
	ReplacePosts(userID uint, postIDs []uint) error
	FindRoleIDsByUserID(userID uint) ([]uint, error)
	CountByUsername(tenantID uint, username string, excludeID uint) int64
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.DB}
}

func (r *userRepository) Create(user *model.SysUser) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(tenantID, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := common.TenantScope(r.db, tenantID).First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByUsername(tenantID uint, username string) (*model.SysUser, error) {
	var user model.SysUser
	err := common.TenantScope(r.db, tenantID).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByUsernameForAuth(username string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindList(tenantID uint, username, phone string, status *int8, deptID uint, page, pageSize int) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	query := common.TenantScope(r.db.Model(&model.SysUser{}), tenantID)

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if deptID > 0 {
		query = query.Where("dept_id = ?", deptID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&users).Error
	return users, total, err
}

func (r *userRepository) Update(user *model.SysUser) error {
	return r.db.Model(user).Select("Username", "Nickname", "Phone", "Email", "Avatar", "Password", "Status", "DeptID", "Remark", "UpdateBy").Updates(user).Error
}

func (r *userRepository) Delete(tenantID, id uint) error {
	return common.TenantScope(r.db, tenantID).Delete(&model.SysUser{}, id).Error
}

func (r *userRepository) UpdateStatus(tenantID, id uint, status int8) error {
	return common.TenantScope(r.db, tenantID).Model(&model.SysUser{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) ResetPassword(tenantID, id uint, password string) error {
	return common.TenantScope(r.db, tenantID).Model(&model.SysUser{}).Where("id = ?", id).Update("password", password).Error
}

func (r *userRepository) ReplaceRoles(userID uint, roleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			ur := model.SysUserRole{UserID: userID, RoleID: roleID}
			if err := tx.Create(&ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRepository) ReplacePosts(userID uint, postIDs []uint) error {
	r.db.Where("user_id = ?", userID).Delete(&model.SysUserPost{})
	for _, postID := range postIDs {
		up := model.SysUserPost{UserID: userID, PostID: postID}
		r.db.Create(&up)
	}
	return nil
}

func (r *userRepository) FindRoleIDsByUserID(userID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.Model(&model.SysUserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRepository) CountByUsername(tenantID uint, username string, excludeID uint) int64 {
	var count int64
	query := common.TenantScope(r.db.Model(&model.SysUser{}), tenantID).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	query.Count(&count)
	return count
}
