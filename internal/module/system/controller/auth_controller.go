package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-admin/internal/cache"
	"go-admin/internal/common"
	captchaModel "go-admin/internal/module/captcha/model"
	captchaService "go-admin/internal/module/captcha/service"
	"go-admin/internal/module/system/dto"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
	"go-admin/internal/module/system/service"

	"github.com/gin-gonic/gin"
)

const (
	maxLoginAttempts = 5
	loginLockDuration = 15 * time.Minute
)

type AuthController struct {
	authService    service.AuthService
	logRepo        repository.LogRepository
	captchaService captchaService.CaptchaService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService:    service.NewAuthService(),
		logRepo:        repository.NewLogRepository(),
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
			Token  string           `json:"captchaToken"`
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

	// 登录限频：同一 IP 5分钟内最多5次失败
	ctx := context.Background()
	ip := c.ClientIP()
	rateKey := fmt.Sprintf("login:rate:%s", ip)
	attempts, _ := cache.Incr(ctx, rateKey)
	if attempts == 1 {
		_ = cache.Expire(ctx, rateKey, loginLockDuration)
	}
	if attempts > maxLoginAttempts {
		ctl.saveLoginLog(c, req.Username, 0, "登录频率过高")
		common.Error(c, common.CodeBadRequest, fmt.Sprintf("登录失败次数过多，请%d分钟后再试", int(loginLockDuration.Minutes())))
		return
	}

	resp, err := ctl.authService.Login(&req)
	if err != nil {
		ctl.saveLoginLog(c, req.Username, 0, err.Error())
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	// 登录成功，清除限频计数
	_ = cache.Del(ctx, rateKey)

	ctl.saveLoginLog(c, req.Username, 1, "登录成功")
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
		token = token[7:]
	}
	_ = ctl.authService.Logout(token)
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

func (ctl *AuthController) saveLoginLog(c *gin.Context, username string, status int8, msg string) {
	ua := c.Request.UserAgent()
	browser := parseUA(ua, []string{"Chrome", "Firefox", "Safari", "Edge", "Opera"})
	os := parseUA(ua, []string{"Windows", "Mac OS X", "Linux", "Android", "iOS"})

	log := &model.SysLoginLog{
		TenantID:  common.GetTenantID(c),
		Username:  username,
		IP:        c.ClientIP(),
		Browser:   browser,
		OS:        os,
		Status:    status,
		Msg:       msg,
		LoginTime: time.Now(),
	}
	_ = ctl.logRepo.CreateLoginLog(log)
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
