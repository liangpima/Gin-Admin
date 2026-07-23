package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

// @Summary 创建用户
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.CreateUserRequest true "用户信息"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user [post]
func (ctl *UserController) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	operatorID := common.GetCurrentUserID(c)
	if err := ctl.userService.Create(tenantID, &req, operatorID); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 更新用户
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.UpdateUserRequest true "用户信息"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user [put]
func (ctl *UserController) Update(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	operatorID := common.GetCurrentUserID(c)
	if err := ctl.userService.Update(tenantID, &req, operatorID); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 删除用户
// @Tags 管理员
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/{id} [delete]
func (ctl *UserController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	tenantID := common.GetTenantID(c)
	if err := ctl.userService.Delete(tenantID, id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 获取用户详情
// @Tags 管理员
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/{id} [get]
func (ctl *UserController) FindByID(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	tenantID := common.GetTenantID(c)
	user, err := ctl.userService.FindByID(tenantID, id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "用户不存在")
		return
	}

	common.Success(c, user)
}

// @Summary 用户列表
// @Tags 管理员
// @Produce json
// @Param username query string false "用户名"
// @Param phone query string false "手机号"
// @Param status query int false "状态"
// @Param deptId query int false "部门ID"
// @Param page query int true "页码"
// @Param pageSize query int true "每页条数"
// @Success 200 {object} common.Response{data=common.PageData}
// @Router /api/v1/system/user/list [get]
func (ctl *UserController) FindList(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	users, total, err := ctl.userService.FindList(tenantID, &req)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessWithPage(c, users, total, req.Page, req.PageSize)
}

// @Summary 修改用户状态
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.StatusRequest true "状态"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/status [put]
func (ctl *UserController) UpdateStatus(c *gin.Context) {
	var req dto.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	if err := ctl.userService.UpdateStatus(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 重置密码
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.ResetPasswordRequest true "新密码"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/resetPwd [put]
func (ctl *UserController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	operatorID := common.GetCurrentUserID(c)
	if req.ID != operatorID {
		common.Forbidden(c, "无权重置其他用户密码")
		return
	}

	if err := ctl.userService.ResetPassword(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 修改用户角色
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.UpdateUserRolesRequest true "角色"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/roles [put]
func (ctl *UserController) UpdateRoles(c *gin.Context) {
	var req dto.UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	if err := ctl.userService.UpdateRoles(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 修改用户部门
// @Tags 管理员
// @Accept json
// @Produce json
// @Param body body dto.UpdateUserDeptRequest true "部门"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/dept [put]
func (ctl *UserController) UpdateDept(c *gin.Context) {
	var req dto.UpdateUserDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	if err := ctl.userService.UpdateDept(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

// @Summary 修改密码
// @Tags 管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} common.Response
// @Router /api/v1/system/user/changePwd [put]
func (ctl *UserController) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	userID := common.GetCurrentUserID(c)
	if err := ctl.userService.ChangePassword(userID, &req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	common.Success(c, nil)
}
