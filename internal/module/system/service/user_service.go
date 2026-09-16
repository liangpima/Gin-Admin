package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"go-admin/config"
	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/logger"
	"go-admin/internal/middleware"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
	"go-admin/internal/module/system/vo"
	"go-admin/pkg/utils"

	"gorm.io/gorm"
)

type UserService interface {
	Create(tenantID uint, req *dto.CreateUserRequest, operatorID uint) error
	Update(tenantID uint, req *dto.UpdateUserRequest, operatorID uint) error
	Delete(tenantID, id uint) error
	FindByID(tenantID, id uint) (interface{}, error)
	FindList(tenantID uint, req *dto.UserListRequest) ([]interface{}, int64, error)
	UpdateStatus(tenantID uint, req *dto.StatusRequest) error
	UpdateRoles(tenantID uint, req *dto.UpdateUserRolesRequest) error
	UpdateDept(tenantID uint, req *dto.UpdateUserDeptRequest) error
	ResetPassword(tenantID uint, req *dto.ResetPasswordRequest) error
	ChangePassword(userID uint, req *dto.ChangePasswordRequest) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService() UserService {
	return &userService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *userService) Create(tenantID uint, req *dto.CreateUserRequest, operatorID uint) error {
	if s.userRepo.CountByUsername(tenantID, req.Username, 0) > 0 {
		return common.NewBizError("用户名已存在")
	}

	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.SysUser{
		TenantBaseModel: common.TenantBaseModel{
			BaseModel: common.BaseModel{
				CreateBy: operatorID,
				UpdateBy: operatorID,
			},
			TenantID: tenantID,
		},
		Username: req.Username,
		Password: hash,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   req.Status,
		DeptID:   req.DeptID,
	}
	user.Remark = req.Remark

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	if len(req.RoleIds) > 0 {
		if err := s.userRepo.ReplaceRoles(user.ID, req.RoleIds); err != nil {
			return err
		}
	}
	if len(req.PostIds) > 0 {
		if err := s.userRepo.ReplacePosts(user.ID, req.PostIds); err != nil {
			return err
		}
	}

	return nil
}

func (s *userService) Update(tenantID uint, req *dto.UpdateUserRequest, operatorID uint) error {
	// 更新不开放修改用户名（见 dto.UpdateUserRequest），因此无需重名校验。
	// 早前这里调用 CountByUsername(tenantID, "", req.ID) —— 传入空用户名恒为 0，
	// 「用户名已存在」是一条永远不会触发的死分支。存在性由下面的 FindByID 保证。
	user, err := s.userRepo.FindByID(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewNotFoundError("用户不存在")
		}
		return err
	}

	user.Nickname = req.Nickname
	user.Email = req.Email
	user.Phone = req.Phone
	user.Status = req.Status
	user.DeptID = req.DeptID
	user.Remark = req.Remark
	user.UpdateBy = operatorID

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if req.RoleIds != nil {
		if err := s.userRepo.ReplaceRoles(user.ID, req.RoleIds); err != nil {
			return err
		}
	}
	if req.PostIds != nil {
		if err := s.userRepo.ReplacePosts(user.ID, req.PostIds); err != nil {
			return err
		}
	}

	return nil
}

func (s *userService) Delete(tenantID, id uint) error {
	return s.userRepo.Delete(tenantID, id)
}

func (s *userService) FindByID(tenantID, id uint) (interface{}, error) {
	user, err := s.userRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "用户不存在")
	}

	type userWithRoles struct {
		model.SysUser
		Roles []vo.RoleInfo `json:"roles"`
	}

	roleIDs, err := s.userRepo.FindRoleIDsByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	roleService := NewRoleService()
	roles, err := roleService.FindByIDs(tenantID, roleIDs)
	if err != nil {
		return nil, err
	}
	roleInfos := make([]vo.RoleInfo, 0, len(roles))
	for _, r := range roles {
		roleInfos = append(roleInfos, vo.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code})
	}

	return userWithRoles{SysUser: *user, Roles: roleInfos}, nil
}

