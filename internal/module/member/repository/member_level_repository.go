package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/member/model"
)

type MemberLevelRepository interface {
	Create(level *model.MemberLevel) error
	Update(level *model.MemberLevel) error
	Delete(id uint) error
	FindByID(id uint) (*model.MemberLevel, error)
	FindList(name string, page, pageSize int) ([]model.MemberLevel, int64, error)
	FindAll() ([]model.MemberLevel, error)
}

type memberLevelRepository struct{}

func NewMemberLevelRepository() MemberLevelRepository {
	return &memberLevelRepository{}
}

func (r *memberLevelRepository) Create(level *model.MemberLevel) error {
	return database.DB.Create(level).Error
}

func (r *memberLevelRepository) Update(level *model.MemberLevel) error {
	return database.DB.Save(level).Error
}

func (r *memberLevelRepository) Delete(id uint) error {
	return database.DB.Delete(&model.MemberLevel{}, id).Error
}

func (r *memberLevelRepository) FindByID(id uint) (*model.MemberLevel, error) {
	var level model.MemberLevel
	err := database.DB.First(&level, id).Error
	return &level, err
}

func (r *memberLevelRepository) FindList(name string, page, pageSize int) ([]model.MemberLevel, int64, error) {
	var levels []model.MemberLevel
	var total int64

	query := database.DB.Model(&model.MemberLevel{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	query.Count(&total)
	err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&levels).Error
	return levels, total, err
}

func (r *memberLevelRepository) FindAll() ([]model.MemberLevel, error) {
	var levels []model.MemberLevel
	err := database.DB.Where("status = 1").Order("sort ASC").Find(&levels).Error
	return levels, err
}
