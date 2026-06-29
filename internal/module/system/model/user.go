package model

import (
	"time"

	"go-admin/internal/common"
)

type SysUser struct {
	common.TenantBaseModel
	Username  string     `gorm:"type:varchar(64);uniqueIndex;comment:用户名" json:"username"`
	Password  string     `gorm:"type:varchar(128);comment:密码" json:"-"`
	Nickname  string     `gorm:"type:varchar(64);comment:昵称" json:"nickname"`
	Email     string     `gorm:"type:varchar(128);comment:邮箱" json:"email"`
	Phone     string     `gorm:"type:varchar(16);comment:手机号" json:"phone"`
	Avatar    string     `gorm:"type:varchar(512);comment:头像" json:"avatar"`
	Status    int8       `gorm:"type:tinyint;default:1;comment:状态 0停用 1正常" json:"status"`
	DeptID    uint       `gorm:"comment:部门ID" json:"deptId"`
	LoginIP   string     `gorm:"type:varchar(128);comment:最后登录IP" json:"loginIp"`
	LoginTime *time.Time `gorm:"comment:最后登录时间" json:"loginTime"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

type SysUserRole struct {
	UserID uint `gorm:"primaryKey;comment:用户ID"`
	RoleID uint `gorm:"primaryKey;comment:角色ID"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

type SysUserPost struct {
	UserID uint `gorm:"primaryKey;comment:用户ID"`
	PostID uint `gorm:"primaryKey;comment:岗位ID"`
}

func (SysUserPost) TableName() string {
	return "sys_user_post"
}
