package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	// CaptchaToken /captcha/verify 校验通过后返回的一次性凭证。
	// 必填：留空意味着攻击者可直接 POST /auth/login 绕过人机校验。
	CaptchaToken string `json:"captchaToken" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	// RefreshToken 待吊销的 refresh token。
	// 登出必须连带吊销它 —— 只拉黑 access token 的话，
	// 持有 refresh token 的人仍可换发新 access token，等于没登出。
	RefreshToken string `json:"refreshToken"`
}
