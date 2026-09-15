package service

import (
	"context"
	"errors"
	"fmt"
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
	// Logout 吊销指定 refresh token。
	// refreshToken 为空时退化为吊销该用户的全部 refresh token
	// （无法判断来源设备，宁可多吊销，也不留下可继续换发 access token 的凭据）。
	Logout(userID uint, refreshToken string) error
	GetUserInfo(userID uint) (*vo.UserInfoResponse, error)
}

type authService struct {
	userRepo    repository.UserRepository
	roleService RoleService
	menuService MenuService
}

func NewAuthService() AuthService {
	return &authService{
		userRepo:    repository.NewUserRepository(),
		roleService: NewRoleService(),
		menuService: NewMenuService(),
	}
}

func (s *authService) Login(req *dto.LoginRequest) (*vo.LoginResponse, error) {
	user, err := s.userRepo.FindByUsernameForAuth(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 与密码错误返回相同信息，避免通过错误提示枚举用户名
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 先校验密码，再检查账号状态，避免未通过验证即暴露账号状态
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status == common.StatusDisabled {
		return nil, errors.New("用户已被禁用")
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
	refreshTTL := time.Duration(config.Cfg.JWT.RefreshExpire) * time.Second
	if err := cache.Set(ctx, cache.RefreshTokenKey(refreshToken), user.ID, refreshTTL); err != nil {
		return nil, fmt.Errorf("存储refresh token失败: %w", err)
	}
	if err := registerRefreshToken(ctx, user.ID, refreshToken, refreshTTL); err != nil {
		return nil, err
	}

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
	claims, err := auth.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token无效")
	}

	ctx := context.Background()

	// 用户维度已被吊销（改密/禁用）时拒绝续期。
	// revokeUserTokens 会同时清掉集合里的 refresh token，这里是第二道防线：
	// 万一清理有遗漏，也不至于让一个已被停用的账号重新换出 access token。
	if revoked, _ := cache.Exists(ctx, fmt.Sprintf("user:token_revoked:%d", claims.UserID)); revoked {
		return nil, errors.New("token已失效，请重新登录")
	}

	exists, _ := cache.Exists(ctx, cache.RefreshTokenKey(req.RefreshToken))
	if !exists {
		return nil, errors.New("refresh token已过期")
	}

	// 轮换：旧 token 立即作废并从用户集合中移除
	_ = cache.Del(ctx, cache.RefreshTokenKey(req.RefreshToken))
	if claims.UserID > 0 {
		_ = cache.SRem(ctx, cache.RefreshTokenSetKey(claims.UserID), req.RefreshToken)
	}

	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.Username, claims.TenantID, claims.DeptID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(claims.UserID, claims.Username, claims.TenantID, claims.DeptID)
	if err != nil {
		return nil, err
	}

	refreshTTL := time.Duration(config.Cfg.JWT.RefreshExpire) * time.Second
	if err := cache.Set(ctx, cache.RefreshTokenKey(refreshToken), claims.UserID, refreshTTL); err != nil {
		return nil, fmt.Errorf("存储refresh token失败: %w", err)
	}
	if err := registerRefreshToken(ctx, claims.UserID, refreshToken, refreshTTL); err != nil {
		return nil, err
	}

	return &vo.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.Cfg.JWT.AccessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) Logout(userID uint, refreshToken string) error {
	ctx := context.Background()

	if refreshToken != "" {
		if err := cache.Del(ctx, cache.RefreshTokenKey(refreshToken)); err != nil {
			return err
		}
		if userID > 0 {
			_ = cache.SRem(ctx, cache.RefreshTokenSetKey(userID), refreshToken)
		}
		return nil
	}

	// 未携带 refreshToken（旧客户端）：无法定位具体会话，
	// 退化为吊销该用户全部 refresh token
	if userID == 0 {
		return nil
	}
	tokens, err := cache.SMembers(ctx, cache.RefreshTokenSetKey(userID))
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		keys = append(keys, cache.RefreshTokenKey(t))
	}
	keys = append(keys, cache.RefreshTokenSetKey(userID))
	return cache.Del(ctx, keys...)
}

// registerRefreshToken 把 refresh token 登记到用户维度的集合中。
//
// 没有这个索引就无法按用户批量吊销：token 本身是随机串，改密或禁用用户时
// 无从得知该用户签发过哪些 token。
func registerRefreshToken(ctx context.Context, userID uint, token string, ttl time.Duration) error {
	setKey := cache.RefreshTokenSetKey(userID)
	if err := cache.SAdd(ctx, setKey, token); err != nil {
		return fmt.Errorf("登记refresh token失败: %w", err)
	}
	// 集合本身不单独续期：略微放宽，确保晚签发的 token 不会因集合过期而漏吊销
	return cache.Expire(ctx, setKey, ttl+24*time.Hour)
}

func (s *authService) GetUserInfo(userID uint) (*vo.UserInfoResponse, error) {
	user, err := s.userRepo.FindByID(0, userID)
	if err != nil {
		return nil, err
	}

	roleIDs, err := s.userRepo.FindRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	roles := make([]vo.RoleInfo, 0, len(roleIDs))
	if len(roleIDs) > 0 {
		userRoles, err := s.roleService.FindByIDs(0, roleIDs)
		if err == nil {
			for _, r := range userRoles {
				roles = append(roles, vo.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code})
			}
		}
	}

	buttons := make([]string, 0)
	menuInfos := make([]vo.MenuInfo, 0)

	if len(roleIDs) > 0 {
		menus, err := s.menuService.FindMenusByRoleIDs(roleIDs)
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
