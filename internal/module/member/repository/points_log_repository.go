package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/member/model"
)

type PointsLogRepository interface {
	Create(log *model.PointsLog) error
	FindList(memberID uint, changeType int8, page, pageSize int) ([]model.PointsLog, int64, error)
}

type pointsLogRepository struct{}

func NewPointsLogRepository() PointsLogRepository {
	return &pointsLogRepository{}
}

func (r *pointsLogRepository) Create(log *model.PointsLog) error {
	return database.DB.Create(log).Error
}

func (r *pointsLogRepository) FindList(memberID uint, changeType int8, page, pageSize int) ([]model.PointsLog, int64, error) {
	var logs []model.PointsLog
	var total int64

	query := database.DB.Model(&model.PointsLog{})
	if memberID > 0 {
		query = query.Where("member_id = ?", memberID)
	}
	if changeType > 0 {
		query = query.Where("type = ?", changeType)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
