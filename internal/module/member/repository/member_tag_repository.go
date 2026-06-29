package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/member/model"
)

type MemberTagRepository interface {
	Create(tag *model.MemberTag) error
	Update(tag *model.MemberTag) error
	Delete(id uint) error
	FindByID(id uint) (*model.MemberTag, error)
	FindList(name string, page, pageSize int) ([]model.MemberTag, int64, error)
	FindAll() ([]model.MemberTag, error)
}

type memberTagRepository struct{}

func NewMemberTagRepository() MemberTagRepository {
	return &memberTagRepository{}
}

func (r *memberTagRepository) Create(tag *model.MemberTag) error {
	return database.DB.Create(tag).Error
}

func (r *memberTagRepository) Update(tag *model.MemberTag) error {
	return database.DB.Save(tag).Error
}

func (r *memberTagRepository) Delete(id uint) error {
	return database.DB.Delete(&model.MemberTag{}, id).Error
}

func (r *memberTagRepository) FindByID(id uint) (*model.MemberTag, error) {
	var tag model.MemberTag
	err := database.DB.First(&tag, id).Error
	return &tag, err
}

func (r *memberTagRepository) FindList(name string, page, pageSize int) ([]model.MemberTag, int64, error) {
	var tags []model.MemberTag
	var total int64

	query := database.DB.Model(&model.MemberTag{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	query.Count(&total)
	err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tags).Error
	return tags, total, err
}

func (r *memberTagRepository) FindAll() ([]model.MemberTag, error) {
	var tags []model.MemberTag
	err := database.DB.Where("status = 1").Order("sort ASC").Find(&tags).Error
	return tags, err
}
