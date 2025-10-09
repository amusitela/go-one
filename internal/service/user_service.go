package service

import (
    "go-one/internal/model"
    "go-one/internal/repository"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
    userRepo repository.UserRepository
    tokenRepo repository.RefreshTokenRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository) *UserService {
    return &UserService{
        userRepo:  userRepo,
        tokenRepo: tokenRepo,
    }
}

// RegisterDTO 注册请求DTO
type RegisterDTO struct {
	Username string
	Email    string
	Password string
}

// RegisterResult 注册结果
type RegisterResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// Register 用户注册
func (s *UserService) Register(ctx *BusinessContext, dto *RegisterDTO) (*RegisterResult, ServiceError) {
	// 参数验证
	if dto.Username == "" || len(dto.Username) < 3 {
		return nil, &ValidationError{
			Message: "用户名长度至少为3个字符",
			Code:    40000,
		}
	}
	if dto.Password == "" || len(dto.Password) < 6 {
		return nil, &ValidationError{
			Message: "密码长度至少为6个字符",
			Code:    40000,
		}
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(dto.Username); err == nil {
		return nil, &BusinessError{
			Message: "用户名已存在",
			Code:    40009,
		}
	}

	// 检查邮箱是否已存在
	if dto.Email != "" {
		if _, err := s.userRepo.FindByEmail(dto.Email); err == nil {
			return nil, &BusinessError{
				Message: "邮箱已被使用",
				Code:    40009,
			}
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &DatabaseError{
			Message: "密码加密失败",
			Err:     err,
		}
	}

	user := &model.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: string(hashedPassword),
		Nickname: dto.Username,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, &DatabaseError{
			Message: "创建用户失败",
			Err:     err,
		}
	}

    // 生成访问令牌
    userIDStr := strconv.FormatUint(uint64(user.ID), 10)
    accessToken, err := GenerateAccessToken(userIDStr)
    if err != nil {
        return nil, &BusinessError{
            Message: "生成令牌失败",
            Code:    50000,
            Err:     err,
        }
    }
    // 生成并持久化刷新令牌（旋转起点，无上游JTI）
    jti := uuid.NewString()
    if err := s.tokenRepo.Create(jti, user.ID, time.Now().Add(JWT.RefreshTokenExpire), ""); err != nil {
        return nil, &DatabaseError{Message: "保存刷新令牌失败", Err: err}
    }
    refreshToken, err := GenerateRefreshToken(userIDStr, jti)
    if err != nil {
        return nil, &BusinessError{Message: "生成刷新令牌失败", Code: 50000, Err: err}
    }

	return &RegisterResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LoginDTO 登录请求DTO
type LoginDTO struct {
	Username string
	Password string
}

// LoginResult 登录结果
type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// Login 用户登录
func (s *UserService) Login(ctx *BusinessContext, dto *LoginDTO) (*LoginResult, ServiceError) {
	// 参数验证
	dto.Username = strings.TrimSpace(dto.Username)
	dto.Password = strings.TrimSpace(dto.Password)

	if dto.Username == "" {
		return nil, &ValidationError{
			Message: "用户名不能为空",
			Code:    40000,
		}
	}
	if dto.Password == "" {
		return nil, &ValidationError{
			Message: "密码不能为空",
			Code:    40000,
		}
	}

	user, err := s.userRepo.FindByUsername(dto.Username)
	if err != nil {
		return nil, &AuthError{
			Message: "用户名或密码错误",
		}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, &AuthError{
			Message: "用户名或密码错误",
		}
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, &BusinessError{
			Message: "账号已被禁用",
			Code:    40003,
		}
	}

    // 生成访问令牌
    userIDStr := strconv.FormatUint(uint64(user.ID), 10)
    accessToken, err := GenerateAccessToken(userIDStr)
    if err != nil {
        return nil, &BusinessError{
            Message: "生成令牌失败",
            Code:    50000,
            Err:     err,
        }
    }

    // 生成并持久化刷新令牌
    jti := uuid.NewString()
    if err := s.tokenRepo.Create(jti, user.ID, time.Now().Add(JWT.RefreshTokenExpire), ""); err != nil {
        return nil, &DatabaseError{Message: "保存刷新令牌失败", Err: err}
    }
    refreshToken, err := GenerateRefreshToken(userIDStr, jti)
    if err != nil {
        return nil, &BusinessError{Message: "生成刷新令牌失败", Code: 50000, Err: err}
    }

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx *BusinessContext) (*model.User, ServiceError) {
    uid64, err := strconv.ParseUint(ctx.UserUUID, 10, 64)
    if err != nil || uid64 == 0 {
        return nil, &AuthError{Message: "无效的用户ID"}
    }
    user, err := s.userRepo.FindByID(uint(uid64))
    if err != nil {
        return nil, &NotFoundError{
            Message: "用户不存在",
        }
    }
    return user, nil
}

// UpdateProfileDTO 更新资料请求DTO
type UpdateProfileDTO struct {
	Nickname string
	Avatar   string
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx *BusinessContext, dto *UpdateProfileDTO) ServiceError {
    uid64, err := strconv.ParseUint(ctx.UserUUID, 10, 64)
    if err != nil || uid64 == 0 {
        return &AuthError{Message: "无效的用户ID"}
    }
    user, err := s.userRepo.FindByID(uint(uid64))
    if err != nil {
        return &NotFoundError{
            Message: "用户不存在",
        }
    }

	// 只更新提供的字段
	if dto.Nickname != "" {
		user.Nickname = strings.TrimSpace(dto.Nickname)
	}
	if dto.Avatar != "" {
		user.Avatar = strings.TrimSpace(dto.Avatar)
	}

	if err := s.userRepo.Update(user); err != nil {
		return &DatabaseError{
			Message: "更新用户信息失败",
			Err:     err,
		}
	}

	return nil
}

// ChangePasswordDTO 修改密码请求DTO
type ChangePasswordDTO struct {
	OldPassword string
	NewPassword string
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx *BusinessContext, dto *ChangePasswordDTO) ServiceError {
	// 参数验证
	if dto.OldPassword == "" {
		return &ValidationError{
			Message: "原密码不能为空",
			Code:    40000,
		}
	}
	if dto.NewPassword == "" || len(dto.NewPassword) < 6 {
		return &ValidationError{
			Message: "新密码长度至少为6个字符",
			Code:    40000,
		}
	}

    uid64, err := strconv.ParseUint(ctx.UserUUID, 10, 64)
    if err != nil || uid64 == 0 {
        return &AuthError{Message: "无效的用户ID"}
    }
    user, err := s.userRepo.FindByID(uint(uid64))
    if err != nil {
        return &NotFoundError{
            Message: "用户不存在",
        }
    }

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.OldPassword)); err != nil {
		return &AuthError{
			Message: "原密码错误",
		}
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return &DatabaseError{
			Message: "密码加密失败",
			Err:     err,
		}
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return &DatabaseError{
			Message: "更新密码失败",
			Err:     err,
		}
	}

	return nil
}

