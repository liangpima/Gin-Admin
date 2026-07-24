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
		return errors.New("用户名已存在")
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
		_ = s.userRepo.ReplaceRoles(user.ID, req.RoleIds)
	}
	if len(req.PostIds) > 0 {
		_ = s.userRepo.ReplacePosts(user.ID, req.PostIds)
	}

	return nil
}

func (s *userService) Update(tenantID uint, req *dto.UpdateUserRequest, operatorID uint) error {
	if s.userRepo.CountByUsername(tenantID, "", req.ID) > 0 {
		return errors.New("用户名已存在")
	}

	user, err := s.userRepo.FindByID(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
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
		_ = s.userRepo.ReplaceRoles(user.ID, req.RoleIds)
	}
	if req.PostIds != nil {
		_ = s.userRepo.ReplacePosts(user.ID, req.PostIds)
	}

	return nil
}

func (s *userService) Delete(tenantID, id uint) error {
	return s.userRepo.Delete(tenantID, id)
}

func (s *userService) FindByID(tenantID, id uint) (interface{}, error) {
	user, err := s.userRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	type userWithRoles struct {
		model.SysUser
		Roles []vo.RoleInfo `json:"roles"`
	}

	roleIDs, _ := s.userRepo.FindRoleIDsByUserID(user.ID)
	roleService := NewRoleService()
	roles, _ := roleService.FindByIDs(roleIDs)
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

	roleService := NewRoleService()
	result := make([]interface{}, len(users))
	for i, u := range users {
		roleIDs, _ := s.userRepo.FindRoleIDsByUserID(u.ID)
		roles, _ := roleService.FindByIDs(roleIDs)
		roleInfos := make([]vo.RoleInfo, 0, len(roles))
		for _, r := range roles {
			roleInfos = append(roleInfos, vo.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code})
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
		return errors.New("用户不存在")
	}
	return s.userRepo.ReplaceRoles(req.ID, req.RoleIds)
}

func (s *userService) UpdateDept(tenantID uint, req *dto.UpdateUserDeptRequest) error {
	user, err := s.userRepo.FindByID(tenantID, req.ID)
	if err != nil {
		return errors.New("用户不存在")
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
		return err
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return errors.New("旧密码错误")
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

// revokeUserTokens 吊销用户的所有 refresh token
func (s *userService) revokeUserTokens(userID uint) {
	ctx := context.Background()
	key := "refresh_token:user:" + fmt.Sprintf("%d", userID)
	_ = cache.Del(ctx, key)
	// 同时设置一个标记，使得该用户的所有旧 access token 失效
	_ = cache.Set(ctx, "user:token_revoked:"+fmt.Sprintf("%d", userID), "1",
		time.Duration(config.Cfg.JWT.AccessExpire)*time.Second)
}

// validatePasswordStrength 校验密码强度：至少包含大写字母、小写字母、数字中的两种
func validatePasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("密码长度不能少于6位")
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
		return errors.New("密码必须包含大写字母、小写字母、数字中的至少两种")
	}
	if strings.ContainsAny(password, " \t\n\r") {
		return errors.New("密码不能包含空格")
	}
	return nil
}
