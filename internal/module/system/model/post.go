package model

import (
	"go-admin/internal/common"
)

type SysPost struct {
	common.BaseModel
	Code   string `gorm:"type:varchar(64);uniqueIndex;comment:岗位编码" json:"code"`
	Name   string `gorm:"type:varchar(64);comment:岗位名称" json:"name"`
	Sort   int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Status int8   `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
}

func (SysPost) TableName() string {
	return "sys_post"
}
