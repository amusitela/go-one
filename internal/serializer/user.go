package serializer

import (
	"go-one/internal/model"
	"time"
)

// UserVTO 用户信息 VTO
type UserVTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthTokenVTO 认证令牌响应 VTO（用于注册和登录）
type AuthTokenVTO struct {
	User         *UserVTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

// TokenPairVTO 令牌对 VTO（用于刷新令牌）
type TokenPairVTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserListVTO 用户列表 VTO
type UserListVTO struct {
	List     []*UserVTO `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}


// BuildUserVTO 将 model.User 转换为 UserVTO
func BuildUserVTO(user *model.User) *UserVTO {
	if user == nil {
		return nil
	}
	return &UserVTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
