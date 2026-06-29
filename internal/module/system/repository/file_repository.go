package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *model.SysFile) error
	FindByID(id uint) (*model.SysFile, error)
	FindList(name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error)
	Delete(id uint) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository() FileRepository {
	return &fileRepository{db: database.DB}
}

func (r *fileRepository) Create(file *model.SysFile) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) FindByID(id uint) (*model.SysFile, error) {
	var file model.SysFile
	err := r.db.First(&file, id).Error
	return &file, err
}

func (r *fileRepository) FindList(name, mimeType, sortOrder string, page, pageSize int) ([]model.SysFile, int64, error) {
	var files []model.SysFile
	var total int64

	query := r.db.Model(&model.SysFile{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if mimeType != "" {
		query = query.Where("mime_type LIKE ?", mimeType+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "id DESC"
	if sortOrder == "asc" {
		order = "id ASC"
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(order).Find(&files).Error
	return files, total, err
}

func (r *fileRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysFile{}, id).Error
}
