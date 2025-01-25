package handlers

import (
	"encoding/json"
	"fmt"
	"healthmates/auth-service/models"
	"healthmates/auth-service/utils"
	"net/http"

	"gorm.io/gorm"
)

// Функция для регистрации
func Register(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var user models.User
	var role models.Role

	// Декодируем данные из тела запроса в структуру
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Получаем роль пользователя из базы данных (по умолчанию роль User)
	if err := db.First(&role, "role_name = ?", "User").Error; err != nil {
		http.Error(w, "Role not found", http.StatusInternalServerError)
		return
	}

	// Устанавливаем роль для пользователя
	user.RoleId = role.Id

	// Хэшируем пароль перед сохранением
	user.HashPassword, err = utils.HashPassword(user.HashPassword)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	// Сохраняем пользователя в базе данных
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s registered successfully", user.Login)
}

// Функция для логина (создание JWT токена)
func Login(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли пользователь в базе данных
	var storedUser models.User
	if err := db.First(&storedUser, "login = ?", user.Login).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем, совпадает ли хэшированный пароль
	if !utils.CheckPasswordHash(user.HashPassword, storedUser.HashPassword) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT
	token, err := utils.GenerateJWT(user.Login)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Отправляем токен в ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
