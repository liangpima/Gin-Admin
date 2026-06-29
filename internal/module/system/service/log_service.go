package service

import (
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

type LogService interface {
	CreateOperationLog(log *model.SysOperationLog) error
	CreateLoginLog(log *model.SysLoginLog) error
	FindOperationLogList(title string, status *int8, page, pageSize int) ([]interface{}, int64, error)
	FindLoginLogList(username string, status *int8, page, pageSize int) ([]interface{}, int64, error)
	ClearOperationLogs() error
	ClearLoginLogs() error
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

func (s *logService) FindOperationLogList(title string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	logs, total, err := s.logRepo.FindOperationLogList(title, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(logs))
	for i, l := range logs {
		result[i] = l
	}
	return result, total, nil
}

func (s *logService) FindLoginLogList(username string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	logs, total, err := s.logRepo.FindLoginLogList(username, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(logs))
	for i, l := range logs {
		result[i] = l
	}
	return result, total, nil
}

func (s *logService) ClearOperationLogs() error {
	return s.logRepo.ClearOperationLogs()
}

func (s *logService) ClearLoginLogs() error {
	return s.logRepo.ClearLoginLogs()
}
