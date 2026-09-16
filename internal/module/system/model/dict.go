package model

import (
	"go-admin/internal/common"
)

type SysDictType struct {
	common.BaseModel
	Name   string `gorm:"type:varchar(128);comment:字典名称" json:"name"`
	Type   string `gorm:"type:varchar(128);uniqueIndex;comment:字典类型" json:"type"`
	Status int8   `gorm:"type:tinyint;comment:状态" json:"status"`
}

func (SysDictType) TableName() string {
	return "sys_dict_type"
}

type SysDictData struct {
	common.BaseModel
	// DictType + Value 组成复合唯一索引：同一类型下键值不允许重复。
	// 没有它的话，同一个 value 可以建出多个 label，前端按下拉取值时
	// 回显哪一条完全取决于查询顺序，属于随机行为。
	// 软删除时 DeleteData 会改写 value 释放该组合，故删除后同名可重建。
	DictType  string `gorm:"type:varchar(128);uniqueIndex:uk_dict_type_value;comment:字典类型" json:"dictType"`
	Label     string `gorm:"type:varchar(128);comment:字典标签" json:"label"`
	Value     string `gorm:"type:varchar(128);uniqueIndex:uk_dict_type_value;comment:字典键值" json:"value"`
	Sort      int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
	CssClass  string `gorm:"type:varchar(128);comment:样式属性" json:"cssClass"`
	ListClass string `gorm:"type:varchar(128);comment:表格回显样式" json:"listClass"`
	Status    int8   `gorm:"type:tinyint;comment:状态" json:"status"`
}

func (SysDictData) TableName() string {
	return "sys_dict_data"
}
