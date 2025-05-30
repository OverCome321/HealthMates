package main

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"healthmates/auth-service/config"
	"healthmates/auth-service/internal/repository"
	"healthmates/auth-service/internal/service"
	"healthmates/auth-service/internal/transport/rest"
)

// Статусы по умолчанию
type statusRecorder struct {
	stdhttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	// === Prometheus метрики ===
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP-запросов",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP-запроса в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	prometheus.MustRegister(requestCount, requestDuration)

	// === База данных и сервисы ===
	db, err := gorm.Open(postgres.Open(cfg.BuildDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	userRepo := repository.NewUserRepository(db)
	socialRepo := repository.NewSocialAuthRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db) // новый репозиторий
	roleRepo := repository.NewRoleRepository(db)
	authSvc := service.NewAuthService(userRepo, socialRepo, refreshRepo, roleRepo, cfg)
	h := rest.NewHandler(authSvc)

	// === Структурированный zap-логгер ===
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// === Router и middleware ===
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httprate.LimitByIP(8, 1*time.Minute))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Метрики + расширенный логгер
	r.Use(func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)

			path := r.URL.Path
			method := r.Method
			status := strconv.Itoa(rec.status)
			duration := time.Since(start).Seconds()

			// Prometheus
			requestCount.WithLabelValues(method, path, status).Inc()
			requestDuration.WithLabelValues(method, path).Observe(duration)

			// zap
			logger.Info("request",
				zap.String("request_id", middleware.GetReqID(r.Context())),
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", rec.status),
				zap.Float64("duration_s", duration),
			)
		})
	})

	// Собственно `/metrics`
	r.Handle("/metrics", promhttp.Handler())

	// Привязываем наши HTTP-роуты
	r.Mount("/", h.Router())

	log.Println("Server listening on port:", cfg.ServerPort)
	log.Fatal(stdhttp.ListenAndServe(":"+cfg.ServerPort, r))
}
