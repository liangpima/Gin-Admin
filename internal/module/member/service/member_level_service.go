package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/member/dto"
	"go-admin/internal/module/member/model"
	"go-admin/internal/module/member/repository"
)

type MemberLevelService interface {
	Create(req *dto.CreateMemberLevelRequest, operatorID uint) error
	Update(req *dto.UpdateMemberLevelRequest, operatorID uint) error
	Delete(id uint) error
	FindList(req *dto.MemberLevelListRequest) ([]model.MemberLevel, int64, error)
	FindAll() ([]model.MemberLevel, error)
}

type memberLevelService struct {
	levelRepo repository.MemberLevelRepository
}

func NewMemberLevelService() MemberLevelService {
	return &memberLevelService{
		levelRepo: repository.NewMemberLevelRepository(),
	}
}

func (s *memberLevelService) Create(req *dto.CreateMemberLevelRequest, operatorID uint) error {
	level := &model.MemberLevel{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
		},
		Name:      req.Name,
		MinPoints: req.MinPoints,
		Discount:  req.Discount,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Status:    req.Status,
	}
	return s.levelRepo.Create(level)
}

func (s *memberLevelService) Update(req *dto.UpdateMemberLevelRequest, operatorID uint) error {
	level, err := s.levelRepo.FindByID(req.ID)
	if err != nil {
		return errors.New("等级不存在")
	}
	level.Name = req.Name
	level.MinPoints = req.MinPoints
	level.Discount = req.Discount
	level.Icon = req.Icon
	level.Sort = req.Sort
	level.Status = req.Status
	level.UpdateBy = operatorID
	return s.levelRepo.Update(level)
}

func (s *memberLevelService) Delete(id uint) error {
	return s.levelRepo.Delete(id)
}

func (s *memberLevelService) FindList(req *dto.MemberLevelListRequest) ([]model.MemberLevel, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return s.levelRepo.FindList(req.Name, req.Page, req.PageSize)
}

func (s *memberLevelService) FindAll() ([]model.MemberLevel, error) {
	return s.levelRepo.FindAll()
}
