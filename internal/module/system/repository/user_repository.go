package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.SysUser) error
	FindByID(id uint) (*model.SysUser, error)
	FindByUsername(username string) (*model.SysUser, error)
	FindList(username, phone string, status *int8, deptID uint, page, pageSize int) ([]model.SysUser, int64, error)
	Update(user *model.SysUser) error
	Delete(id uint) error
	UpdateStatus(id uint, status int8) error
	ResetPassword(id uint, password string) error
	ReplaceRoles(userID uint, roleIDs []uint) error
	ReplacePosts(userID uint, postIDs []uint) error
	FindRolesByUserID(userID uint) ([]model.SysRole, error)
	FindRoleIDsByUserID(userID uint) ([]uint, error)
	CountByUsername(username string, excludeID uint) int64
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

func (r *userRepository) FindByID(id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByUsername(username string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindList(username, phone string, status *int8, deptID uint, page, pageSize int) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	query := r.db.Model(&model.SysUser{})

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

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysUser{}, id).Error
}

func (r *userRepository) UpdateStatus(id uint, status int8) error {
	return r.db.Model(&model.SysUser{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) ResetPassword(id uint, password string) error {
	return r.db.Model(&model.SysUser{}).Where("id = ?", id).Update("password", password).Error
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

func (r *userRepository) FindRolesByUserID(userID uint) ([]model.SysRole, error) {
	var roles []model.SysRole
	err := r.db.Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role.id").
		Where("sys_user_role.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *userRepository) FindRoleIDsByUserID(userID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.Model(&model.SysUserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRepository) CountByUsername(username string, excludeID uint) int64 {
	var count int64
	query := r.db.Model(&model.SysUser{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	query.Count(&count)
	return count
}
