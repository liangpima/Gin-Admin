package service

import (
	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"errors"
	"strings"

	"gorm.io/gorm"
)

// maskedValue 敏感配置项在接口返回时使用的占位符。
//
// 前端若原样回传该值，说明用户没有修改它，BatchSave 会跳过写入，
// 以免把真实密钥覆盖成占位符本身。
const maskedValue = "******"

// sensitiveKeyFragments 配置项 key 命中以下片段即视为敏感信息，对外返回时打码。
// 宁可多打码（管理员仍可覆写），也不要让密钥回传到浏览器。
var sensitiveKeyFragments = []string{"secret", "password", "passwd", "key", "pem", "private", "token"}

func isSensitiveConfigKey(key string) bool {
	lower := strings.ToLower(key)
	for _, frag := range sensitiveKeyFragments {
		if strings.Contains(lower, frag) {
			return true
		}
	}
	return false
}

// maskConfig 对敏感配置项的值打码
func maskConfig(c model.SysConfig) model.SysConfig {
	if c.Value != "" && isSensitiveConfigKey(c.ConfigKey) {
		c.Value = maskedValue
	}
	return c
}

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
	// FindByPrefix 供接口使用，敏感项的值会被打码
	FindByPrefix(prefix string) ([]interface{}, error)
	// FindByPrefixRaw 供内部读取配置使用，返回真实值
	FindByPrefixRaw(prefix string) ([]interface{}, error)
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
		Name:      name,
		ConfigKey: key,
		Value:     value,
		Type:      typ,
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
	config.ConfigKey = key
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

// FindList 配置分页列表；敏感项的值会打码后再返回
func (s *configService) FindList(name string, page, pageSize int) ([]interface{}, int64, error) {
	configs, total, err := s.configRepo.FindList(name, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(configs))
	for i, c := range configs {
		result[i] = maskConfig(c)
	}
	return result, total, nil
}

// FindByPrefix 按前缀查询配置（接口用），敏感项的值会打码
func (s *configService) FindByPrefix(prefix string) ([]interface{}, error) {
	configs, err := s.configRepo.FindByKeyPrefix(prefix)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(configs))
	for i, c := range configs {
		result[i] = maskConfig(c)
	}
	return result, nil
}

// FindByPrefixRaw 按前缀查询配置并返回真实值，仅供服务内部读取密钥等配置使用
func (s *configService) FindByPrefixRaw(prefix string) ([]interface{}, error) {
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
		// 前端原样回传打码占位符，说明该项未被修改，跳过以保留原值
		if item.Value == maskedValue {
			continue
		}

		config := &model.SysConfig{
			BaseModel: common.BaseModel{
				UpdateBy: operatorID,
			},
			ConfigKey: prefix + item.Key,
			Value:     item.Value,
			Type:      1,
		}
		if err := s.configRepo.UpsertByKey(config); err != nil {
			return err
		}
	}
	return nil
}

// LoadOSSConfig 从 sys_config 表读取 oss.* 配置，返回 key-value map。
// 需要真实密钥，因此使用 FindByPrefixRaw。
func LoadOSSConfig() map[string]string {
	svc := NewConfigService()
	results, _ := svc.FindByPrefixRaw("oss.")

	cfgMap := make(map[string]string)
	for _, r := range results {
		if cfg, ok := r.(model.SysConfig); ok {
			key := cfg.ConfigKey
			if len(key) > 4 && key[:4] == "oss." {
				cfgMap[key[4:]] = cfg.Value
			}
		}
	}
	return cfgMap
}
