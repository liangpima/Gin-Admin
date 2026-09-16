package repository

import (
	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"gorm.io/gorm"
)

type DeptRepository interface {
	Create(dept *model.SysDept) error
	FindByID(id uint) (*model.SysDept, error)
	FindAll() ([]model.SysDept, error)
	Update(dept *model.SysDept) error
	Delete(id uint) error
	// CountByParentID 统计直属子部门数量，用于删除前校验
	CountByParentID(parentID uint) (int64, error)
	// FindParentID 返回部门的父节点 ID，ok=false 表示部门不存在（用于父级成环校验）
	FindParentID(id uint) (uint, bool, error)
}

type deptRepository struct {
	db *gorm.DB
}

func NewDeptRepository() DeptRepository {
	return &deptRepository{db: database.DB}
}

func (r *deptRepository) Create(dept *model.SysDept) error {
	return r.db.Create(dept).Error
}

func (r *deptRepository) FindByID(id uint) (*model.SysDept, error) {
	var dept model.SysDept
	err := r.db.First(&dept, id).Error
	return &dept, err
}

func (r *deptRepository) FindAll() ([]model.SysDept, error) {
	var depts []model.SysDept
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&depts).Error
	return depts, err
}

func (r *deptRepository) Update(dept *model.SysDept) error {
	return r.db.Model(dept).Select("ParentID", "Name", "Sort", "Leader", "Phone", "Email", "Status", "Remark", "UpdateBy").Updates(dept).Error
}

func (r *deptRepository) Delete(id uint) error {
	return r.db.Delete(&model.SysDept{}, id).Error
}

func (r *deptRepository) CountByParentID(parentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SysDept{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// FindParentID 查询部门的父节点 ID
func (r *deptRepository) FindParentID(id uint) (uint, bool, error) {
	var parents []uint
	if err := r.db.Model(&model.SysDept{}).Where("id = ?", id).Pluck("parent_id", &parents).Error; err != nil {
		return 0, false, err
	}
	if len(parents) == 0 {
		return 0, false, nil
	}
	return parents[0], true, nil
}
