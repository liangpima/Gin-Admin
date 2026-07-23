package repository

import (
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/module/member/model"

	"gorm.io/gorm"
)

type MemberRepository interface {
	Create(member *model.Member) error
	Update(member *model.Member) error
	Delete(tenantID, id uint) error
	FindByID(tenantID, id uint) (*model.Member, error)
	FindByPhone(tenantID uint, phone string) (*model.Member, error)
	FindByWechatOpenid(tenantID uint, openid string) (*model.Member, error)
	FindList(tenantID uint, phone, nickname string, levelID uint, status int8, page, pageSize int) ([]model.Member, int64, error)
	UpdateStatus(tenantID, id uint, status int8) error
	ReplaceTags(memberID uint, tagIDs []uint) error
	FindTagIDsByMemberID(memberID uint) ([]uint, error)
	UpdatePoints(memberID uint, points int64) error
	FindMaxMemberNo(tenantID uint) (string, error)
}

type memberRepository struct{}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}

func (r *memberRepository) Create(member *model.Member) error {
	return database.DB.Create(member).Error
}

func (r *memberRepository) Update(member *model.Member) error {
	return database.DB.Save(member).Error
}

func (r *memberRepository) Delete(tenantID, id uint) error {
	return common.TenantScope(database.DB, tenantID).Delete(&model.Member{}, id).Error
}

func (r *memberRepository) FindByID(tenantID, id uint) (*model.Member, error) {
	var member model.Member
	err := common.TenantScope(database.DB, tenantID).First(&member, id).Error
	return &member, err
}

func (r *memberRepository) FindByPhone(tenantID uint, phone string) (*model.Member, error) {
	var member model.Member
	err := common.TenantScope(database.DB, tenantID).Where("phone = ?", phone).First(&member).Error
	return &member, err
}

func (r *memberRepository) FindByWechatOpenid(tenantID uint, openid string) (*model.Member, error) {
	var member model.Member
	err := common.TenantScope(database.DB, tenantID).Where("wechat_openid = ?", openid).First(&member).Error
	return &member, err
}

func (r *memberRepository) FindList(tenantID uint, phone, nickname string, levelID uint, status int8, page, pageSize int) ([]model.Member, int64, error) {
	var members []model.Member
	var total int64

	query := common.TenantScope(database.DB.Model(&model.Member{}), tenantID)

	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if levelID > 0 {
		query = query.Where("level_id = ?", levelID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&members).Error
	return members, total, err
}

func (r *memberRepository) UpdateStatus(tenantID, id uint, status int8) error {
	return common.TenantScope(database.DB, tenantID).Model(&model.Member{}).Where("id = ?", id).Update("status", status).Error
}

func (r *memberRepository) ReplaceTags(memberID uint, tagIDs []uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("member_id = ?", memberID).Delete(&model.MemberTagRel{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}
		rels := make([]model.MemberTagRel, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			rels = append(rels, model.MemberTagRel{MemberID: memberID, TagID: tagID})
		}
		return tx.Create(&rels).Error
	})
}

func (r *memberRepository) FindTagIDsByMemberID(memberID uint) ([]uint, error) {
	var tagIDs []uint
	err := database.DB.Model(&model.MemberTagRel{}).
		Where("member_id = ?", memberID).
		Pluck("tag_id", &tagIDs).Error
	return tagIDs, err
}

func (r *memberRepository) UpdatePoints(memberID uint, points int64) error {
	return database.DB.Model(&model.Member{}).Where("id = ?", memberID).UpdateColumn("points", points).Error
}

func (r *memberRepository) FindMaxMemberNo(tenantID uint) (string, error) {
	var memberNo string
	err := common.TenantScope(database.DB.Model(&model.Member{}), tenantID).
		Select("member_no").
		Where("member_no != ''").
		Order("member_no DESC").
		Limit(1).
		Pluck("member_no", &memberNo).Error
	return memberNo, err
}
