package service

import (
	"testing"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

// TestDictServiceCreateTypeValidatesCodeFormat 字典类型编码是代码里的字面量
// （前端 useDict('sys_user_status')），格式不对等于建完就废，必须挡住。
func TestDictServiceCreateTypeValidatesCodeFormat(t *testing.T) {
	cases := []struct {
		name string
		typ  string
	}{
		{"含中文", "用户状态"},
		{"含大写字母", "SysUserStatus"},
		{"以数字开头", "1status"},
		{"以大写开头", "Status"},
		{"含空格", "sys user status"},
		{"含短横线", "sys-user-status"},
		{"单字符", "a"},
		{"空字符串", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &mockDictRepo{}
			svc := newTestDictService(repo)

			err := svc.CreateType(&dto.CreateDictTypeRequest{Name: "测试", Type: c.typ}, 1)
			if err == nil {
				t.Fatalf("编码 %q 应被拒绝，实际通过", c.typ)
			}
			assertBizError(t, err, common.CodeBadRequest)
			if repo.createdType != nil {
				t.Errorf("校验失败不应落库，实际写入了: %+v", repo.createdType)
			}
		})
	}
}

func TestDictServiceCreateTypeRejectsDuplicate(t *testing.T) {
	repo := &mockDictRepo{
		findTypeByTypeFn: func(string) (*model.SysDictType, error) {
			// 复现真实实现：查到记录时也返回非 nil 指针
			return &model.SysDictType{BaseModel: common.BaseModel{ID: 9}, Type: "sys_user_status"}, nil
		},
	}
	svc := newTestDictService(repo)

	err := svc.CreateType(&dto.CreateDictTypeRequest{Name: "用户状态", Type: "sys_user_status"}, 1)
	assertBizError(t, err, common.CodeBadRequest)
	if repo.createdType != nil {
		t.Errorf("重名时不应落库")
	}
}

// TestDictServiceCreateTypeNotFoundMeansCreatable 钉住那个曾经踩过的坑：
// FindTypeByType 返回的是 (非nil指针, ErrRecordNotFound)，
// 若用 `existing != nil` 判断重名，新建字典类型会永远报「已存在」。
func TestDictServiceCreateTypeNotFoundMeansCreatable(t *testing.T) {
	repo := &mockDictRepo{} // 默认即返回 (&SysDictType{}, ErrRecordNotFound)
	svc := newTestDictService(repo)

	err := svc.CreateType(&dto.CreateDictTypeRequest{Name: "用户状态", Type: "sys_user_status"}, 7)
	if err != nil {
		t.Fatalf("查不到记录时应允许创建，实际报错: %v", err)
	}
	if repo.createdType == nil {
		t.Fatal("应写入字典类型")
	}
	if repo.createdType.CreateBy != 7 || repo.createdType.UpdateBy != 7 {
		t.Errorf("操作人应写入 CreateBy/UpdateBy，实际 %d/%d",
			repo.createdType.CreateBy, repo.createdType.UpdateBy)
	}
	if repo.createdType.Status != 1 {
		t.Errorf("新建字典类型应默认启用，实际 status=%d", repo.createdType.Status)
	}
}

// TestDictServiceCreateTypeMapsDuplicateKeyToBizError 并发下靠唯一索引兜底，
// 撞索引必须转成「已存在」的 400，而不是把 DB 错误暴露成 500。
func TestDictServiceCreateTypeMapsDuplicateKeyToBizError(t *testing.T) {
	repo := &mockDictRepo{
		createTypeFn: func(*model.SysDictType) error {
			return common.ErrDuplicateKey
		},
	}
	svc := newTestDictService(repo)

	assertBizError(t, svc.CreateType(&dto.CreateDictTypeRequest{Name: "x", Type: "sys_x"}, 1),
		common.CodeBadRequest)
}

func TestDictServiceCreateDataRequiresExistingType(t *testing.T) {
	repo := &mockDictRepo{} // FindTypeByType → ErrRecordNotFound
	svc := newTestDictService(repo)

	err := svc.CreateData(&dto.CreateDictDataRequest{
		DictType: "sys_not_exist", Label: "正常", Value: "1",
	}, 1)

	// 归属类型不存在属于「目标不存在」，必须是 404 而不是 400
	assertBizError(t, err, common.CodeNotFound)
	if repo.createdData != nil {
		t.Errorf("类型不存在时不应写入字典数据")
	}
}

