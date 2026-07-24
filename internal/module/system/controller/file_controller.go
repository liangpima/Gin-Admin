package controller

import (
	"net/url"

	"go-admin/config"
	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/service"
	"go-admin/pkg/upload"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService service.FileService
}

func NewFileController() *FileController {
	return &FileController{fileService: service.NewFileService()}
}

// @Summary 上传文件
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "文件"
// @Success 200 {object} common.Response
// @Router /system/file/upload [post]
func (ctl *FileController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "请选择文件")
		return
	}

	maxSize := int64(config.Cfg.Upload.MaxSize) * 1024 * 1024
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 默认 10MB
	}
	if file.Size > maxSize {
		common.Error(c, common.CodeFileTooLarge, "文件大小不能超过10MB")
		return
	}

	storagePath, err := upload.Upload(file)
	if err != nil {
		common.Error(c, common.CodeUploadFailed, "上传失败: "+err.Error())
		return
	}

	fileURL := upload.GetURL(storagePath)

	dbFile := &model.SysFile{
		Name:     file.Filename,
		Path:     storagePath,
		URL:      fileURL,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}
	dbFile.CreateBy = common.GetCurrentUserID(c)
	dbFile.UpdateBy = common.GetCurrentUserID(c)

	if err := ctl.fileService.Create(dbFile); err != nil {
		common.Error(c, common.CodeInternalError, "保存记录失败")
		return
	}

	common.Success(c, gin.H{
		"id":   dbFile.ID,
		"name": dbFile.Name,
		"url":  dbFile.URL,
		"size": dbFile.Size,
	})
}

// @Summary 文件列表
// @Tags 文件
// @Produce json
// @Security BearerAuth
// @Param name query string false "文件名"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} common.Response
// @Router /system/file/list [get]
func (ctl *FileController) FindList(c *gin.Context) {
	name := c.Query("name")
	mimeType := c.Query("mimeType")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	page, pageSize := common.GetPageInfo(c)

	list, total, err := ctl.fileService.FindList(name, mimeType, sortOrder, page, pageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.SuccessWithPage(c, list, total, page, pageSize)
}

// @Summary 获取文件详情
// @Tags 文件
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} common.Response
// @Router /system/file/{id} [get]
func (ctl *FileController) FindByID(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	file, err := ctl.fileService.FindByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "文件不存在")
		return
	}
	common.Success(c, file)
}

// @Summary 删除文件
// @Tags 文件
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} common.Response
// @Router /system/file/{id} [delete]
func (ctl *FileController) Delete(c *gin.Context) {
	id, err := common.GetUintParam(c, "id")
	if err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}
	file, err := ctl.fileService.FindByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "文件不存在")
		return
	}

	_ = upload.Delete(file.Path)

	if err := ctl.fileService.Delete(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, nil)
}

func encodeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.String()
}
