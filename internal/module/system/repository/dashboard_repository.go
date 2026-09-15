package repository

import (
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetStats(tenantID uint) (*model.DashboardStats, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{db: database.DB}
}

// GetStats 汇总仪表盘统计数据。
//
// sys_user / sys_role / sys_operation_log 属于多租户表，必须按租户过滤，
// 否则会把其他租户的数量也统计进来；
// sys_menu / sys_dept / sys_post / sys_config 为全局表，不做租户过滤。
func (r *dashboardRepository) GetStats(tenantID uint) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	counters := []struct {
		query *gorm.DB
		dest  *int64
	}{
		{common.TenantScope(r.db.Model(&model.SysUser{}), tenantID), &stats.UserCount},
		{common.TenantScope(r.db.Model(&model.SysRole{}), tenantID), &stats.RoleCount},
		{r.db.Model(&model.SysMenu{}), &stats.MenuCount},
		{r.db.Model(&model.SysDept{}), &stats.DeptCount},
		{r.db.Model(&model.SysPost{}), &stats.PostCount},
		{r.db.Model(&model.SysConfig{}), &stats.ConfigCount},
		{common.TenantScope(r.db.Model(&model.SysOperationLog{}), tenantID), &stats.LogCount},
	}

	for _, c := range counters {
		if err := c.query.Count(c.dest).Error; err != nil {
			return nil, err
		}
	}

	return stats, nil
}
