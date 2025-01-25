package main

import (
	"fmt"
	"healthmates/auth-service/models"
	"log"
	"net/http"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var db *gorm.DB // Глобальная переменная для хранения соединения с базой данных

// Инициализация базы данных
func InitDB() {
	var err error
	dsn := "sqlserver://admin:admin@10.10.101.5:1433?database=HealthMatesDB"
	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// Автоматически мигрируем модели в базу данных
	err = db.AutoMigrate(&models.User{}, &models.Role{})
	if err != nil {
		log.Fatalf("Could not migrate database: %v", err)
	}

	fmt.Println("Database connected and migrated successfully.")
}

func main() {
	InitDB()

	// Запуск HTTP сервера и обработка запросов
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// Это просто тестовый эндпоинт, если нужно будет его использовать в будущем
		w.Write([]byte("Test endpoint hit. Database is connected and migrated."))
	})

	// Запускаем сервер на порту 8080
	log.Fatal(http.ListenAndServe(":8080", nil))
}
