package service

import (
	"time"

	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

type LogService interface {
	CreateOperationLog(log *model.SysOperationLog) error
	CreateLoginLog(log *model.SysLoginLog) error
	FindOperationLogList(tenantID uint, title string, status *int8, page, pageSize int) ([]interface{}, int64, error)
	FindLoginLogList(tenantID uint, username string, status *int8, page, pageSize int) ([]interface{}, int64, error)
	ClearOperationLogs(tenantID uint) error
	ClearLoginLogs(tenantID uint) error
	// CleanExpiredLogs 按保留天数清理超期日志，供定时任务调用
	CleanExpiredLogs(retentionDays int) (int64, error)
}

type logService struct {
	logRepo repository.LogRepository
}

func NewLogService() LogService {
	return &logService{
		logRepo: repository.NewLogRepository(),
	}
}

func (s *logService) CreateOperationLog(log *model.SysOperationLog) error {
	return s.logRepo.CreateOperationLog(log)
}

func (s *logService) CreateLoginLog(log *model.SysLoginLog) error {
	return s.logRepo.CreateLoginLog(log)
}

func (s *logService) FindOperationLogList(tenantID uint, title string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	logs, total, err := s.logRepo.FindOperationLogList(tenantID, title, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(logs))
	for i, l := range logs {
		result[i] = l
	}
	return result, total, nil
}

func (s *logService) FindLoginLogList(tenantID uint, username string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	logs, total, err := s.logRepo.FindLoginLogList(tenantID, username, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(logs))
	for i, l := range logs {
		result[i] = l
	}
	return result, total, nil
}

func (s *logService) ClearOperationLogs(tenantID uint) error {
	return s.logRepo.ClearOperationLogs(tenantID)
}

func (s *logService) ClearLoginLogs(tenantID uint) error {
	return s.logRepo.ClearLoginLogs(tenantID)
}

// CleanExpiredLogs 清理超过保留期的操作日志与登录日志，返回删除总条数。
// retentionDays <= 0 时不做清理。
func (s *logService) CleanExpiredLogs(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}

	before := time.Now().AddDate(0, 0, -retentionDays)

	opDeleted, err := s.logRepo.DeleteOperationLogsBefore(before)
	if err != nil {
		return 0, err
	}

	loginDeleted, err := s.logRepo.DeleteLoginLogsBefore(before)
	if err != nil {
		return opDeleted, err
	}

	return opDeleted + loginDeleted, nil
}
