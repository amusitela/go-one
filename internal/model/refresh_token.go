package model

import "time"

// RefreshToken 用于持久化和旋转的刷新令牌记录
type RefreshToken struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    JTI         string    `gorm:"uniqueIndex;size:64;not null" json:"jti"`
    UserID      uint      `gorm:"index;not null" json:"user_id"`
    ExpiresAt   time.Time `json:"expires_at"`
    Revoked     bool      `gorm:"default:false" json:"revoked"`
    RotatedFrom string    `gorm:"size:64" json:"rotated_from"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

