package controller

import (
	"strconv"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/member/dto"
	"go-admin/internal/module/member/repository"
	"go-admin/internal/module/member/service"
	systemModel "go-admin/internal/module/system/model"
	systemService "go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

type MemberController struct {
	memberService service.MemberService
	levelRepo     repository.MemberLevelRepository
	tagRepo       repository.MemberTagRepository
}

func NewMemberController() *MemberController {
	return &MemberController{
		memberService: service.NewMemberService(),
		levelRepo:     repository.NewMemberLevelRepository(),
		tagRepo:       repository.NewMemberTagRepository(),
	}
}

func (ctl *MemberController) Create(c *gin.Context) {
	var req dto.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	operatorID := common.GetCurrentUserID(c)
	tenantID := common.GetTenantID(c)

	digits := 6
	configService := systemService.NewConfigService()
	if configVal, err := configService.FindByKey("site.memberIdDigits"); err == nil {
		if cfg, ok := configVal.(*systemModel.SysConfig); ok && cfg.Value != "" {
			if n, err := strconv.Atoi(cfg.Value); err == nil && n >= 4 {
				digits = n
			}
		}
	}

	if err := ctl.memberService.Create(&req, operatorID, tenantID, digits); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *MemberController) Update(c *gin.Context) {
	var req dto.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	operatorID := common.GetCurrentUserID(c)
	tenantID := common.GetTenantID(c)
	if err := ctl.memberService.Update(&req, operatorID, tenantID); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *MemberController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	tenantID := common.GetTenantID(c)
	if err := ctl.memberService.Delete(tenantID, id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *MemberController) FindByID(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	tenantID := common.GetTenantID(c)
	member, err := ctl.memberService.FindByID(tenantID, id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "会员不存在")
		return
	}
	common.Success(c, member)
}

func (ctl *MemberController) FindList(c *gin.Context) {
	var req dto.MemberListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	tenantID := common.GetTenantID(c)
	list, total, err := ctl.memberService.FindList(tenantID, &req)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

func (ctl *MemberController) UpdateStatus(c *gin.Context) {
	var req dto.UpdateMemberStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	tenantID := common.GetTenantID(c)
	if err := ctl.memberService.UpdateStatus(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *MemberController) UpdateTags(c *gin.Context) {
	var req dto.UpdateMemberTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	tenantID := common.GetTenantID(c)
	if err := ctl.memberService.UpdateTags(tenantID, &req); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *MemberController) FindAllLevels(c *gin.Context) {
	levels, err := ctl.levelRepo.FindAll()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, levels)
}

func (ctl *MemberController) FindAllTags(c *gin.Context) {
	tags, err := ctl.tagRepo.FindAll()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, tags)
}

func (ctl *MemberController) UpdateLastVisit(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	tenantID := common.GetTenantID(c)
	member, err := ctl.memberService.FindByID(tenantID, req.ID)
	if err != nil {
		common.Error(c, common.CodeNotFound, "会员不存在")
		return
	}
	now := time.Now()
	member.LastVisitTime = &now
	_ = repository.NewMemberRepository().Update(member)
	common.Success(c, nil)
}
