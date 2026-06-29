package model

import (
	"go-admin/internal/common"
)

type SysRole struct {
	common.TenantBaseModel
	Name      string    `gorm:"type:varchar(64);comment:角色名称" json:"name"`
	Code      string    `gorm:"type:varchar(64);uniqueIndex;comment:角色编码" json:"code"`
	Sort      int       `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Status    int8      `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	DataScope int8      `gorm:"type:tinyint;default:1;comment:数据权限范围" json:"dataScope"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type SysRoleMenu struct {
	RoleID uint `gorm:"primaryKey;comment:角色ID"`
	MenuID uint `gorm:"primaryKey;comment:菜单ID"`
}

func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}
