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
	// 注意：不能用 `existing, _ := ...; if existing != nil` 判断重名 ——
	// FindTypeByType 无论查没查到都返回非 nil 指针（&dictType, err），
	// 该条件恒为真，会让新建字典类型永远报「已存在」。
	// 判定依据只能是 err：nil 表示查到，ErrRecordNotFound 表示可以创建。
	existing, err := s.dictRepo.FindTypeByType(typ)
	if err == nil {
		if existing != nil && existing.ID > 0 {
			return common.NewBizError("字典类型已存在")
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
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

	if err := s.dictRepo.CreateType(dictType); err != nil {
		// Count 校验有时间窗口，并发下靠全局唯一的 uk_type 兜底
		if errors.Is(err, common.ErrDuplicateKey) {
			return common.NewBizError("字典类型已存在")
		}
		return err
	}
	return nil
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
			return common.NewNotFoundError("字典类型不存在")
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
