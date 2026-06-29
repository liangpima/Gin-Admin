package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService service.PostService
}

func NewPostController() *PostController {
	return &PostController{postService: service.NewPostService()}
}

func (ctl *PostController) Create(c *gin.Context) {
	var req struct {
		Code   string `json:"code" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Sort   int    `json:"sort"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.postService.Create(req.Name, req.Code, req.Sort, req.Status, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *PostController) Update(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Code   string `json:"code" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Sort   int    `json:"sort"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.postService.Update(req.ID, req.Name, req.Code, req.Sort, req.Status, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *PostController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.postService.Delete(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *PostController) FindList(c *gin.Context) {
	var req struct {
		Name     string `form:"name"`
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
	}
	c.ShouldBindQuery(&req)
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 10 }
	list, total, err := ctl.postService.FindList(req.Name, nil, req.Page, req.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}
