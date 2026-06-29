package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	logService service.LogService
}

func NewLogController() *LogController {
	return &LogController{logService: service.NewLogService()}
}

func (ctl *LogController) FindOperationLogList(c *gin.Context) {
	var req struct {
		Title    string `form:"title"`
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
	}
	c.ShouldBindQuery(&req)
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 10 }
	list, total, err := ctl.logService.FindOperationLogList(req.Title, nil, req.Page, req.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

func (ctl *LogController) FindLoginLogList(c *gin.Context) {
	var req struct {
		Username string `form:"username"`
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
	}
	c.ShouldBindQuery(&req)
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 10 }
	list, total, err := ctl.logService.FindLoginLogList(req.Username, nil, req.Page, req.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}
