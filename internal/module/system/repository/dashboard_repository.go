package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetStats() (*model.DashboardStats, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{db: database.DB}
}

func (r *dashboardRepository) GetStats() (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	r.db.Model(&model.SysUser{}).Count(&stats.UserCount)
	r.db.Model(&model.SysRole{}).Count(&stats.RoleCount)
	r.db.Model(&model.SysMenu{}).Count(&stats.MenuCount)
	r.db.Model(&model.SysDept{}).Count(&stats.DeptCount)
	r.db.Model(&model.SysPost{}).Count(&stats.PostCount)
	r.db.Model(&model.SysConfig{}).Count(&stats.ConfigCount)
	r.db.Model(&model.SysOperationLog{}).Count(&stats.LogCount)

	return stats, nil
}
