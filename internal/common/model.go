package common

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreateBy  uint           `gorm:"comment:创建者ID" json:"createBy"`
	UpdateBy  uint           `gorm:"comment:更新者ID" json:"updateBy"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
	Remark    string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
}

type TenantBaseModel struct {
	BaseModel
	TenantID uint `gorm:"index;comment:租户ID" json:"tenantId"`
}
