package controller

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type DeptController struct {
	deptService service.DeptService
}

func NewDeptController() *DeptController {
	return &DeptController{
		deptService: service.NewDeptService(),
	}
}

func (ctl *DeptController) Create(c *gin.Context) {
	var req dto.CreateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	operatorID := common.GetCurrentUserID(c)
	if err := ctl.deptService.Create(&req, operatorID); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *DeptController) Update(c *gin.Context) {
	var req dto.UpdateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	operatorID := common.GetCurrentUserID(c)
	if err := ctl.deptService.Update(&req, operatorID); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *DeptController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	if err := ctl.deptService.Delete(id); err != nil {
		// 前置条件不满足属于调用方问题，按 400 返回
		if errors.Is(err, service.ErrDeptHasChildren) {
			common.Error(c, common.CodeBadRequest, err.Error())
			return
		}
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *DeptController) FindByID(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	dept, err := ctl.deptService.FindByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "部门不存在")
		return
	}

	common.Success(c, dept)
}

func (ctl *DeptController) FindTree(c *gin.Context) {
	depts, err := ctl.deptService.FindTree()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, depts)
}
