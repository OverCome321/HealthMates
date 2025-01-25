package models

import (
	"time"
)

type User struct {
	Id           int       `gorm:"primaryKey;autoIncrement"`
	Login        string    `gorm:"size:100;not null;unique"`
	HashPassword string    `gorm:"size:255;not null"`
	CreatedDate  time.Time `gorm:"autoCreateTime"` // Автоматически устанавливается время создания
	RoleId       int       `gorm:"not null"`
	Role         Role      `gorm:"foreignKey:RoleId"`
}
