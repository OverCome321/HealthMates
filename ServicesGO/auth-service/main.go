package main

import (
	"fmt"
	"healthmates/auth-service/handlers"
	"healthmates/auth-service/models"
	"healthmates/auth-service/utils"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Инициализация базы данных
func InitDB() {
	// Создаем базу данных, если она не существует
	if err := utils.InitDatabase(); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	// Получаем параметры подключения из переменных окружения
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "healthmates")

	// Подключаемся к созданной базе данных
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	
	log.Printf("Подключение к базе данных: %s:%s/%s", dbHost, dbPort, dbName)
	
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Автоматически мигрируем модели в базу данных
	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.SocialAuth{})
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	// Создаем роль User, если она не существует
	var role models.Role
	if err := db.First(&role, "role_name = ?", "User").Error; err != nil {
		role = models.Role{RoleName: "User"}
		if err := db.Create(&role).Error; err != nil {
			log.Fatalf("Ошибка создания роли User: %v", err)
		}
	}

	fmt.Println("База данных подключена и мигрирована успешно.")
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Устанавливаем JWT секрет
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	os.Setenv("JWT_SECRET", jwtSecret)

	InitDB()

	// Запуск HTTP сервера и обработка запросов
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test endpoint hit. Database is connected and migrated."))
	})

	// Регистрация по email/телефону
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(w, r, db)
	})

	// Авторизация по email/телефону
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(w, r, db)
	})

	// Авторизация через Google
	http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		handlers.GoogleAuth(w, r, db)
	})

	// Callback для OAuth
	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		handlers.OAuthCallback(w, r, db)
	})

	// Запускаем сервер на порту 8080
	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
