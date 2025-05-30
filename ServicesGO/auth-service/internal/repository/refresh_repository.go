package repository

import (
	"healthmates/auth-service/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	Find(tokenString string) (*models.RefreshToken, error)
	Revoke(id int) error
}

type refreshRepoGorm struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshRepoGorm{db: db}
}

func (r *refreshRepoGorm) Create(t *models.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *refreshRepoGorm) Find(tokenString string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.Where("token = ?", tokenString).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *refreshRepoGorm) Revoke(id int) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked", true).
		Error
}
