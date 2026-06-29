package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type DictRepository interface {
	CreateType(dictType *model.SysDictType) error
	FindTypeByID(id uint) (*model.SysDictType, error)
	FindTypeByType(typ string) (*model.SysDictType, error)
	FindTypeList(name string, page, pageSize int) ([]model.SysDictType, int64, error)
	UpdateType(dictType *model.SysDictType) error
	DeleteType(id uint) error

	CreateData(dictData *model.SysDictData) error
	FindDataByID(id uint) (*model.SysDictData, error)
	FindDataByType(typ string) ([]model.SysDictData, error)
	FindDataList(typ string, page, pageSize int) ([]model.SysDictData, int64, error)
	UpdateData(dictData *model.SysDictData) error
	DeleteData(id uint) error
}

type dictRepository struct {
	db *gorm.DB
}

func NewDictRepository() DictRepository {
	return &dictRepository{db: database.DB}
}

func (r *dictRepository) CreateType(dictType *model.SysDictType) error {
	return r.db.Create(dictType).Error
}

func (r *dictRepository) FindTypeByID(id uint) (*model.SysDictType, error) {
	var dictType model.SysDictType
	err := r.db.First(&dictType, id).Error
	return &dictType, err
}

func (r *dictRepository) FindTypeByType(typ string) (*model.SysDictType, error) {
	var dictType model.SysDictType
	err := r.db.Where("type = ?", typ).First(&dictType).Error
	return &dictType, err
}

func (r *dictRepository) FindTypeList(name string, page, pageSize int) ([]model.SysDictType, int64, error) {
	var dictTypes []model.SysDictType
	var total int64

	query := r.db.Model(&model.SysDictType{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&dictTypes).Error
	return dictTypes, total, err
}

func (r *dictRepository) UpdateType(dictType *model.SysDictType) error {
	return r.db.Model(dictType).Select("Name", "Type", "Status", "Remark", "UpdateBy").Updates(dictType).Error
}

func (r *dictRepository) DeleteType(id uint) error {
	return r.db.Delete(&model.SysDictType{}, id).Error
}

func (r *dictRepository) CreateData(dictData *model.SysDictData) error {
	return r.db.Create(dictData).Error
}

func (r *dictRepository) FindDataByID(id uint) (*model.SysDictData, error) {
	var dictData model.SysDictData
	err := r.db.First(&dictData, id).Error
	return &dictData, err
}

func (r *dictRepository) FindDataByType(typ string) ([]model.SysDictData, error) {
	var dictData []model.SysDictData
	err := r.db.Where("dict_type = ? AND status = ?", typ, 1).Order("sort ASC, id ASC").Find(&dictData).Error
	return dictData, err
}

func (r *dictRepository) FindDataList(typ string, page, pageSize int) ([]model.SysDictData, int64, error) {
	var dictData []model.SysDictData
	var total int64

	query := r.db.Model(&model.SysDictData{})
	if typ != "" {
		query = query.Where("dict_type = ?", typ)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&dictData).Error
	return dictData, total, err
}

func (r *dictRepository) UpdateData(dictData *model.SysDictData) error {
	return r.db.Model(dictData).Select("DictType", "Label", "Value", "Sort", "CssClass", "ListClass", "Status", "Remark", "UpdateBy").Updates(dictData).Error
}

func (r *dictRepository) DeleteData(id uint) error {
	return r.db.Delete(&model.SysDictData{}, id).Error
}
