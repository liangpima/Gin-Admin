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
