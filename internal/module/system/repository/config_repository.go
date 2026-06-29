package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type ConfigRepository interface {
	Create(config *model.SysConfig) error
	FindByID(id uint) (*model.SysConfig, error)
	FindByKey(key string) (*model.SysConfig, error)
	FindList(name string, page, pageSize int) ([]model.SysConfig, int64, error)
	Update(config *model.SysConfig) error
	Delete(id uint) error
	FindByKeyPrefix(prefix string) ([]model.SysConfig, error)
	UpsertByKey(config *model.SysConfig) error
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigRepository() ConfigRepository {
	return &configRepository{db: database.DB}
}

func (r *configRepository) Create(config *model.SysConfig) error {
	return r.db.Create(config).Error
}

func (r *configRepository) FindByID(id uint) (*model.SysConfig, error) {
	var config model.SysConfig
	err := r.db.First(&config, id).Error
	return &config, err
}

func (r *configRepository) FindByKey(key string) (*model.SysConfig, error) {
	var config model.SysConfig
	err := r.db.Where("`key` = ?", key).First(&config).Error
	return &config, err
}

func (r *configRepository) FindList(name string, page, pageSize int) ([]model.SysConfig, int64, error) {
	var configs []model.SysConfig
	var total int64

	query := r.db.Model(&model.SysConfig{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&configs).Error
	return configs, total, err
}

func (r *configRepository) Update(config *model.SysConfig) error {
	return r.db.Model(config).Select("Name", "Key", "Value", "Type", "Remark", "UpdateBy").Updates(config).Error
}

func (r *configRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysConfig{}, id).Error
}

func (r *configRepository) FindByKeyPrefix(prefix string) ([]model.SysConfig, error) {
	var configs []model.SysConfig
	err := r.db.Where("`key` LIKE ?", prefix+"%").Order("id ASC").Find(&configs).Error
	return configs, err
}

func (r *configRepository) UpsertByKey(config *model.SysConfig) error {
	var existing model.SysConfig
	err := r.db.Where("`key` = ?", config.Key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(config).Error
	}
	if err != nil {
		return err
	}
	existing.Value = config.Value
	existing.UpdateBy = config.UpdateBy
	return r.db.Model(&existing).Select("Value", "UpdateBy").Updates(&existing).Error
}
