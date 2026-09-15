package auth

import (
	"fmt"
	"time"

	"go-admin/config"

	"github.com/golang-jwt/jwt/v5"
)

// Token 签发者标识：用于区分 Access Token 与 Refresh Token，
// 避免长效的 Refresh Token 被当作 Access Token 直接访问受保护接口。
const (
	IssuerAccess  = "go-admin"
	IssuerRefresh = "go-admin-refresh"
)

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	TenantID uint   `json:"tenantId"`
	DeptID   uint   `json:"deptId"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, username string, tenantID, deptID uint) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TenantID: tenantID,
		DeptID:   deptID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Cfg.JWT.AccessExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    IssuerAccess,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

func GenerateRefreshToken(userID uint, username string, tenantID, deptID uint) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TenantID: tenantID,
		DeptID:   deptID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Cfg.JWT.RefreshExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    IssuerRefresh,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

// ParseToken 解析并校验 Access Token。
// 会校验签发者，确保 Refresh Token 无法冒充 Access Token 访问受保护接口。
func ParseToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, IssuerAccess)
}

// ParseRefreshToken 解析并校验 Refresh Token，仅供刷新令牌流程使用。
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, IssuerRefresh)
}

// parseToken 解析 token，并校验签名算法、有效期与签发者
func parseToken(tokenString, expectedIssuer string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 显式限定签名算法，防止 alg 混淆攻击（如 alg=none）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetJWTSecret()), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// 校验签发者，隔离 Access / Refresh 两种 token 的用途
	if claims.Issuer != expectedIssuer {
		return nil, jwt.ErrTokenInvalidIssuer
	}

	return claims, nil
}
