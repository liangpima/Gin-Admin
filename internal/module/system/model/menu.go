package model

import (
	"go-admin/internal/common"
)

type SysMenu struct {
	common.BaseModel
	ParentID   uint      `gorm:"default:0;comment:父菜单ID" json:"parentId"`
	Name       string    `gorm:"type:varchar(64);comment:菜单名称" json:"name"`
	Path       string    `gorm:"type:varchar(200);comment:路由地址" json:"path"`
	Component  string    `gorm:"type:varchar(200);comment:组件路径" json:"component"`
	Redirect   string    `gorm:"type:varchar(200);comment:重定向" json:"redirect"`
	Icon       string    `gorm:"type:varchar(64);comment:图标" json:"icon"`
	Title      string    `gorm:"type:varchar(64);comment:标题" json:"title"`
	Type       int8      `gorm:"type:tinyint;comment:类型 0目录 1菜单 2按钮" json:"type"`
	Permission string    `gorm:"type:varchar(200);comment:权限标识" json:"permission"`
	Sort       int       `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Visible    int8      `gorm:"type:tinyint;default:1;comment:是否可见" json:"visible"`
	Status     int8      `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	IsExternal int8      `gorm:"type:tinyint;default:0;comment:是否外链" json:"isExternal"`
	IsCache    int8      `gorm:"type:tinyint;default:1;comment:是否缓存" json:"isCache"`
	Children   []SysMenu `gorm:"-" json:"children"`
}

func (SysMenu) TableName() string {
	return "sys_menu"
}
