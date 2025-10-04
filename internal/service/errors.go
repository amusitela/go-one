package service

import "fmt"

// ServiceError 服务层错误接口
type ServiceError interface {
	error
	GetCode() int
	GetMessage() string
}

// ValidationError 参数验证错误
type ValidationError struct {
	Message string
	Code    int
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) GetCode() int {
	return e.Code
}

func (e *ValidationError) GetMessage() string {
	return e.Message
}

// DatabaseError 数据库错误
type DatabaseError struct {
	Message string
	Err     error
}

func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DatabaseError) GetCode() int {
	return 50001 // 数据库错误码
}

func (e *DatabaseError) GetMessage() string {
	return e.Message
}

// ExternalAPIError 外部API调用错误
type ExternalAPIError struct {
	Message string
	Err     error
}

func (e *ExternalAPIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ExternalAPIError) GetCode() int {
	return 50002 // 外部API错误码
}

func (e *ExternalAPIError) GetMessage() string {
	return e.Message
}

// AuthError 认证相关错误
type AuthError struct {
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AuthError) GetCode() int {
	return 40001 // 认证错误码
}

func (e *AuthError) GetMessage() string {
	return e.Message
}

// NotFoundError 资源未找到错误
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (e *NotFoundError) GetCode() int {
	return 40004 // 未找到错误码
}

func (e *NotFoundError) GetMessage() string {
	return e.Message
}

// BusinessError 业务逻辑错误
type BusinessError struct {
	Message string
	Code    int
	Err     error
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *BusinessError) GetCode() int {
	return e.Code
}

func (e *BusinessError) GetMessage() string {
	return e.Message
}
