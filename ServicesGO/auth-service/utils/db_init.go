package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// InitDatabase подключается к базе данных
func InitDatabase() error {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "sa")
	dbPassword := getEnv("DB_PASSWORD", "admin")
	dbName := getEnv("DB_NAME", "healthmates")

	// Подключаемся к базе данных
	connStr := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbPort, dbName)
	
	log.Printf("Подключение к PostgreSQL: %s:%s", dbHost, dbPort)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к postgres: %v", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("ошибка проверки подключения: %v", err)
	}

	log.Printf("Успешное подключение к базе данных %s", dbName)
	return nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 