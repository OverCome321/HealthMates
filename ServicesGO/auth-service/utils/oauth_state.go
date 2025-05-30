package utils

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

// GenerateState создаёт случайный state для CSRF-защиты
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetStateCookie сохраняет state в безопасной cookie
func SetStateCookie(w http.ResponseWriter, state string) {
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // в продакшене должно быть true (HTTPS)
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, cookie)
}

// GetStateFromCookie читает state из cookie
func GetStateFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("oauth_state")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
