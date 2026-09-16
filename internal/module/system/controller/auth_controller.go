package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	"go-admin/internal/logger"
	captchaService "go-admin/internal/module/captcha/service"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/service"
	"go-admin/pkg/auth"

	"github.com/gin-gonic/gin"
)

const (
	maxLoginAttempts  = 5
	loginLockDuration = 15 * time.Minute
)

type AuthController struct {
	authService    service.AuthService
	logService     service.LogService
	captchaService captchaService.CaptchaService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService:    service.NewAuthService(),
		logService:     service.NewLogService(),
		captchaService: captchaService.NewCaptchaService(),
	}
}

// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "登录参数"
// @Success 200 {object} common.Response{data=vo.LoginResponse}
// @Router /api/v1/auth/login [post]
func (ctl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	// 人机校验：凭证由 /captcha/verify 校验通过后签发，一次性。
	//
	// 早前的写法在第一次 ShouldBindJSON 之后又绑定一次请求体取验证码坐标 ——
	// 请求体已被首次绑定读尽，二次绑定必然失败且错误被丢弃（Points 恒为空），
	// 于是整个校验分支从未执行过，登录实际没有任何人机校验。
	if !captchaService.ConsumeVerifiedToken(req.CaptchaToken) {
		common.Error(c, common.CodeBadRequest, "验证码无效或已失效，请重新验证")
		return
	}

	// 登录限频：IP 与账号双维度计数。
	// 仅按 IP 计数时攻击者更换 IP 即可绕过，因此同时对账号维度计数。
	ctx := context.Background()
	ipKey := "login:fail:ip:" + common.NormalizeIP(c.ClientIP())
	accountKey := "login:fail:account:" + req.Username

	if loginLocked(ctx, ipKey, accountKey) {
		ctl.saveLoginLog(c, 0, req.Username, 0, "登录频率过高")
		common.Error(c, common.CodeBadRequest,
			fmt.Sprintf("登录失败次数过多，请%d分钟后再试", int(loginLockDuration.Minutes())))
		return
	}

	resp, err := ctl.authService.Login(&req)
	if err != nil {
		recordLoginFailure(ctx, ipKey, accountKey)
		ctl.saveLoginLog(c, 0, req.Username, 0, err.Error())
		// 失败原因按语义区分：凭证错误/账号禁用等业务问题回 400，
		// Redis 不可用等系统问题回 500 —— 后者不该记成「请求参数错误」
		common.FailWith(c, err)
		return
	}

	// 登录成功，清除失败计数
	_ = cache.Del(ctx, ipKey, accountKey)

	// 从签发的 token 解析租户，使登录日志归属到正确租户
	loginTenantID := uint(0)
	if claims, err := auth.ParseToken(resp.AccessToken); err == nil {
		loginTenantID = claims.TenantID
	}
	ctl.saveLoginLog(c, loginTenantID, req.Username, 1, "登录成功")
	common.Success(c, resp)
}

// @Summary 刷新Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "RefreshToken"
// @Success 200 {object} common.Response{data=vo.LoginResponse}
// @Router /api/v1/auth/refresh [post]
func (ctl *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	resp, err := ctl.authService.RefreshToken(&req)
	if err != nil {
		common.Error(c, common.CodeUnauthorized, err.Error())
		return
	}

	common.Success(c, resp)
}

// @Summary 退出登录
// @Tags 认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} common.Response
// @Router /api/v1/auth/logout [post]
func (ctl *AuthController) Logout(c *gin.Context) {
	// 旧客户端不传请求体，绑定失败属预期，忽略
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 {
		accessToken := authHeader[7:]
		// 将 access token 加入黑名单，剩余有效时间作为过期时间
		if claims, err := auth.ParseToken(accessToken); err == nil {
			if claims.ExpiresAt != nil {
				if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
					if err := cache.RevokeToken(context.Background(), accessToken, ttl); err != nil {
						logger.Log.Warnf("token加入黑名单失败: %v", err)
					}
				}
			}
			// 连带吊销 refresh token。
			// 之前这里把 access token 当作 refresh token 去删（键名 refresh_token:<accessToken>），
			// 删的是一个从未存在的键，导致登出后 refresh token 仍可换发新 access token。
			if err := ctl.authService.Logout(claims.UserID, req.RefreshToken); err != nil {
				logger.Log.Warnf("吊销refresh token失败: %v", err)
			}
		}
	}
	common.Success(c, nil)
}

// @Summary 获取用户信息
// @Tags 认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} common.Response{data=vo.UserInfoResponse}
// @Router /api/v1/auth/userInfo [get]
func (ctl *AuthController) GetUserInfo(c *gin.Context) {
	userID := common.GetCurrentUserID(c)
	if userID == 0 {
		common.Unauthorized(c, "未登录")
		return
	}

	resp, err := ctl.authService.GetUserInfo(userID)
	if err != nil {
		common.FailWith(c, err)
		return
	}

	common.Success(c, resp)
}

// loginLocked 判断任一维度的失败次数是否已达上限。
// Redis 不可用（读取报错）时不做限制，避免缓存故障导致正常用户无法登录。
func loginLocked(ctx context.Context, keys ...string) bool {
	for _, key := range keys {
		v, err := cache.Get(ctx, key)
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		if n >= maxLoginAttempts {
			return true
		}
	}
	return false
}

// recordLoginFailure 记录一次登录失败，并在首次失败时设置计数过期时间
func recordLoginFailure(ctx context.Context, keys ...string) {
	for _, key := range keys {
		n, err := cache.Incr(ctx, key)
		if err != nil {
			continue
		}
		if n == 1 {
			_ = cache.Expire(ctx, key, loginLockDuration)
		}
	}
}

// saveLoginLog 记录登录日志；tenantID 为登录用户的租户，未识别时传 0
func (ctl *AuthController) saveLoginLog(c *gin.Context, tenantID uint, username string, status int8, msg string) {
	ua := c.Request.UserAgent()
	browser := parseUA(ua, []string{"Chrome", "Firefox", "Safari", "Edge", "Opera"})
	os := parseUA(ua, []string{"Windows", "Mac OS X", "Linux", "Android", "iOS"})

	log := &model.SysLoginLog{
		TenantID:  tenantID,
		Username:  username,
		IP:        common.NormalizeIP(c.ClientIP()),
		Browser:   browser,
		OS:        os,
		Status:    status,
		Msg:       msg,
		LoginTime: time.Now(),
	}
	if err := ctl.logService.CreateLoginLog(log); err != nil {
		logger.Log.Warnf("记录登录日志失败: %v", err)
	}
}

func parseUA(ua string, keywords []string) string {
	lower := strings.ToLower(ua)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return kw
		}
	}
	return "Unknown"
}
