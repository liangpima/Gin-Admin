package service

import (
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

// mockDictRepo 是 DictRepository 的可控替身（函数字段模式，见 mock_repository_test.go）。
//
// 默认行为刻意模仿真实实现的两处特征：
//   - FindTypeByType 查不到时返回**非 nil 指针 + ErrRecordNotFound**
//     （真实实现是 `return &dictType, err`），这正是 Service 里必须用 err
//     而不是 `existing != nil` 判断重名的原因，测试要能复现这个坑
//   - 其余单条查询查不到时返回 ErrRecordNotFound
type mockDictRepo struct {
	createTypeFn      func(*model.SysDictType) error
	findTypeByIDFn    func(id uint) (*model.SysDictType, error)
	findTypeByTypeFn  func(typ string) (*model.SysDictType, error)
	updateTypeFn      func(*model.SysDictType) error
	deleteTypeFn      func(id uint) error
	createDataFn      func(*model.SysDictData) error
	findDataByIDFn    func(id uint) (*model.SysDictData, error)
	countDataByValueFn func(dictType, value string, excludeID uint) (int64, error)
	countDataByTypeFn func(dictType string) (int64, error)
	updateDataFn      func(*model.SysDictData) error
	deleteDataFn      func(id uint) error

	// 调用记录
	createdType    *model.SysDictType
	createdData    *model.SysDictData
	updatedType    *model.SysDictType
	updatedData    *model.SysDictData
	deleteTypeHits int
}

func (m *mockDictRepo) CreateType(dictType *model.SysDictType) error {
	m.createdType = dictType
	if m.createTypeFn != nil {
		return m.createTypeFn(dictType)
	}
	return nil
}

func (m *mockDictRepo) FindTypeByID(id uint) (*model.SysDictType, error) {
	if m.findTypeByIDFn != nil {
		return m.findTypeByIDFn(id)
	}
	return &model.SysDictType{}, gorm.ErrRecordNotFound
}

func (m *mockDictRepo) FindTypeByType(typ string) (*model.SysDictType, error) {
	if m.findTypeByTypeFn != nil {
		return m.findTypeByTypeFn(typ)
	}
	return &model.SysDictType{}, gorm.ErrRecordNotFound
}

func (m *mockDictRepo) FindTypeList(name string, page, pageSize int) ([]model.SysDictType, int64, error) {
	return nil, 0, nil
}

func (m *mockDictRepo) UpdateType(dictType *model.SysDictType) error {
	m.updatedType = dictType
	if m.updateTypeFn != nil {
		return m.updateTypeFn(dictType)
	}
	return nil
}

func (m *mockDictRepo) DeleteType(id uint) error {
	m.deleteTypeHits++
	if m.deleteTypeFn != nil {
		return m.deleteTypeFn(id)
	}
	return nil
}

func (m *mockDictRepo) CreateData(dictData *model.SysDictData) error {
	m.createdData = dictData
	if m.createDataFn != nil {
		return m.createDataFn(dictData)
	}
	return nil
}

func (m *mockDictRepo) FindDataByID(id uint) (*model.SysDictData, error) {
	if m.findDataByIDFn != nil {
		return m.findDataByIDFn(id)
	}
	return &model.SysDictData{}, gorm.ErrRecordNotFound
}

func (m *mockDictRepo) FindDataByType(typ string) ([]model.SysDictData, error) {
	return nil, nil
}

func (m *mockDictRepo) FindDataList(typ string, page, pageSize int) ([]model.SysDictData, int64, error) {
	return nil, 0, nil
}

func (m *mockDictRepo) CountDataByValue(dictType, value string, excludeID uint) (int64, error) {
	if m.countDataByValueFn != nil {
		return m.countDataByValueFn(dictType, value, excludeID)
	}
	return 0, nil
}

func (m *mockDictRepo) CountDataByType(dictType string) (int64, error) {
	if m.countDataByTypeFn != nil {
		return m.countDataByTypeFn(dictType)
	}
	return 0, nil
}

func (m *mockDictRepo) UpdateData(dictData *model.SysDictData) error {
	m.updatedData = dictData
	if m.updateDataFn != nil {
		return m.updateDataFn(dictData)
	}
	return nil
}

func (m *mockDictRepo) DeleteData(id uint) error {
	if m.deleteDataFn != nil {
		return m.deleteDataFn(id)
	}
	return nil
}

func newTestDictService(repo *mockDictRepo) *dictService {
	return &dictService{dictRepo: repo}
}
