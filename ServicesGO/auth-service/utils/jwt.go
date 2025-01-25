package utils

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("secret_key") // Ключ для подписи токена (его стоит хранить безопасно)

// Структура для данных, которые будем включать в JWT
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Функция для генерации JWT токена
func GenerateJWT(username string) (string, error) {
	// Определяем время истечения токена
	expirationTime := time.Now().Add(24 * time.Hour)

	// Создаем новый токен с использованием HMAC SHA256 алгоритма
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Создаем новый токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("could not sign the token: %v", err)
	}

	return tokenString, nil
}
