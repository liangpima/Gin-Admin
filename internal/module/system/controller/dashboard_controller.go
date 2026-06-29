package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{
		dashboardService: service.NewDashboardService(),
	}
}

// @Summary 获取统计数据
// @Tags 仪表盘
// @Produce json
// @Success 200 {object} common.Response{data=model.DashboardStats}
// @Router /api/v1/dashboard/stats [get]
func (ctl *DashboardController) GetStats(c *gin.Context) {
	stats, err := ctl.dashboardService.GetStats()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, stats)
}
