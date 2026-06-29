package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *model.SysPost) error
	FindByID(id uint) (*model.SysPost, error)
	FindAll() ([]model.SysPost, error)
	FindList(name string, status *int8, page, pageSize int) ([]model.SysPost, int64, error)
	Update(post *model.SysPost) error
	Delete(id uint) error
	CountByCode(code string, excludeID uint) int64
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository() PostRepository {
	return &postRepository{db: database.DB}
}

func (r *postRepository) Create(post *model.SysPost) error {
	return r.db.Create(post).Error
}

func (r *postRepository) FindByID(id uint) (*model.SysPost, error) {
	var post model.SysPost
	err := r.db.First(&post, id).Error
	return &post, err
}

func (r *postRepository) FindAll() ([]model.SysPost, error) {
	var posts []model.SysPost
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&posts).Error
	return posts, err
}

func (r *postRepository) FindList(name string, status *int8, page, pageSize int) ([]model.SysPost, int64, error) {
	var posts []model.SysPost
	var total int64

	query := r.db.Model(&model.SysPost{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&posts).Error
	return posts, total, err
}

func (r *postRepository) Update(post *model.SysPost) error {
	return r.db.Model(post).Select("Code", "Name", "Sort", "Status", "Remark", "UpdateBy").Updates(post).Error
}

func (r *postRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysPost{}, id).Error
}

func (r *postRepository) CountByCode(code string, excludeID uint) int64 {
	var count int64
	query := r.db.Model(&model.SysPost{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	query.Count(&count)
	return count
}
