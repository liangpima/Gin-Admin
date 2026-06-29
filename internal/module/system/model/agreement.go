package model

import (
	"go-admin/internal/common"
)

type SysAgreement struct {
	common.BaseModel
	Title   string `gorm:"type:varchar(128);comment:标题" json:"title"`
	Content string `gorm:"type:longtext;comment:内容" json:"content"`
	Type    string `gorm:"type:varchar(32);index;comment:类型" json:"type"`
	Sort    int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Status  int8   `gorm:"type:tinyint;default:1;comment:状态 0停用 1正常" json:"status"`
}

func (SysAgreement) TableName() string {
	return "sys_agreement"
}
