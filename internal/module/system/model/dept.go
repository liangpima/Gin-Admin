package model

import (
	"go-admin/internal/common"
)

type SysDept struct {
	common.BaseModel
	ParentID uint      `gorm:"default:0;comment:父部门ID" json:"parentId"`
	Name     string    `gorm:"type:varchar(64);comment:部门名称" json:"name"`
	Sort     int       `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Leader   string    `gorm:"type:varchar(64);comment:负责人" json:"leader"`
	Phone    string    `gorm:"type:varchar(16);comment:联系电话" json:"phone"`
	Email    string    `gorm:"type:varchar(128);comment:邮箱" json:"email"`
	Status   int8      `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	Children []SysDept `gorm:"-" json:"children"`
}

func (SysDept) TableName() string {
	return "sys_dept"
}
