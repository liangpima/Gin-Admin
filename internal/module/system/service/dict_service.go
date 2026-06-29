package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type DictService interface {
	CreateType(name, typ string, operatorID uint) error
	FindTypeList(name string, page, pageSize int) ([]interface{}, int64, error)
	FindTypeByID(id uint) (interface{}, error)
	UpdateType(id uint, name string, operatorID uint) error
	DeleteType(id uint) error
	CreateData(dictType, label, value string, sort int, operatorID uint) error
	FindDataByType(typ string) ([]model.SysDictData, error)
	FindDataList(typ string, page, pageSize int) ([]interface{}, int64, error)
	DeleteData(id uint) error
}

type dictService struct {
	dictRepo repository.DictRepository
}

func NewDictService() DictService {
	return &dictService{
		dictRepo: repository.NewDictRepository(),
	}
}

func (s *dictService) CreateType(name, typ string, operatorID uint) error {
	existing, _ := s.dictRepo.FindTypeByType(typ)
	if existing != nil {
		return errors.New("字典类型已存在")
	}

	dictType := &model.SysDictType{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		Name:   name,
		Type:   typ,
		Status: 1,
	}

	return s.dictRepo.CreateType(dictType)
}

func (s *dictService) FindTypeList(name string, page, pageSize int) ([]interface{}, int64, error) {
	types, total, err := s.dictRepo.FindTypeList(name, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(types))
	for i, t := range types {
		result[i] = t
	}
	return result, total, nil
}

func (s *dictService) FindTypeByID(id uint) (interface{}, error) {
	return s.dictRepo.FindTypeByID(id)
}

func (s *dictService) UpdateType(id uint, name string, operatorID uint) error {
	dictType, err := s.dictRepo.FindTypeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字典类型不存在")
		}
		return err
	}
	dictType.Name = name
	dictType.UpdateBy = operatorID
	return s.dictRepo.UpdateType(dictType)
}

func (s *dictService) DeleteType(id uint) error {
	return s.dictRepo.DeleteType(id)
}

func (s *dictService) CreateData(dictType, label, value string, sort int, operatorID uint) error {
	dictData := &model.SysDictData{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		DictType: dictType,
		Label:    label,
		Value:    value,
		Sort:     sort,
		Status:   1,
	}

	return s.dictRepo.CreateData(dictData)
}

func (s *dictService) FindDataByType(typ string) ([]model.SysDictData, error) {
	return s.dictRepo.FindDataByType(typ)
}

func (s *dictService) FindDataList(typ string, page, pageSize int) ([]interface{}, int64, error) {
	data, total, err := s.dictRepo.FindDataList(typ, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(data))
	for i, d := range data {
		result[i] = d
	}
	return result, total, nil
}

func (s *dictService) DeleteData(id uint) error {
	return s.dictRepo.DeleteData(id)
}
