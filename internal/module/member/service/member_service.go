package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/module/member/dto"
	"go-admin/internal/module/member/model"
	"go-admin/internal/module/member/repository"
)

type MemberService interface {
	Create(req *dto.CreateMemberRequest, operatorID, tenantID uint, digits int) error
	Update(req *dto.UpdateMemberRequest, operatorID, tenantID uint) error
	Delete(tenantID, id uint) error
	FindByID(tenantID, id uint) (*model.Member, error)
	FindList(tenantID uint, req *dto.MemberListRequest) ([]interface{}, int64, error)
	UpdateStatus(tenantID uint, req *dto.UpdateMemberStatusRequest) error
	UpdateTags(tenantID uint, req *dto.UpdateMemberTagsRequest) error
}

type memberService struct {
	memberRepo repository.MemberRepository
	tagRepo    repository.MemberTagRepository
	levelRepo  repository.MemberLevelRepository
}

func NewMemberService() MemberService {
	return &memberService{
		memberRepo: repository.NewMemberRepository(),
		tagRepo:    repository.NewMemberTagRepository(),
		levelRepo:  repository.NewMemberLevelRepository(),
	}
}

func (s *memberService) Create(req *dto.CreateMemberRequest, operatorID, tenantID uint, digits int) error {
	if req.Phone != "" {
		existing, _ := s.memberRepo.FindByPhone(tenantID, req.Phone)
		if existing != nil && existing.ID > 0 {
			return errors.New("手机号已注册")
		}
	}

	if digits < 4 {
		digits = 6
	}

	memberNo, err := s.generateMemberNo(tenantID, digits)
	if err != nil {
		return fmt.Errorf("生成会员编号失败: %v", err)
	}

	member := &model.Member{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
			TenantID: tenantID,
		},
		MemberNo:     memberNo,
		Username:     req.Username,
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Phone:        req.Phone,
		Gender:       req.Gender,
		LevelID:      req.LevelID,
		Status:       req.Status,
		Points:       0,
		RegisterTime: time.Now(),
	}
	member.Remark = req.Remark

	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			member.Birthday = &t
		}
	}

	if err := s.memberRepo.Create(member); err != nil {
		return err
	}

	if len(req.TagIds) > 0 {
		_ = s.memberRepo.ReplaceTags(member.ID, req.TagIds)
	}

	return nil
}

func (s *memberService) Update(req *dto.UpdateMemberRequest, operatorID, tenantID uint) error {
	member, err := s.memberRepo.FindByID(tenantID, req.ID)
	if err != nil {
		return errors.New("会员不存在")
	}

	member.Username = req.Username
	member.Nickname = req.Nickname
	member.Avatar = req.Avatar
	member.Gender = req.Gender
	member.LevelID = req.LevelID
	member.Status = req.Status
	member.UpdateBy = operatorID
	member.Remark = req.Remark

	if req.Phone != "" {
		member.Phone = req.Phone
	}

	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			member.Birthday = &t
		}
	} else {
		member.Birthday = nil
	}

	if err := s.memberRepo.Update(member); err != nil {
		return err
	}

	if req.TagIds != nil {
		_ = s.memberRepo.ReplaceTags(member.ID, req.TagIds)
	}

	return nil
}

func (s *memberService) Delete(tenantID, id uint) error {
	return s.memberRepo.Delete(tenantID, id)
}

func (s *memberService) FindByID(tenantID, id uint) (*model.Member, error) {
	return s.memberRepo.FindByID(tenantID, id)
}

func (s *memberService) FindList(tenantID uint, req *dto.MemberListRequest) ([]interface{}, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	status := int8(-1)
	if req.Status != nil {
		status = *req.Status
	}

	members, total, err := s.memberRepo.FindList(tenantID, req.Phone, req.Nickname, req.LevelID, status, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	type memberWithTag struct {
		model.Member
		Tags []model.MemberTag `json:"tags"`
	}

	result := make([]interface{}, len(members))
	for i, m := range members {
		tags, _ := s.memberRepo.FindTagsByMemberID(m.ID)
		result[i] = memberWithTag{Member: m, Tags: tags}
	}
	return result, total, nil
}

func (s *memberService) UpdateStatus(tenantID uint, req *dto.UpdateMemberStatusRequest) error {
	return s.memberRepo.UpdateStatus(tenantID, req.ID, req.Status)
}

func (s *memberService) generateMemberNo(tenantID uint, digits int) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("member:no:%d", tenantID)

	// 使用 Redis INCR 原子递增，避免并发重复
	seq, err := cache.Incr(ctx, key)
	if err != nil {
		// Redis 不可用时回退到数据库查询（有竞态风险，但可接受降级）
		maxNo, _ := s.memberRepo.FindMaxMemberNo(tenantID)
		startNum := int(math.Pow10(digits-1)) + 1
		nextNum := startNum
		if maxNo != "" {
			n, parseErr := strconv.Atoi(maxNo)
			if parseErr == nil && n >= startNum {
				nextNum = n + 1
			}
		}
		format := fmt.Sprintf("%%0%dd", digits)
		return fmt.Sprintf(format, nextNum), nil
	}

	// 首次初始化：如果序列为1，设置初始值为当前数据库最大值+1
	if seq == 1 {
		maxNo, _ := s.memberRepo.FindMaxMemberNo(tenantID)
		startNum := int(math.Pow10(digits-1)) + 1
		if maxNo != "" {
			if n, parseErr := strconv.Atoi(maxNo); parseErr == nil && n >= startNum {
				startNum = n + 1
			}
		}
		// 用 SETNX 设置初始值，失败则说明其他 goroutine 已设置
		_ = cache.Set(ctx, key, strconv.Itoa(startNum+1), 0)
		format := fmt.Sprintf("%%0%dd", digits)
		return fmt.Sprintf(format, startNum), nil
	}

	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, seq), nil
}

func (s *memberService) UpdateTags(tenantID uint, req *dto.UpdateMemberTagsRequest) error {
	_, err := s.memberRepo.FindByID(tenantID, req.ID)
	if err != nil {
		return errors.New("会员不存在")
	}
	return s.memberRepo.ReplaceTags(req.ID, req.TagIds)
}
