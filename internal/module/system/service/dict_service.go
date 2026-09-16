package service

import (
	"errors"
	"regexp"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

// dictTypePattern 字典类型编码的合法格式：小写字母开头，仅含小写字母、数字、下划线。
//
// 为什么要卡格式：这个编码是**代码里的字面量**（前端 useDict('sys_user_status')、
// 后端按类型取值），填中文或带空格虽然能存进库，但代码里根本没法引用，
// 属于建完就废的数据。长度上限 64 远小于列宽 128，留出余量。
var dictTypePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

type DictService interface {
	CreateType(req *dto.CreateDictTypeRequest, operatorID uint) error
	FindTypeList(name string, page, pageSize int) ([]interface{}, int64, error)
	FindTypeByID(id uint) (interface{}, error)
	UpdateType(id uint, req *dto.UpdateDictTypeRequest, operatorID uint) error
	DeleteType(id uint) error

	CreateData(req *dto.CreateDictDataRequest, operatorID uint) error
	FindDataByType(typ string) ([]model.SysDictData, error)
	FindDataList(typ string, page, pageSize int) ([]interface{}, int64, error)
	UpdateData(id uint, req *dto.UpdateDictDataRequest, operatorID uint) error
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

func (s *dictService) CreateType(req *dto.CreateDictTypeRequest, operatorID uint) error {
	if !dictTypePattern.MatchString(req.Type) {
		return common.NewBizError("字典类型编码只能由小写字母、数字、下划线组成，且以字母开头")
	}

	// 注意：不能用 `existing, _ := ...; if existing != nil` 判断重名 ——
	// FindTypeByType 无论查没查到都返回非 nil 指针（&dictType, err），
	// 该条件恒为真，会让新建字典类型永远报「已存在」。
	// 判定依据只能是 err：nil 表示查到，ErrRecordNotFound 表示可以创建。
	existing, err := s.dictRepo.FindTypeByType(req.Type)
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
		Name:   req.Name,
		Type:   req.Type,
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
	dictType, err := s.dictRepo.FindTypeByID(id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "字典类型不存在")
	}
	return dictType, nil
}

func (s *dictService) UpdateType(id uint, req *dto.UpdateDictTypeRequest, operatorID uint) error {
	dictType, err := s.dictRepo.FindTypeByID(id)
	if err != nil {
		return common.NotFoundOrErr(err, "字典类型不存在")
	}

	dictType.Name = req.Name
	dictType.Remark = req.Remark
	if req.Status != nil {
		dictType.Status = *req.Status
	}
	dictType.UpdateBy = operatorID

	return s.dictRepo.UpdateType(dictType)
}

// DeleteType 删除字典类型。
//
// 有子级字典数据时拒绝删除：数据表只按 dict_type 字符串关联，没有外键约束，
// 类型删掉后这些数据会变成孤儿 —— 列表里查不到、按类型也取不到，
// 却仍占着 uk_dict_type_value，将来重建同名类型时会莫名报「键值重复」。
func (s *dictService) DeleteType(id uint) error {
	dictType, err := s.dictRepo.FindTypeByID(id)
	if err != nil {
		return common.NotFoundOrErr(err, "字典类型不存在")
	}

	count, err := s.dictRepo.CountDataByType(dictType.Type)
	if err != nil {
		return err
	}
	if count > 0 {
		return common.NewBizError("该字典类型下还有字典数据，请先删除")
	}

	return s.dictRepo.DeleteType(id)
}

func (s *dictService) CreateData(req *dto.CreateDictDataRequest, operatorID uint) error {
	// 先确认归属类型存在：字典数据靠 dict_type 字符串关联，
	// 类型不存在时插入的数据永远不会被任何页面读到（列表要按类型查）。
	if _, err := s.dictRepo.FindTypeByType(req.DictType); err != nil {
		return common.NotFoundOrErr(err, "字典类型不存在")
	}

	if err := s.ensureValueUnique(req.DictType, req.Value, 0); err != nil {
		return err
	}

	dictData := &model.SysDictData{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
			// Remark 是 BaseModel 的提升字段，复合字面量里不能直接写，必须从内层给
			Remark: req.Remark,
		},
		DictType:  req.DictType,
		Label:     req.Label,
		Value:     req.Value,
		Sort:      req.Sort,
		CssClass:  req.CssClass,
		ListClass: req.ListClass,
		Status:    1,
	}

	if err := s.dictRepo.CreateData(dictData); err != nil {
		if errors.Is(err, common.ErrDuplicateKey) {
			return common.NewBizError("该字典类型下已存在相同的键值")
		}
		return err
	}
	return nil
}

// ensureValueUnique 校验同一类型下键值不重复，excludeID > 0 时排除自身（更新场景）
func (s *dictService) ensureValueUnique(dictType, value string, excludeID uint) error {
	count, err := s.dictRepo.CountDataByValue(dictType, value, excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return common.NewBizError("该字典类型下已存在相同的键值")
	}
	return nil
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

func (s *dictService) UpdateData(id uint, req *dto.UpdateDictDataRequest, operatorID uint) error {
	dictData, err := s.dictRepo.FindDataByID(id)
	if err != nil {
		return common.NotFoundOrErr(err, "字典数据不存在")
	}

	if err := s.ensureValueUnique(dictData.DictType, req.Value, dictData.ID); err != nil {
		return err
	}

	dictData.Label = req.Label
	dictData.Value = req.Value
	dictData.Sort = req.Sort
	dictData.CssClass = req.CssClass
	dictData.ListClass = req.ListClass
	dictData.Remark = req.Remark
	if req.Status != nil {
		dictData.Status = *req.Status
	}
	dictData.UpdateBy = operatorID

	if err := s.dictRepo.UpdateData(dictData); err != nil {
		if errors.Is(err, common.ErrDuplicateKey) {
			return common.NewBizError("该字典类型下已存在相同的键值")
		}
		return err
	}
	return nil
}

func (s *dictService) DeleteData(id uint) error {
	if err := s.dictRepo.DeleteData(id); err != nil {
		return common.NotFoundOrErr(err, "字典数据不存在")
	}
	return nil
}
