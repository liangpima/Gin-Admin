package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type DeptService interface {
	Create(req *dto.CreateDeptRequest, operatorID uint) error
	Update(req *dto.UpdateDeptRequest, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (interface{}, error)
	FindTree() ([]model.SysDept, error)
}

type deptService struct {
	deptRepo repository.DeptRepository
}

func NewDeptService() DeptService {
	return &deptService{
		deptRepo: repository.NewDeptRepository(),
	}
}

func (s *deptService) Create(req *dto.CreateDeptRequest, operatorID uint) error {
	dept := &model.SysDept{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		ParentID: req.ParentID,
		Name:     req.Name,
		Sort:     req.Sort,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   req.Status,
	}

	return s.deptRepo.Create(dept)
}

func (s *deptService) Update(req *dto.UpdateDeptRequest, operatorID uint) error {
	dept, err := s.deptRepo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	dept.ParentID = req.ParentID
	dept.Name = req.Name
	dept.Sort = req.Sort
	dept.Leader = req.Leader
	dept.Phone = req.Phone
	dept.Email = req.Email
	dept.Status = req.Status
	dept.UpdateBy = operatorID

	return s.deptRepo.Update(dept)
}

func (s *deptService) Delete(id uint) error {
	return s.deptRepo.Delete(id)
}

func (s *deptService) FindByID(id uint) (interface{}, error) {
	return s.deptRepo.FindByID(id)
}

func (s *deptService) FindTree() ([]model.SysDept, error) {
	depts, err := s.deptRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return buildDeptTree(depts, 0), nil
}

func buildDeptTree(depts []model.SysDept, parentID uint) []model.SysDept {
	tree := make([]model.SysDept, 0)
	for _, dept := range depts {
		if dept.ParentID == parentID {
			children := buildDeptTree(depts, dept.ID)
			dept.Children = children
			tree = append(tree, dept)
		}
	}
	return tree
}
