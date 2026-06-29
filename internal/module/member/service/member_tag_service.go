package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/member/dto"
	"go-admin/internal/module/member/model"
	"go-admin/internal/module/member/repository"
)

type MemberTagService interface {
	Create(req *dto.CreateMemberTagRequest, operatorID uint) error
	Update(req *dto.UpdateMemberTagRequest, operatorID uint) error
	Delete(id uint) error
	FindList(req *dto.MemberTagListRequest) ([]model.MemberTag, int64, error)
	FindAll() ([]model.MemberTag, error)
}

type memberTagService struct {
	tagRepo repository.MemberTagRepository
}

func NewMemberTagService() MemberTagService {
	return &memberTagService{
		tagRepo: repository.NewMemberTagRepository(),
	}
}

func (s *memberTagService) Create(req *dto.CreateMemberTagRequest, operatorID uint) error {
	tag := &model.MemberTag{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
		},
		Name:   req.Name,
		Color:  req.Color,
		Sort:   req.Sort,
		Status: req.Status,
	}
	return s.tagRepo.Create(tag)
}

func (s *memberTagService) Update(req *dto.UpdateMemberTagRequest, operatorID uint) error {
	tag, err := s.tagRepo.FindByID(req.ID)
	if err != nil {
		return errors.New("标签不存在")
	}
	tag.Name = req.Name
	tag.Color = req.Color
	tag.Sort = req.Sort
	tag.Status = req.Status
	tag.UpdateBy = operatorID
	return s.tagRepo.Update(tag)
}

func (s *memberTagService) Delete(id uint) error {
	return s.tagRepo.Delete(id)
}

func (s *memberTagService) FindList(req *dto.MemberTagListRequest) ([]model.MemberTag, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return s.tagRepo.FindList(req.Name, req.Page, req.PageSize)
}

func (s *memberTagService) FindAll() ([]model.MemberTag, error) {
	return s.tagRepo.FindAll()
}
