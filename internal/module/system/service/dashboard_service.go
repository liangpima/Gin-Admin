package service

import (
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

type DashboardService interface {
	GetStats() (*model.DashboardStats, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardService() DashboardService {
	return &dashboardService{
		dashboardRepo: repository.NewDashboardRepository(),
	}
}

func (s *dashboardService) GetStats() (*model.DashboardStats, error) {
	return s.dashboardRepo.GetStats()
}
