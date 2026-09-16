package service

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

type FileService interface {
	Create(tenantID uint, file *model.SysFile) error
	FindByID(tenantID, id uint) (*model.SysFile, error)
	FindList(tenantID uint, name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error)
	Delete(tenantID, id uint) error
}

type fileService struct {
	fileRepo repository.FileRepository
}

func NewFileService() FileService {
	return &fileService{
		fileRepo: repository.NewFileRepository(),
	}
}

// Create 创建文件记录，自动绑定租户
func (s *fileService) Create(tenantID uint, file *model.SysFile) error {
	file.TenantID = tenantID
	return s.fileRepo.Create(file)
}

func (s *fileService) FindByID(tenantID, id uint) (*model.SysFile, error) {
	file, err := s.fileRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "文件不存在")
	}
	return file, nil
}

func (s *fileService) FindList(tenantID uint, name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error) {
	return s.fileRepo.FindList(tenantID, name, mimeType, sortOrder, page, pageSize)
}

func (s *fileService) Delete(tenantID, id uint) error {
	return s.fileRepo.Delete(tenantID, id)
}