func TestDictServiceCreateDataRejectsDuplicateValue(t *testing.T) {
	repo := &mockDictRepo{
		findTypeByTypeFn: func(string) (*model.SysDictType, error) {
			return &model.SysDictType{BaseModel: common.BaseModel{ID: 1}, Type: "sys_user_status"}, nil
		},
		countDataByValueFn: func(dictType, value string, excludeID uint) (int64, error) {
			if excludeID != 0 {
				t.Errorf("创建场景不应传 excludeID，实际 %d", excludeID)
			}
			return 1, nil
		},
	}
	svc := newTestDictService(repo)

	err := svc.CreateData(&dto.CreateDictDataRequest{
		DictType: "sys_user_status", Label: "正常", Value: "1",
	}, 1)
	assertBizError(t, err, common.CodeBadRequest)
	if repo.createdData != nil {
		t.Errorf("键值重复时不应写入")
	}
}

func TestDictServiceCreateDataSuccess(t *testing.T) {
	repo := &mockDictRepo{
		findTypeByTypeFn: func(string) (*model.SysDictType, error) {
			return &model.SysDictType{BaseModel: common.BaseModel{ID: 1}, Type: "sys_user_status"}, nil
		},
	}
	svc := newTestDictService(repo)

	err := svc.CreateData(&dto.CreateDictDataRequest{
		DictType:  "sys_user_status",
		Label:     "正常",
		Value:     "1",
		Sort:      1,
		ListClass: "success",
		Remark:    "默认启用",
	}, 3)
	if err != nil {
		t.Fatalf("创建字典数据失败: %v", err)
	}
	if repo.createdData == nil {
		t.Fatal("应写入字典数据")
	}
	if repo.createdData.Status != 1 {
		t.Errorf("新建字典数据应默认启用，实际 %d", repo.createdData.Status)
	}
	if repo.createdData.CreateBy != 3 {
		t.Errorf("操作人应为 3，实际 %d", repo.createdData.CreateBy)
	}
	// Remark 是 BaseModel 的提升字段，容易被漏写（复合字面量里写不进去）
	if repo.createdData.Remark != "默认启用" {
		t.Errorf("备注应写入 BaseModel.Remark，实际 %q", repo.createdData.Remark)
	}
}

func TestDictServiceUpdateDataNotFound(t *testing.T) {
	repo := &mockDictRepo{} // FindDataByID → ErrRecordNotFound
	svc := newTestDictService(repo)

	assertBizError(t, svc.UpdateData(99, &dto.UpdateDictDataRequest{Label: "x", Value: "1"}, 1),
		common.CodeNotFound)
	if repo.updatedData != nil {
		t.Errorf("记录不存在时不应更新")
	}
}

// TestDictServiceUpdateDataAllowsKeepingOwnValue 更新时必须排除自身，
// 否则「只改标签、键值不动」会被误判成重复。
func TestDictServiceUpdateDataAllowsKeepingOwnValue(t *testing.T) {
	repo := &mockDictRepo{
		findDataByIDFn: func(id uint) (*model.SysDictData, error) {
			return &model.SysDictData{
				BaseModel: common.BaseModel{ID: id},
				DictType:  "sys_user_status",
				Label:     "正常",
				Value:     "1",
				Status:    1,
			}, nil
		},
		countDataByValueFn: func(dictType, value string, excludeID uint) (int64, error) {
			if excludeID != 5 {
				t.Errorf("更新场景必须排除自身，实际 excludeID=%d", excludeID)
			}
			return 0, nil
		},
	}
	svc := newTestDictService(repo)

	err := svc.UpdateData(5, &dto.UpdateDictDataRequest{Label: "启用", Value: "1"}, 2)
	if err != nil {
		t.Fatalf("保持自身键值应允许更新: %v", err)
	}
	if repo.updatedData == nil || repo.updatedData.Label != "启用" {
		t.Errorf("应更新标签，实际 %+v", repo.updatedData)
	}
	if repo.updatedData.UpdateBy != 2 {
		t.Errorf("应记录操作人，实际 %d", repo.updatedData.UpdateBy)
	}
}

