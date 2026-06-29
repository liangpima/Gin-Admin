package controller

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type DictController struct {
	dictService service.DictService
}

func NewDictController() *DictController {
	return &DictController{dictService: service.NewDictService()}
}

func (ctl *DictController) CreateType(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.dictService.CreateType(req.Name, req.Type, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *DictController) DeleteType(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.dictService.DeleteType(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *DictController) FindTypeList(c *gin.Context) {
	var req struct {
		Name     string `form:"name"`
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
	}
	c.ShouldBindQuery(&req)
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 10 }
	list, total, err := ctl.dictService.FindTypeList(req.Name, req.Page, req.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

func (ctl *DictController) CreateData(c *gin.Context) {
	var req struct {
		DictType string `json:"dictType" binding:"required"`
		Label    string `json:"label" binding:"required"`
		Value    string `json:"value" binding:"required"`
		Sort     int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.dictService.CreateData(req.DictType, req.Label, req.Value, req.Sort, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *DictController) DeleteData(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.dictService.DeleteData(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *DictController) FindDataList(c *gin.Context) {
	dictType := c.Query("dictType")
	page := 1
	pageSize := 10
	list, total, err := ctl.dictService.FindDataList(dictType, page, pageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, page, pageSize)
}
