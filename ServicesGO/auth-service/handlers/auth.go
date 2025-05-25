package handlers

import (
	"encoding/json"
	"fmt"
	"healthmates/auth-service/models"
	"healthmates/auth-service/utils"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	googleOauthConfig = &oauth2.Config{
		ClientID:     "SECRET",
		ClientSecret: "SECRET",
		RedirectURL:  "http://localhost:8081/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
)

type LoginRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Provider string `json:"provider"` // google, yandex, etc.
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Provider string `json:"provider"`
}

// Функция для регистрации
func Register(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, что предоставлен хотя бы один способ связи
	if req.Email == "" && req.Phone == "" {
		http.Error(w, "Email or phone is required", http.StatusBadRequest)
		return
	}

	// Проверяем пароль
	if req.Password == "" && req.Provider == "" {
		http.Error(w, "Password is required for local registration", http.StatusBadRequest)
		return
	}

	// Получаем роль пользователя
	var role models.Role
	if err := db.First(&role, "role_name = ?", "User").Error; err != nil {
		http.Error(w, "Role not found", http.StatusInternalServerError)
		return
	}

	// Хэшируем пароль, если он предоставлен
	var hashedPassword string
	var err error
	if req.Password != "" {
		hashedPassword, err = utils.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Could not hash password", http.StatusInternalServerError)
			return
		}
	}

	// Создаем нового пользователя
	user := models.User{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
		RoleId:   role.Id,
	}

	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.Email, user.Phone)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"token":   token,
	})
}

// Функция для логина
func Login(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	query := db.Model(&models.User{})

	// Ищем пользователя по email или телефону
	if req.Email != "" {
		query = query.Where("email = ?", req.Email)
	} else if req.Phone != "" {
		query = query.Where("phone = ?", req.Phone)
	} else {
		http.Error(w, "Email or phone is required", http.StatusBadRequest)
		return
	}

	if err := query.First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль, если это не социальная аутентификация
	if req.Provider == "" {
		if req.Password == "" {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}

		if !utils.CheckPasswordHash(req.Password, user.Password) {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
	}

	// Обновляем время последнего входа
	user.LastLogin = time.Now()
	db.Save(&user)

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.Email, user.Phone)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func GoogleAuth(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Генерируем URL для аутентификации через Google
	url := googleOauthConfig.AuthCodeURL("state-token")

	// Перенаправляем пользователя на страницу входа Google
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func OAuthCallback(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Проверяем state token для безопасности
	state := r.FormValue("state")
	if state != "state-token" {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	// Получаем код авторизации
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем access token для отладки
	log.Printf("Access token: %s", token.AccessToken)

	// Получаем информацию о пользователе с помощью нового endpoint
	client := googleOauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
		GoogleID string `json:"sub"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, существует ли пользователь
	var user models.User
	result := db.Where("email = ?", userInfo.Email).First(&user)

	if result.Error != nil {
		// Если пользователь не существует, создаем нового
		var role models.Role
		if err := db.First(&role, "role_name = ?", "User").Error; err != nil {
			http.Error(w, "Role not found", http.StatusInternalServerError)
			return
		}

		user = models.User{
			Email:     userInfo.Email,
			RoleId:    role.Id,
			LastLogin: time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}

		// Создаем запись о социальной аутентификации
		socialAuth := models.SocialAuth{
			UserId:     user.Id,
			Provider:   "google",
			ProviderId: userInfo.GoogleID,
		}

		if err := db.Create(&socialAuth).Error; err != nil {
			http.Error(w, "Could not create social auth", http.StatusInternalServerError)
			return
		}
	} else {
		// Обновляем время последнего входа
		user.LastLogin = time.Now()
		db.Save(&user)
	}

	// Генерируем JWT токен
	jwtToken, err := utils.GenerateJWT(user.Email, user.Phone)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Редиректим на специальный URL для мобильного приложения
	redirectUrl := fmt.Sprintf("https://close-window/?token=%s", jwtToken)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}
