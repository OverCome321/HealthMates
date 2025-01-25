package main

import (
	"healthmates/auth-service/models"
	"healthmates/auth-service/utils"
	"testing"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// Функция для инициализации базы данных в тестах
func InitTestDB() (*gorm.DB, error) {
	dsn := "sqlserver://admin:admin@10.10.101.5:1433?database=HealthMatesDB"
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Автоматически мигрируем модели в базу данных для тестов
	err = db.AutoMigrate(&models.User{}, &models.Role{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Функция для удаления тестового пользователя
func DeleteTestUser(db *gorm.DB, login string) error {
	var user models.User
	if err := db.First(&user, "login = ?", login).Error; err != nil {
		return err
	}

	// Удаляем пользователя
	if err := db.Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

// Тест для создания пользователя и авторизации
func TestUserCreationAndLogin(t *testing.T) {
	// Инициализируем базу данных
	db, err := InitTestDB()
	if err != nil {
		t.Fatalf("Could not connect to the database: %v", err)
	}

	// Удаляем тестового пользователя, если он существует
	err = DeleteTestUser(db, "testuser")
	if err != nil {
		t.Logf("No existing test user found or error deleting: %v", err)
	}

	// Создаем пользователя для теста
	user := models.User{
		Login:        "testuser",
		HashPassword: "password123", // Пароль, который будет хэшироваться
	}

	// Получаем роль "User"
	var role models.Role
	if err := db.First(&role, "role_name = ?", "User").Error; err != nil {
		t.Fatalf("Error fetching role: %v", err)
	}

	// Устанавливаем роль для пользователя
	user.RoleId = role.Id

	// Хэшируем пароль перед сохранением
	hashedPassword, err := utils.HashPassword(user.HashPassword)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	user.HashPassword = hashedPassword

	// Сохраняем пользователя в базе данных
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	// Пытаемся авторизоваться с созданными данными
	loginUser := models.User{
		Login:        "testuser",
		HashPassword: "password123", // Попробуем ввести пароль, который использовался при регистрации
	}

	var storedUser models.User
	if err := db.First(&storedUser, "login = ?", loginUser.Login).Error; err != nil {
		t.Fatalf("Error fetching user: %v", err)
	}

	// Проверяем, совпадает ли хэшированный пароль
	if !utils.CheckPasswordHash(loginUser.HashPassword, storedUser.HashPassword) {
		t.Fatalf("Invalid credentials.")
	} else {
		t.Logf("User logged in successfully.")
	}

	// Удаляем тестового пользователя после теста
	if err := DeleteTestUser(db, "testuser"); err != nil {
		t.Fatalf("Error deleting test user: %v", err)
	} else {
		t.Logf("Test user deleted successfully.")
	}
}
