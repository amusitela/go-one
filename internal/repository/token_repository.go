package repository

import (
    "time"

    "go-one/internal/model"
    "gorm.io/gorm"
)

type RefreshTokenRepository interface {
    Create(jti string, userID uint, expiresAt time.Time, rotatedFrom string) error
    FindByJTI(jti string) (*model.RefreshToken, error)
    RevokeByJTI(jti string) error
}

type refreshTokenRepository struct {
    db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
    return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(jti string, userID uint, expiresAt time.Time, rotatedFrom string) error {
    rt := &model.RefreshToken{
        JTI:         jti,
        UserID:      userID,
        ExpiresAt:   expiresAt,
        Revoked:     false,
        RotatedFrom: rotatedFrom,
    }
    return r.db.Create(rt).Error
}

func (r *refreshTokenRepository) FindByJTI(jti string) (*model.RefreshToken, error) {
    var rt model.RefreshToken
    if err := r.db.Where("jti = ?", jti).First(&rt).Error; err != nil {
        return nil, err
    }
    return &rt, nil
}

func (r *refreshTokenRepository) RevokeByJTI(jti string) error {
    return r.db.Model(&model.RefreshToken{}).Where("jti = ?", jti).Update("revoked", true).Error
}

