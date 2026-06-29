package service

import (
	"errors"

	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type AgreementService interface {
	Create(title, content, typ string, sort int, status int8, operatorID uint) error
	Update(id uint, title, content, typ string, sort int, status int8, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (*model.SysAgreement, error)
	FindByType(typ string) (*model.SysAgreement, error)
	FindList(name, typ string, status *int8, page, pageSize int) ([]model.SysAgreement, int64, error)
}

type agreementService struct {
	agreementRepo repository.AgreementRepository
}

func NewAgreementService() AgreementService {
	return &agreementService{
		agreementRepo: repository.NewAgreementRepository(),
	}
}

func (s *agreementService) Create(title, content, typ string, sort int, status int8, operatorID uint) error {
	agreement := &model.SysAgreement{
		Title:   title,
		Content: content,
		Type:    typ,
		Sort:    sort,
		Status:  status,
	}
	agreement.CreateBy = operatorID
	agreement.UpdateBy = operatorID
	return s.agreementRepo.Create(agreement)
}

func (s *agreementService) Update(id uint, title, content, typ string, sort int, status int8, operatorID uint) error {
	agreement, err := s.agreementRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在")
		}
		return err
	}

	agreement.Title = title
	agreement.Content = content
	agreement.Type = typ
	agreement.Sort = sort
	agreement.Status = status
	agreement.UpdateBy = operatorID

	return s.agreementRepo.Update(agreement)
}

func (s *agreementService) Delete(id uint) error {
	return s.agreementRepo.Delete(id)
}

func (s *agreementService) FindByID(id uint) (*model.SysAgreement, error) {
	return s.agreementRepo.FindByID(id)
}

func (s *agreementService) FindByType(typ string) (*model.SysAgreement, error) {
	return s.agreementRepo.FindByType(typ)
}

func (s *agreementService) FindList(name, typ string, status *int8, page, pageSize int) ([]model.SysAgreement, int64, error) {
	return s.agreementRepo.FindList(name, typ, status, page, pageSize)
}
