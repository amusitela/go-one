package service

import (
	"context"
	"go-one/internal/model"
)

// BusinessContext 业务上下文，包含业务逻辑需要的上下文信息
// 替代gin.Context，让service层与HTTP传输层解耦
type BusinessContext struct {
	// 请求上下文
	Context context.Context

	// 用户身份信息（来自JWT token）
	UserUUID string      // JWT中的用户ID
	Claims   *JWTClaims  // 完整的JWT claims
	Account  *model.User // 用户信息

	// 请求元数据
	RequestID   string
	TraceID     string
	ClientIP    string
	UserAgent   string
	RequestTime int64
}

// NewBusinessContext 创建业务上下文
func NewBusinessContext(ctx context.Context) *BusinessContext {
	return &BusinessContext{
		Context: ctx,
	}
}

// WithUserUUID 设置用户ID
func (bc *BusinessContext) WithUserUUID(userID string) *BusinessContext {
	bc.UserUUID = userID
	return bc
}

// WithClaims 设置JWT claims
func (bc *BusinessContext) WithClaims(claims *JWTClaims) *BusinessContext {
	bc.Claims = claims
	if claims != nil {
		bc.UserUUID = claims.UserID
		bc.Account = claims.Account
	}
	return bc
}

// WithAccount 设置用户详情
func (bc *BusinessContext) WithAccount(account *model.User) *BusinessContext {
	bc.Account = account
	return bc
}

// WithRequestID 设置请求ID
func (bc *BusinessContext) WithRequestID(requestID string) *BusinessContext {
	bc.RequestID = requestID
	return bc
}

// WithTraceID 设置追踪ID
func (bc *BusinessContext) WithTraceID(traceID string) *BusinessContext {
	bc.TraceID = traceID
	return bc
}

// WithClientIP 设置客户端IP
func (bc *BusinessContext) WithClientIP(clientIP string) *BusinessContext {
	bc.ClientIP = clientIP
	return bc
}

// WithUserAgent 设置用户代理
func (bc *BusinessContext) WithUserAgent(userAgent string) *BusinessContext {
	bc.UserAgent = userAgent
	return bc
}

// WithRequestTime 设置请求时间
func (bc *BusinessContext) WithRequestTime(requestTime int64) *BusinessContext {
	bc.RequestTime = requestTime
	return bc
}

// IsAuthenticated 检查是否已认证
func (bc *BusinessContext) IsAuthenticated() bool {
	return bc.UserUUID != "" && bc.Claims != nil
}

// GetRequiredAccount 获取必需的账号信息（用于需要认证的接口）
func (bc *BusinessContext) GetRequiredAccount() (*model.User, error) {
	if bc.Account == nil || bc.UserUUID == "" {
		return nil, &AuthError{Message: "用户未认证"}
	}
	return bc.Account, nil
}
