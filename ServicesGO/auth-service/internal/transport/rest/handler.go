package rest

import (
	"encoding/json"
	"healthmates/auth-service/internal/service"
	"healthmates/auth-service/utils"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	auth *service.AuthService
}

func NewHandler(authSvc *service.AuthService) *Handler {
	return &Handler{auth: authSvc}
}

// Возвращает маршрутизатор с маршрутам и middleware
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	// Валидация JSON + время ожидания
	r.Use(JSONContentType)
	r.Use(Timeout(5 * time.Second))

	r.Post("/auth/register", h.handleRegister)
	r.Post("/auth/refresh", h.handleTokenRefresh)
	r.Post("/auth/login", h.handleLogin)
	r.Get("/auth/google", h.handleGoogleAuth)
	r.Get("/auth/google/callback", h.handleGoogleCallback)
	r.HandleFunc("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})
	r.HandleFunc("/docs", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently).ServeHTTP)
	r.Handle("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))
	return r
}

// Middleware: ставим Content-Type: application/json
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Middleware: таймаут на контексте
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, `{"code":"timeout","message":"request timed out"}`)
	}
}

//
// DTOs
//

type registerDTO struct {
	Email    string `json:"email"    validate:"required,email"`
	Phone    string `json:"phone"    validate:"required_without=Email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginDTO struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

//
// Handlers with Swagger annotations
//

// @Summary      Регистрация пользователя
// @Description  Создаёт новый аккаунт и возвращает access+refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerDTO  true  "Registration Data"
// @Success      201   {object}  map[string]string  "{"access_token":"..."}"
// @Failure      400   {object}  utils.APIError
// @Router       /auth/register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var dto registerDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "cannot parse JSON")
		return
	}
	if err := validate.Struct(dto); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// теперь Register возвращает (access, refresh, error)
	access, refresh, err := h.auth.Register(dto.Email, dto.Phone, dto.Password)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "register_failed", err.Error())
		return
	}

	// ставим refresh-token в HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(service.RefreshTokenTTL),
	})
	// отдаем клиенту access-token
	utils.RespondJSON(w, http.StatusCreated, map[string]string{"access_token": access})
}

// @Summary      Вход в систему
// @Description  Проверяет credentials и выдаёт access+refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginDTO  true  "Login Data"
// @Success      200   {object}  map[string]string  "{"access_token":"..."}"
// @Failure      400   {object}  utils.APIError
// @Failure      401   {object}  utils.APIError
// @Router       /auth/login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var dto loginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "cannot parse JSON")
		return
	}
	if err := validate.Struct(dto); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	access, refresh, err := h.auth.Login(dto.Email, dto.Password)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "login_failed", err.Error())
		return
	}

	// Ставим refresh-token в HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(service.RefreshTokenTTL),
	})
	// Отдаём только access-token в JSON
	utils.RespondJSON(w, http.StatusOK, map[string]string{"access_token": access})
}

// @Summary      Обновление токена
// @Description  Принимает refresh-token из cookie и выдаёт новые токены
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string  "{"access_token":"..."}"
// @Failure      401 {object} utils.APIError
// @Router       /auth/refresh [post]
func (h *Handler) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	// читаем cookie
	c, err := r.Cookie("refresh_token")
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "no_refresh", "refresh token missing")
		return
	}
	newAccess, newRefresh, err := h.auth.Refresh(c.Value)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "refresh_failed", err.Error())
		return
	}
	// ставим новую cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(service.RefreshTokenTTL),
	})
	utils.RespondJSON(w, http.StatusOK, map[string]string{"token": newAccess})
}
