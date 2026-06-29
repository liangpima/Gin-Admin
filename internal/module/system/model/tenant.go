package model

import (
	"time"

	"go-admin/internal/common"
)

type SysTenant struct {
	common.BaseModel
	Name          string     `gorm:"type:varchar(128);comment:租户名称" json:"name"`
	ContactName   string     `gorm:"type:varchar(64);comment:联系人" json:"contactName"`
	ContactPhone  string     `gorm:"type:varchar(16);comment:联系电话" json:"contactPhone"`
	Status        int8       `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	ExpireTime    *time.Time `gorm:"comment:过期时间" json:"expireTime"`
}

func (SysTenant) TableName() string {
	return "sys_tenant"
}
