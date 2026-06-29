package model

import (
	"go-admin/internal/common"
)

type SysFile struct {
	common.TenantBaseModel
	Name        string `gorm:"type:varchar(200);comment:原始文件名" json:"name"`
	StorageName string `gorm:"type:varchar(200);comment:存储文件名" json:"storageName"`
	Path        string `gorm:"type:varchar(500);comment:文件路径" json:"path"`
	URL         string `gorm:"type:varchar(500);comment:访问URL" json:"url"`
	Size        int64  `gorm:"comment:文件大小" json:"size"`
	MimeType    string `gorm:"type:varchar(128);comment:MIME类型" json:"mimeType"`
	StorageType int8   `gorm:"type:tinyint;default:0;comment:存储类型 0本地 1OSS" json:"storageType"`
}

func (SysFile) TableName() string {
	return "sys_file"
}
