package service

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"errors"

	"gorm.io/gorm"
)

type ConfigItem struct {
	Key   string
	Value string
}

type ConfigService interface {
	Create(name, key, value string, typ int8, operatorID uint) error
	Update(id uint, name, key, value string, typ int8, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (interface{}, error)
	FindByKey(key string) (interface{}, error)
	FindList(name string, page, pageSize int) ([]interface{}, int64, error)
	FindByPrefix(prefix string) ([]interface{}, error)
	BatchSave(prefix string, items []ConfigItem, operatorID uint) error
}

type configService struct {
	configRepo repository.ConfigRepository
}

func NewConfigService() ConfigService {
	return &configService{
		configRepo: repository.NewConfigRepository(),
	}
}

func (s *configService) Create(name, key, value string, typ int8, operatorID uint) error {
	config := &model.SysConfig{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		Name:  name,
		Key:   key,
		Value: value,
		Type:  typ,
	}

	return s.configRepo.Create(config)
}

func (s *configService) Update(id uint, name, key, value string, typ int8, operatorID uint) error {
	config, err := s.configRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("配置不存在")
		}
		return err
	}

	config.Name = name
	config.Key = key
	config.Value = value
	config.Type = typ
	config.UpdateBy = operatorID

	return s.configRepo.Update(config)
}

func (s *configService) Delete(id uint) error {
	return s.configRepo.Delete(id)
}

func (s *configService) FindByID(id uint) (interface{}, error) {
	return s.configRepo.FindByID(id)
}

func (s *configService) FindByKey(key string) (interface{}, error) {
	return s.configRepo.FindByKey(key)
}

func (s *configService) FindList(name string, page, pageSize int) ([]interface{}, int64, error) {
	configs, total, err := s.configRepo.FindList(name, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(configs))
	for i, c := range configs {
		result[i] = c
	}
	return result, total, nil
}

func (s *configService) FindByPrefix(prefix string) ([]interface{}, error) {
	configs, err := s.configRepo.FindByKeyPrefix(prefix)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(configs))
	for i, c := range configs {
		result[i] = c
	}
	return result, nil
}

func (s *configService) BatchSave(prefix string, items []ConfigItem, operatorID uint) error {
	for _, item := range items {
		config := &model.SysConfig{
			BaseModel: common.BaseModel{
				UpdateBy: operatorID,
			},
			Key:   prefix + item.Key,
			Value: item.Value,
			Type:  1,
		}
		if err := s.configRepo.UpsertByKey(config); err != nil {
			return err
		}
	}
	return nil
}
