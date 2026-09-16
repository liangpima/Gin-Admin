package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

// ErrDeptHasChildren 删除部门时存在下级部门。
//
// 用哨兵错误暴露出来，让 Controller 能把它归为 400（参数/前置条件不满足），
// 而不是笼统地返回 500 —— 这是调用方的用法问题，不是服务端故障。
var ErrDeptHasChildren = common.NewBizError("存在下级部门，请先删除下级部门")

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
			return common.NewNotFoundError("部门不存在")
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

// Delete 删除部门。
//
// 存在子部门时拒绝删除：直接删父节点会让子部门的 parent_id 悬空，
// 而 FindTree 从 parent_id=0 递归构建，这棵子树会「从界面上消失」，
// 数据却还在库里，既看不见也删不掉，成为孤儿数据。
func (s *deptService) Delete(id uint) error {
	children, err := s.deptRepo.CountByParentID(id)
	if err != nil {
		return err
	}
	if children > 0 {
		return ErrDeptHasChildren
	}
	return s.deptRepo.Delete(id)
}

func (s *deptService) FindByID(id uint) (interface{}, error) {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "部门不存在")
	}
	return dept, nil
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
