package model

import (
	"go-admin/internal/common"
)

type SysConfig struct {
	common.BaseModel
	Name     string `gorm:"type:varchar(128);comment:参数名称" json:"name"`
	ConfigKey string `gorm:"type:varchar(200);uniqueIndex;column:config_key;comment:参数键名" json:"key"`
	Value    string `gorm:"type:text;comment:参数键值" json:"value"`
	Type     int8   `gorm:"type:tinyint;default:1;comment:系统内置 0是 1否" json:"type"`
}

func (SysConfig) TableName() string {
	return "sys_config"
}
