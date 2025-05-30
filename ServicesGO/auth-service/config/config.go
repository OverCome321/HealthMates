package config

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config хранит настройки приложения
type Config struct {
	ServerPort                                 string
	DBHost, DBPort, DBUser, DBPassword, DBName string
	JWTSecret                                  string
	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GoogleConfig       *oauth2.Config
}

// Load читает ENV и валидирует
func Load() *Config {
	c := &Config{
		ServerPort:         mustGet("PORT", "8080"),
		DBHost:             mustGet("DB_HOST", "localhost"),
		DBPort:             mustGet("DB_PORT", "5432"),
		DBUser:             mustGet("DB_USER", "sa"),
		DBPassword:         mustGet("DB_PASSWORD", ""),
		DBName:             mustGet("DB_NAME", "healthmates"),
		JWTSecret:          mustGet("JWT_SECRET", ""),
		GoogleClientID:     mustGet("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleClientSecret: mustGet("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		GoogleRedirectURL:  mustGet("GOOGLE_OAUTH_REDIRECT_URL", ""),
	}
	c.GoogleConfig = &oauth2.Config{
		ClientID:     c.GoogleClientID,
		ClientSecret: c.GoogleClientSecret,
		RedirectURL:  c.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return c
}

func mustGet(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def == "" {
		log.Fatalf("required env missing: %s", key)
	}
	return def
}

// BuildDSN возвращает строку подключения Postgres
func (c *Config) BuildDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort,
	)
}
