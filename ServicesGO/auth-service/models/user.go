package models

import (
	"time"
)

type User struct {
	Id          int       `gorm:"primaryKey;autoIncrement"`
	Email       string    `gorm:"size:100;unique"`
	Phone       string    `gorm:"size:20;unique"`
	Password    string    `gorm:"size:255"` // Хэшированный пароль
	CreatedDate time.Time `gorm:"autoCreateTime"`
	RoleId      int       `gorm:"not null"`
	Role        Role      `gorm:"foreignKey:RoleId"`
	IsActive    bool      `gorm:"default:true"`
	LastLogin   time.Time
	SocialAuths []SocialAuth `gorm:"foreignKey:UserId"`
}

type SocialAuth struct {
	Id            int       `gorm:"primaryKey;autoIncrement"`
	UserId        int       `gorm:"not null"`
	Provider      string    `gorm:"size:50;not null"` // google, yandex, etc.
	ProviderId    string    `gorm:"size:255;not null"` // ID from provider
	AccessToken   string    `gorm:"size:255"`
	RefreshToken  string    `gorm:"size:255"`
	TokenExpiry   time.Time
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
