package repository

import (
	"healthmates/auth-service/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(name string) (*models.Role, error)
}

type roleRepoGorm struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepoGorm{db: db}
}

func (r *roleRepoGorm) FindByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Where("role_name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