// TestDictServiceUpdateDataStatusOptional status 用指针是为了区分
// 「没传」和「传了 0（停用）」，不传时不能把状态重置。
func TestDictServiceUpdateDataStatusOptional(t *testing.T) {
	repo := &mockDictRepo{
		findDataByIDFn: func(id uint) (*model.SysDictData, error) {
			return &model.SysDictData{BaseModel: common.BaseModel{ID: id}, DictType: "t", Value: "1", Status: 0}, nil
		},
	}
	svc := newTestDictService(repo)

	if err := svc.UpdateData(1, &dto.UpdateDictDataRequest{Label: "x", Value: "1"}, 1); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if repo.updatedData.Status != 0 {
		t.Errorf("未传 status 时应保持原值 0，实际 %d", repo.updatedData.Status)
	}

	// 显式传 1 才改
	enabled := int8(1)
	if err := svc.UpdateData(1, &dto.UpdateDictDataRequest{Label: "x", Value: "1", Status: &enabled}, 1); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if repo.updatedData.Status != 1 {
		t.Errorf("显式传 status=1 应生效，实际 %d", repo.updatedData.Status)
	}
}

func TestDictServiceUpdateDataRejectsDuplicateValue(t *testing.T) {
	repo := &mockDictRepo{
		findDataByIDFn: func(id uint) (*model.SysDictData, error) {
			return &model.SysDictData{BaseModel: common.BaseModel{ID: id}, DictType: "sys_user_status", Value: "1"}, nil
		},
		countDataByValueFn: func(string, string, uint) (int64, error) { return 1, nil },
	}
	svc := newTestDictService(repo)

	assertBizError(t, svc.UpdateData(1, &dto.UpdateDictDataRequest{Label: "x", Value: "2"}, 1),
		common.CodeBadRequest)
	if repo.updatedData != nil {
		t.Errorf("键值重复时不应更新")
	}
}

func TestDictServiceDeleteTypeNotFound(t *testing.T) {
	svc := newTestDictService(&mockDictRepo{})

	assertBizError(t, svc.DeleteType(88), common.CodeNotFound)
}

// TestDictServiceDeleteTypeRejectsWhenHasData 字典数据只靠 dict_type 字符串关联，
// 没有外键约束 —— 类型先删掉会留下孤儿数据，必须拦住。
func TestDictServiceDeleteTypeRejectsWhenHasData(t *testing.T) {
	repo := &mockDictRepo{
		findTypeByIDFn: func(id uint) (*model.SysDictType, error) {
			return &model.SysDictType{BaseModel: common.BaseModel{ID: id}, Type: "sys_user_status"}, nil
		},
		countDataByTypeFn: func(dictType string) (int64, error) {
			if dictType != "sys_user_status" {
				t.Errorf("应按类型编码统计，实际 %q", dictType)
			}
			return 2, nil
		},
	}
	svc := newTestDictService(repo)

	assertBizError(t, svc.DeleteType(1), common.CodeBadRequest)
	if repo.deleteTypeHits != 0 {
		t.Errorf("有子级数据时不应执行删除，实际调用了 %d 次", repo.deleteTypeHits)
	}
}

func TestDictServiceDeleteTypeSuccess(t *testing.T) {
	repo := &mockDictRepo{
		findTypeByIDFn: func(id uint) (*model.SysDictType, error) {
			return &model.SysDictType{BaseModel: common.BaseModel{ID: id}, Type: "sys_empty"}, nil
		},
	}
	svc := newTestDictService(repo)

	if err := svc.DeleteType(1); err != nil {
		t.Fatalf("删除空类型应成功: %v", err)
	}
	if repo.deleteTypeHits != 1 {
		t.Errorf("应调用一次删除，实际 %d", repo.deleteTypeHits)
	}
}

func TestDictServiceDeleteDataNotFound(t *testing.T) {
	repo := &mockDictRepo{
		deleteDataFn: func(uint) error { return gorm.ErrRecordNotFound },
	}
	svc := newTestDictService(repo)

	assertBizError(t, svc.DeleteData(66), common.CodeNotFound)
}

func TestDictServiceDeleteDataSuccess(t *testing.T) {
	repo := &mockDictRepo{}
	svc := newTestDictService(repo)

	if err := svc.DeleteData(1); err != nil {
		t.Fatalf("删除字典数据失败: %v", err)
	}
}
