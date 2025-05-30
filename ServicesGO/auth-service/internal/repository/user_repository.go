package repository

import (
	"time"

	"healthmates/auth-service/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error) // ← добавили
	UpdateLastLogin(id int, when time.Time) error
}

type userRepoGorm struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepoGorm{db: db}
}

func (r *userRepoGorm) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepoGorm) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoGorm) FindByID(id int) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoGorm) UpdateLastLogin(id int, when time.Time) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", id).
		Update("last_login", when).
		Error
}
