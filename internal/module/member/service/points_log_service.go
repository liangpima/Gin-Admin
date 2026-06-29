package service

import (
	"go-admin/internal/module/member/dto"
	"go-admin/internal/module/member/model"
	"go-admin/internal/module/member/repository"
)

type PointsLogService interface {
	FindList(req *dto.PointsLogListRequest) ([]model.PointsLog, int64, error)
}

type pointsLogService struct {
	pointsLogRepo repository.PointsLogRepository
}

func NewPointsLogService() PointsLogService {
	return &pointsLogService{
		pointsLogRepo: repository.NewPointsLogRepository(),
	}
}

func (s *pointsLogService) FindList(req *dto.PointsLogListRequest) ([]model.PointsLog, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return s.pointsLogRepo.FindList(req.MemberID, req.Type, req.Page, req.PageSize)
}