func (s *userService) FindList(tenantID uint, req *dto.UserListRequest) ([]interface{}, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	users, total, err := s.userRepo.FindList(tenantID, req.Username, req.Phone, req.Status, req.DeptID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	type userWithRoles struct {
		model.SysUser
		Roles []vo.RoleInfo `json:"roles"`
	}

	result := make([]interface{}, len(users))
	if len(users) == 0 {
		return result, total, nil
	}

	// 批量加载角色关系与角色详情，把 2N+1 次查询压缩为固定 3 次
	userIDs := make([]uint, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	roleIDsByUser, err := s.userRepo.FindRoleIDsByUserIDs(userIDs)
	if err != nil {
		return nil, 0, err
	}

	roleIDSet := make(map[uint]struct{})
	for _, ids := range roleIDsByUser {
		for _, id := range ids {
			roleIDSet[id] = struct{}{}
		}
	}
	distinctRoleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		distinctRoleIDs = append(distinctRoleIDs, id)
	}

	roleByID := make(map[uint]model.SysRole, len(distinctRoleIDs))
	if len(distinctRoleIDs) > 0 {
		roles, err := NewRoleService().FindByIDs(tenantID, distinctRoleIDs)
		if err != nil {
			return nil, 0, err
		}
		for _, r := range roles {
			roleByID[r.ID] = r
		}
	}

	for i, u := range users {
		ids := roleIDsByUser[u.ID]
		roleInfos := make([]vo.RoleInfo, 0, len(ids))
		for _, rid := range ids {
			if r, ok := roleByID[rid]; ok {
				roleInfos = append(roleInfos, vo.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code})
			}
		}
		result[i] = userWithRoles{SysUser: u, Roles: roleInfos}
	}
	return result, total, nil
}

func (s *userService) UpdateStatus(tenantID uint, req *dto.StatusRequest) error {
	if err := s.userRepo.UpdateStatus(tenantID, req.ID, req.Status); err != nil {
		return err
	}
	// 禁用用户时吊销其 Token
	if req.Status == common.StatusDisabled {
		s.revokeUserTokens(req.ID)
	}
	return nil
}

func (s *userService) UpdateRoles(tenantID uint, req *dto.UpdateUserRolesRequest) error {
	_, err := s.userRepo.FindByID(tenantID, req.ID)
	if err != nil {
		return common.NewNotFoundError("用户不存在")
	}
	if err := s.userRepo.ReplaceRoles(req.ID, req.RoleIds); err != nil {
		return err
	}
	// 角色变更后立即失效该用户的角色缓存，否则最长 60s 内仍按旧角色鉴权
	middleware.ClearRoleCache(tenantID, req.ID)
	return nil
}

func (s *userService) UpdateDept(tenantID uint, req *dto.UpdateUserDeptRequest) error {
	user, err := s.userRepo.FindByID(tenantID, req.ID)
	if err != nil {
		return common.NewNotFoundError("用户不存在")
	}
	user.DeptID = req.DeptID
	return s.userRepo.Update(user)
}

func (s *userService) ResetPassword(tenantID uint, req *dto.ResetPasswordRequest) error {
	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	return s.userRepo.ResetPassword(tenantID, req.ID, hash)
}

func (s *userService) ChangePassword(userID uint, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(0, userID)
	if err != nil {
		return common.NotFoundOrErr(err, "用户不存在")
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return common.NewBizError("旧密码错误")
	}

	if err := validatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	if err := s.userRepo.ResetPassword(0, userID, hash); err != nil {
		return err
	}

	// 密码修改后吊销所有 Token
	s.revokeUserTokens(userID)
	return nil
}

// revokeUserTokens 吊销用户的所有 refresh token，并使其旧 access token 失效。
//
// refresh token 本身是随机串，必须依靠登录时登记的用户维度集合才能枚举出来；
// 早前直接删 refresh_token:user:<id> 是删了一个从未写入的键，等于没吊销。
func (s *userService) revokeUserTokens(userID uint) {
	ctx := context.Background()

	tokens, err := cache.SMembers(ctx, cache.RefreshTokenSetKey(userID))
	if err != nil {
		logger.Log.Warnf("读取refresh token列表失败: %v", err)
	}

	keys := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		keys = append(keys, cache.RefreshTokenKey(t))
	}
	keys = append(keys, cache.RefreshTokenSetKey(userID))

	if err := cache.Del(ctx, keys...); err != nil {
		logger.Log.Warnf("吊销refresh token失败: %v", err)
	}

	// 同时设置一个标记，使得该用户的所有旧 access token 失效
	if err := cache.Set(ctx, fmt.Sprintf("user:token_revoked:%d", userID), "1",
		time.Duration(config.Cfg.JWT.AccessExpire)*time.Second); err != nil {
		logger.Log.Warnf("设置token吊销标记失败: %v", err)
	}
}

// validatePasswordStrength 校验密码强度：至少包含大写字母、小写字母、数字中的两种
func validatePasswordStrength(password string) error {
	if len(password) < 6 {
		return common.NewBizError("密码长度不能少于6位")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	types := 0
	if hasUpper {
		types++
	}
	if hasLower {
		types++
	}
	if hasDigit {
		types++
	}
	if types < 2 {
		return common.NewBizError("密码必须包含大写字母、小写字母、数字中的至少两种")
	}
	if strings.ContainsAny(password, " \t\n\r") {
		return common.NewBizError("密码不能包含空格")
	}
	return nil
}
