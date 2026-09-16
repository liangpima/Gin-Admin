package controller

import (
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
		common.FailWith(c, err)
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
		common.FailWith(c, err)
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

	// service 已用 BizError 标记「存在下级部门」，FailWith 会自动给出 400；
	// 其余（删除失败等）为系统错误，返回 500
	if err := ctl.deptService.Delete(id); err != nil {
		common.FailWith(c, err)
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
		common.FailWith(c, err)
		return
	}

	common.Success(c, dept)
}

func (ctl *DeptController) FindTree(c *gin.Context) {
	depts, err := ctl.deptService.FindTree()
	if err != nil {
		common.FailWith(c, err)
		return
	}

	common.Success(c, depts)
}
