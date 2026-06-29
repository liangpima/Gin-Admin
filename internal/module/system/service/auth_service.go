package service

import (
	"context"
	"errors"
	"time"

	"go-admin/config"
	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
	"go-admin/internal/module/system/vo"
	"go-admin/pkg/auth"
	"go-admin/pkg/utils"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(req *dto.LoginRequest) (*vo.LoginResponse, error)
	RefreshToken(req *dto.RefreshTokenRequest) (*vo.LoginResponse, error)
	Logout(token string) error
	GetUserInfo(userID uint) (*vo.UserInfoResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	menuRepo repository.MenuRepository
}

func NewAuthService() AuthService {
	return &authService{
		userRepo: repository.NewUserRepository(),
		roleRepo: repository.NewRoleRepository(),
		menuRepo: repository.NewMenuRepository(),
	}
}

func (s *authService) Login(req *dto.LoginRequest) (*vo.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	if user.Status == common.StatusDisabled {
		return nil, errors.New("用户已被禁用")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("密码错误")
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, user.TenantID, user.DeptID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, user.TenantID, user.DeptID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	_ = cache.Set(ctx, "refresh_token:"+refreshToken, user.ID, time.Duration(config.Cfg.JWT.RefreshExpire)*time.Second)

	now := time.Now()
	user.LoginTime = &now
	_ = s.userRepo.Update(user)

	return &vo.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.Cfg.JWT.AccessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) RefreshToken(req *dto.RefreshTokenRequest) (*vo.LoginResponse, error) {
	claims, err := auth.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token无效")
	}

	ctx := context.Background()
	exists, _ := cache.Exists(ctx, "refresh_token:"+req.RefreshToken)
	if !exists {
		return nil, errors.New("refresh token已过期")
	}

	_ = cache.Del(ctx, "refresh_token:"+req.RefreshToken)

	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.Username, claims.TenantID, claims.DeptID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(claims.UserID, claims.Username, claims.TenantID, claims.DeptID)
	if err != nil {
		return nil, err
	}

	_ = cache.Set(ctx, "refresh_token:"+refreshToken, claims.UserID, time.Duration(config.Cfg.JWT.RefreshExpire)*time.Second)

	return &vo.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.Cfg.JWT.AccessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) Logout(token string) error {
	ctx := context.Background()
	return cache.Del(ctx, "refresh_token:"+token)
}

func (s *authService) GetUserInfo(userID uint) (*vo.UserInfoResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	userRoles, err := s.userRepo.FindRolesByUserID(userID)
	if err != nil {
		return nil, err
	}

	roles := make([]vo.RoleInfo, 0, len(userRoles))
	roleIDs := make([]uint, 0, len(userRoles))
	for _, r := range userRoles {
		roles = append(roles, vo.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code})
		roleIDs = append(roleIDs, r.ID)
	}

	buttons := make([]string, 0)
	menuInfos := make([]vo.MenuInfo, 0)

	if len(roleIDs) > 0 {
		menus, err := s.menuRepo.FindMenusByRoleIDs(roleIDs)
		if err == nil {
			for _, m := range menus {
				if m.Type == common.MenuTypeButton && m.Permission != "" {
					buttons = append(buttons, m.Permission)
				}
			}
			menuTree := buildMenuTree(menus, 0)
			menuInfos = convertToMenuInfo(menuTree)
		}
	}

	return &vo.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Phone:    user.Phone,
		Roles:    roles,
		Buttons:  buttons,
		Menus:    menuInfos,
	}, nil
}

func buildMenuTree(menus []model.SysMenu, parentID uint) []model.SysMenu {
	tree := make([]model.SysMenu, 0)
	for _, menu := range menus {
		if menu.ParentID == parentID {
			children := buildMenuTree(menus, menu.ID)
			menu.Children = children
			tree = append(tree, menu)
		}
	}
	return tree
}

func convertToMenuInfo(menus []model.SysMenu) []vo.MenuInfo {
	result := make([]vo.MenuInfo, 0, len(menus))
	for _, m := range menus {
		info := vo.MenuInfo{
			ID:        m.ID,
			ParentID:  m.ParentID,
			Name:      m.Name,
			Path:      m.Path,
			Component: m.Component,
			Redirect:  m.Redirect,
			Icon:      m.Icon,
			Title:     m.Title,
			Type:      m.Type,
			Sort:      m.Sort,
			IsCache:   m.IsCache,
			Visible:   m.Visible,
			Children:  convertToMenuInfo(m.Children),
		}
		result = append(result, info)
	}
	return result
}
