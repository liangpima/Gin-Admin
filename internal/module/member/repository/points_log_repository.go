package repository

import (
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/module/member/model"
)

type PointsLogRepository interface {
	Create(tenantID uint, log *model.PointsLog) error
	FindList(tenantID uint, memberID uint, changeType int8, page, pageSize int) ([]model.PointsLog, int64, error)
}

type pointsLogRepository struct{}

func NewPointsLogRepository() PointsLogRepository {
	return &pointsLogRepository{}
}

// Create 创建积分变动记录，自动绑定租户
func (r *pointsLogRepository) Create(tenantID uint, log *model.PointsLog) error {
	log.TenantID = tenantID
	return database.DB.Create(log).Error
}

func (r *pointsLogRepository) FindList(tenantID uint, memberID uint, changeType int8, page, pageSize int) ([]model.PointsLog, int64, error) {
	var logs []model.PointsLog
	var total int64

	query := common.TenantScope(database.DB, tenantID).Model(&model.PointsLog{})
	if memberID > 0 {
		query = query.Where("member_id = ?", memberID)
	}
	if changeType > 0 {
		query = query.Where("type = ?", changeType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
