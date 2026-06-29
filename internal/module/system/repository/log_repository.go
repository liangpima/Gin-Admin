package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type LogRepository interface {
	CreateOperationLog(log *model.SysOperationLog) error
	CreateLoginLog(log *model.SysLoginLog) error
	FindOperationLogList(title string, status *int8, page, pageSize int) ([]model.SysOperationLog, int64, error)
	FindLoginLogList(username string, status *int8, page, pageSize int) ([]model.SysLoginLog, int64, error)
	ClearOperationLogs() error
	ClearLoginLogs() error
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository() LogRepository {
	return &logRepository{db: database.DB}
}

func (r *logRepository) CreateOperationLog(log *model.SysOperationLog) error {
	return r.db.Create(log).Error
}

func (r *logRepository) CreateLoginLog(log *model.SysLoginLog) error {
	return r.db.Create(log).Error
}

func (r *logRepository) FindOperationLogList(title string, status *int8, page, pageSize int) ([]model.SysOperationLog, int64, error) {
	var logs []model.SysOperationLog
	var total int64

	query := r.db.Model(&model.SysOperationLog{})
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	return logs, total, err
}

func (r *logRepository) FindLoginLogList(username string, status *int8, page, pageSize int) ([]model.SysLoginLog, int64, error) {
	var logs []model.SysLoginLog
	var total int64

	query := r.db.Model(&model.SysLoginLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	return logs, total, err
}

func (r *logRepository) ClearOperationLogs() error {
	return r.db.Where("1 = 1").Delete(&model.SysOperationLog{}).Error
}

func (r *logRepository) ClearLoginLogs() error {
	return r.db.Where("1 = 1").Delete(&model.SysLoginLog{}).Error
}