// ListUsersQuery 用户列表查询参数
type ListUsersQuery struct {
	Page     int
	PageSize int
}

// ListUsersResult 用户列表结果
type ListUsersResult struct {
	List     []model.User
	Total    int64
	Page     int
	PageSize int
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx *BusinessContext, query *ListUsersQuery) (*ListUsersResult, ServiceError) {
	// 参数校验和默认值
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	users, total, err := s.userRepo.List(query.Page, query.PageSize)
	if err != nil {
		return nil, &DatabaseError{
			Message: "查询用户列表失败",
			Err:     err,
		}
	}

	return &ListUsersResult{
		List:     users,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// RefreshTokenDTO 刷新令牌请求DTO
type RefreshTokenDTO struct {
	RefreshToken string
}

// RefreshTokenResult 刷新令牌结果
type RefreshTokenResult struct {
    AccessToken  string
    RefreshToken string
}

// RefreshToken 刷新访问令牌
func (s *UserService) RefreshToken(ctx *BusinessContext, dto *RefreshTokenDTO) (*RefreshTokenResult, ServiceError) {
	// 参数验证
	if dto.RefreshToken == "" {
		return nil, &ValidationError{
			Message: "刷新令牌不能为空",
			Code:    40000,
		}
	}

    // 验证refresh token 签名与类型
    claims, err := ValidateRefreshToken(dto.RefreshToken)
    if err != nil {
        return nil, &AuthError{
            Message: "刷新令牌无效或已过期",
        }
    }

    // 校验JTI是否存在、未撤销且未过期（持久化校验）
    if claims.JTI == "" {
        return nil, &AuthError{Message: "无效的刷新令牌标识"}
    }
    record, recErr := s.tokenRepo.FindByJTI(claims.JTI)
    if recErr != nil || record == nil {
        return nil, &AuthError{Message: "刷新令牌不存在或已撤销"}
    }
    if record.Revoked || time.Now().After(record.ExpiresAt) {
        return nil, &AuthError{Message: "刷新令牌已失效"}
    }

    // 获取用户信息
    userID, err := strconv.ParseUint(claims.UserID, 10, 64)
    if err != nil {
        return nil, &AuthError{
            Message: "无效的用户ID",
        }
    }

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return nil, &NotFoundError{
			Message: "用户不存在",
		}
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, &BusinessError{
			Message: "账号已被禁用",
			Code:    40003,
		}
	}

    // 撤销当前refresh token并旋转生成新的refresh token
    _ = s.tokenRepo.RevokeByJTI(claims.JTI)
    newJTI := uuid.NewString()
    if err := s.tokenRepo.Create(newJTI, user.ID, time.Now().Add(JWT.RefreshTokenExpire), claims.JTI); err != nil {
        return nil, &DatabaseError{Message: "保存新刷新令牌失败", Err: err}
    }

    // 下发新的访问令牌与刷新令牌
    accessToken, err := GenerateAccessToken(strconv.FormatUint(uint64(user.ID), 10))
    if err != nil {
        return nil, &BusinessError{Message: "生成访问令牌失败", Code: 50000, Err: err}
    }
    refreshToken, err := GenerateRefreshToken(strconv.FormatUint(uint64(user.ID), 10), newJTI)
    if err != nil {
        return nil, &BusinessError{Message: "生成刷新令牌失败", Code: 50000, Err: err}
    }

	return &RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LogoutDTO 登出请求DTO（撤销刷新令牌）
type LogoutDTO struct {
    RefreshToken string
}

// Logout 撤销刷新令牌（登出）
func (s *UserService) Logout(ctx *BusinessContext, dto *LogoutDTO) ServiceError {
    if dto.RefreshToken == "" {
        return &ValidationError{Message: "刷新令牌不能为空", Code: 40000}
    }
    claims, err := ValidateRefreshToken(dto.RefreshToken)
    if err != nil {
        return &AuthError{Message: "刷新令牌无效或已过期"}
    }
    if claims.JTI == "" {
        return &AuthError{Message: "无效的刷新令牌标识"}
    }
    // 标记撤销（幂等）
    if err := s.tokenRepo.RevokeByJTI(claims.JTI); err != nil {
        return &DatabaseError{Message: "撤销刷新令牌失败", Err: err}
    }
    return nil
}
