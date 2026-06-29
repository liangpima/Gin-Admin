package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService service.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{
		roleService: service.NewRoleService(),
	}
}

func (ctl *RoleController) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	operatorID := common.GetCurrentUserID(c)
	if err := ctl.roleService.Create(&req, operatorID); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *RoleController) Update(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	operatorID := common.GetCurrentUserID(c)
	if err := ctl.roleService.Update(&req, operatorID); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *RoleController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	if err := ctl.roleService.Delete(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *RoleController) FindByID(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	role, err := ctl.roleService.FindByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "角色不存在")
		return
	}

	common.Success(c, role)
}

func (ctl *RoleController) FindList(c *gin.Context) {
	var req dto.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	roles, total, err := ctl.roleService.FindList(&req)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessWithPage(c, roles, total, req.Page, req.PageSize)
}

func (ctl *RoleController) UpdateStatus(c *gin.Context) {
	var req dto.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	if err := ctl.roleService.UpdateStatus(&req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (ctl *RoleController) FindAll(c *gin.Context) {
	roles, err := ctl.roleService.FindAll()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, roles)
}
