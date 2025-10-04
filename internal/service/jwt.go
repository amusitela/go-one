package service

import (
	"fmt"
	"go-one/internal/model"
	"go-one/util"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration
}

var JWT *JWTConfig

// TokenType token类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID    string      `json:"user_id"`
	TokenType TokenType   `json:"token_type"`
	Account   *model.User `json:"account,omitempty"` // refresh token不包含完整用户信息
	jwt.RegisteredClaims
}

// InitJWT 初始化JWT配置
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		util.Log().Warning("JWT_SECRET 未配置，使用默认值（不安全）")
		secret = "default_jwt_secret_key"
	}

	accessExpire := int64(3600) // 默认1小时
	if v := os.Getenv("JWT_ACCESS_TOKEN_EXPIRE"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			accessExpire = parsed
		}
	}

	refreshExpire := int64(604800) // 默认7天
	if v := os.Getenv("JWT_REFRESH_TOKEN_EXPIRE"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			refreshExpire = parsed
		}
	}

	JWT = &JWTConfig{
		Secret:             secret,
		AccessTokenExpire:  time.Duration(accessExpire) * time.Second,
		RefreshTokenExpire: time.Duration(refreshExpire) * time.Second,
	}

	util.Log().Info("JWT配置初始化完成")
}

// GenerateAccessToken 生成访问令牌（包含完整用户信息）
func GenerateAccessToken(user *model.User) (string, error) {
	claims := JWTClaims{
		UserID:    fmt.Sprintf("%d", user.ID),
		TokenType: AccessToken,
		Account:   user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT.AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT.Secret))
}

// GenerateRefreshToken 生成刷新令牌（仅包含用户ID）
func GenerateRefreshToken(userID string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		TokenType: RefreshToken,
		Account:   nil, // refresh token不包含用户信息
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT.RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT.Secret))
}

// GenerateTokenPair 生成token对（access + refresh）
func GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	userID := fmt.Sprintf("%d", user.ID)

	accessToken, err = GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseJWT 解析JWT token
func ParseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// ValidateAccessToken 验证访问令牌
func ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	return claims, nil
}
