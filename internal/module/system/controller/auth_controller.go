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
	captchaModel "go-admin/internal/module/captcha/model"
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

	// 验证码校验（如果提供了 captchaToken）
	if req.CaptchaToken != "" {
		// 从请求中获取验证码坐标（前端点击验证码后会提交）
		var captchaReq struct {
			Token  string              `json:"captchaToken"`
			Points []captchaModel.Point `json:"captchaPoints"`
		}
		_ = c.ShouldBindJSON(&captchaReq)

		if len(captchaReq.Points) > 0 {
			resp, err := ctl.captchaService.Verify(captchaReq.Token, captchaReq.Points)
			if err != nil || !resp.Success {
				common.Error(c, common.CodeBadRequest, "验证码错误")
				return
			}
		}
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
		common.Error(c, common.CodeBadRequest, err.Error())
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
	token := c.GetHeader("Authorization")
	if token != "" && len(token) > 7 {
		accessToken := token[7:]
		// 将 access token 加入黑名单，剩余有效时间作为过期时间
		if claims, err := auth.ParseToken(accessToken); err == nil && claims.ExpiresAt != nil {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				if err := cache.RevokeToken(context.Background(), accessToken, ttl); err != nil {
					logger.Log.Warnf("token加入黑名单失败: %v", err)
				}
			}
		}
		_ = ctl.authService.Logout(token[7:])
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
		common.Error(c, common.CodeInternalError, err.Error())
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
