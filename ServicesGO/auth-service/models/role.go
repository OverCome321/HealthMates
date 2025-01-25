package models

type Role struct {
	Id       int    `gorm:"primaryKey;autoIncrement"`
	RoleName string `gorm:"size:100;not null"`
}
