package model

import (
	"go-admin/internal/common"
)

type SysDictType struct {
	common.BaseModel
	Name   string `gorm:"type:varchar(128);comment:字典名称" json:"name"`
	Type   string `gorm:"type:varchar(128);uniqueIndex;comment:字典类型" json:"type"`
	Status int8   `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
}

func (SysDictType) TableName() string {
	return "sys_dict_type"
}

type SysDictData struct {
	common.BaseModel
	DictType string `gorm:"type:varchar(128);index;comment:字典类型" json:"dictType"`
	Label    string `gorm:"type:varchar(128);comment:字典标签" json:"label"`
	Value    string `gorm:"type:varchar(128);comment:字典键值" json:"value"`
	Sort     int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
	CssClass string `gorm:"type:varchar(128);comment:样式属性" json:"cssClass"`
	ListClass string `gorm:"type:varchar(128);comment:表格回显样式" json:"listClass"`
	Status   int8   `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
}

func (SysDictData) TableName() string {
	return "sys_dict_data"
}
