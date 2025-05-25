package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Структура для данных, которые будем включать в JWT
type Claims struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	jwt.RegisteredClaims
}

// Функция для генерации JWT токена
func GenerateJWT(email, phone string) (string, error) {
	// Создаем claims с данными пользователя
	claims := &Claims{
		Email: email,
		Phone: phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Создаем токен с claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Получаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
