package service

import (
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

type FileService interface {
	Create(file *model.SysFile) error
	FindByID(id uint) (*model.SysFile, error)
	FindList(name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error)
	Delete(id uint) error
}

type fileService struct {
	fileRepo repository.FileRepository
}

func NewFileService() FileService {
	return &fileService{
		fileRepo: repository.NewFileRepository(),
	}
}

func (s *fileService) Create(file *model.SysFile) error {
	return s.fileRepo.Create(file)
}

func (s *fileService) FindByID(id uint) (*model.SysFile, error) {
	return s.fileRepo.FindByID(id)
}

func (s *fileService) FindList(name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error) {
	return s.fileRepo.FindList(name, mimeType, sortOrder, page, pageSize)
}

func (s *fileService) Delete(id uint) error {
	return s.fileRepo.Delete(id)
}
