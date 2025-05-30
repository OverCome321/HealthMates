package repository

import (
	"time"

	"healthmates/auth-service/models"

	"gorm.io/gorm"
)

type SocialAuthRepository interface {
	FindByProvider(provider, providerID string) (*models.SocialAuth, error)
	Create(auth *models.SocialAuth) error
	UpdateTokens(id int, access, refresh string, expiry time.Time) error
}

type socialRepoGorm struct{ db *gorm.DB }

func NewSocialAuthRepository(db *gorm.DB) SocialAuthRepository {
	return &socialRepoGorm{db: db}
}

func (r *socialRepoGorm) FindByProvider(provider, providerID string) (*models.SocialAuth, error) {
	var s models.SocialAuth
	err := r.db.
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *socialRepoGorm) Create(auth *models.SocialAuth) error {
	return r.db.Create(auth).Error
}

func (r *socialRepoGorm) UpdateTokens(id int, access, refresh string, expiry time.Time) error {
	return r.db.Model(&models.SocialAuth{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"access_token":  access,
			"refresh_token": refresh,
			"token_expiry":  expiry,
		}).Error
}
