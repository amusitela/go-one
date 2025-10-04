package conf

import (
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

// JWTClaims JWT声明
type JWTClaims struct {
	UserID string `json:"user_id"`
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

// GenerateJWT 生成JWT token
func GenerateJWT(userID string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT.AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT.Secret))
}

// GenerateRefreshToken 生成刷新token
func GenerateRefreshToken(userID string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT.RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT.Secret))
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
