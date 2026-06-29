package common

import "gorm.io/gorm"

// TenantScope 返回附加了 tenant_id 过滤条件的 DB 实例
// 如果 tenantID 为 0，返回原 DB（不过滤，用于通知回调等无需租户隔离的场景）
func TenantScope(db *gorm.DB, tenantID uint) *gorm.DB {
	if tenantID == 0 {
		return db
	}
	return db.Where("tenant_id = ?", tenantID)
}
