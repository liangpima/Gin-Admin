package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
	"go-admin/pkg/sanitize"

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
		Title: title,
		// 入库前净化：content 由富文本编辑器产出，是**原始 HTML**，
		// 必须在这里（写入口）过滤，而不是指望各渲染点自己处理。
		// 放大因素详见 pkg/sanitize 包注释：本表是全局表（无 tenant_id，
		// 所有租户共享），且 token 存在非 httpOnly Cookie 里，一次 XSS 即可接管账号。
		Content: sanitize.RichText(content),
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
			return common.NewNotFoundError("记录不存在")
		}
		return err
	}

	agreement.Title = title
	// 同 Create：更新路径也必须净化，否则可以先存干净内容再改成恶意内容绕过
	agreement.Content = sanitize.RichText(content)
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
	agreement, err := s.agreementRepo.FindByID(id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "记录不存在")
	}
	return agreement, nil
}

func (s *agreementService) FindByType(typ string) (*model.SysAgreement, error) {
	return s.agreementRepo.FindByType(typ)
}

func (s *agreementService) FindList(name, typ string, status *int8, page, pageSize int) ([]model.SysAgreement, int64, error) {
	return s.agreementRepo.FindList(name, typ, status, page, pageSize)
}
