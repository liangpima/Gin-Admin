package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/service"
	"go-admin/pkg/upload"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	configService service.ConfigService
}

func NewConfigController() *ConfigController {
	return &ConfigController{configService: service.NewConfigService()}
}

func (ctl *ConfigController) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
		Type  int8   `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.configService.Create(req.Name, req.Key, req.Value, req.Type, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *ConfigController) Update(c *gin.Context) {
	var req struct {
		ID    uint   `json:"id" binding:"required"`
		Name  string `json:"name" binding:"required"`
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
		Type  int8   `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	if err := ctl.configService.Update(req.ID, req.Name, req.Key, req.Value, req.Type, common.GetCurrentUserID(c)); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *ConfigController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.configService.Delete(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func (ctl *ConfigController) FindList(c *gin.Context) {
	var req struct {
		Name     string `form:"name"`
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
	}
	c.ShouldBindQuery(&req)
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 10 }
	list, total, err := ctl.configService.FindList(req.Name, req.Page, req.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, req.Page, req.PageSize)
}

func (ctl *ConfigController) SiteInfo(c *gin.Context) {
	list, err := ctl.configService.FindByPrefix("site.")
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	result := map[string]string{}
	for _, item := range list {
		switch cfg := item.(type) {
		case model.SysConfig:
			result[cfg.Key] = cfg.Value
		case *model.SysConfig:
			result[cfg.Key] = cfg.Value
		}
	}
	common.Success(c, result)
}

func (ctl *ConfigController) FindByPrefix(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		common.Error(c, common.CodeBadRequest, "prefix不能为空")
		return
	}
	list, err := ctl.configService.FindByPrefix(prefix)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, list)
}

func (ctl *ConfigController) BatchSave(c *gin.Context) {
	var req struct {
		Prefix string `json:"prefix" binding:"required"`
		Items  []struct {
			Key   string `json:"key" binding:"required"`
			Value string `json:"value"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}
	operatorID := common.GetCurrentUserID(c)
	items := make([]service.ConfigItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.ConfigItem{Key: item.Key, Value: item.Value}
	}
	if err := ctl.configService.BatchSave(req.Prefix, items, operatorID); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	if req.Prefix == "oss." {
		upload.Reload()
	}
	common.Success(c, nil)
}

func (ctl *ConfigController) UploadCert(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "请选择文件")
		return
	}

	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".pem": true, ".key": true, ".crt": true, ".cer": true}
	if !allowedExts[ext] {
		common.Error(c, common.CodeBadRequest, "仅支持 .pem/.key/.crt/.cer 文件")
		return
	}

	if file.Size > 2*1024*1024 {
		common.Error(c, common.CodeBadRequest, "文件大小不能超过 2MB")
		return
	}

	saveDir := filepath.Join("uploads", "certs")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		common.Error(c, common.CodeInternalError, "创建目录失败")
		return
	}

	filename := fmt.Sprintf("wechat_%d%s", time.Now().UnixMilli(), ext)
	savePath := filepath.Join(saveDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		common.Error(c, common.CodeInternalError, "保存文件失败")
		return
	}

	common.Success(c, gin.H{
		"path":     savePath,
		"filename": file.Filename,
	})
}
