package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"healthmates/auth-service/config"
	"healthmates/auth-service/internal/repository"
	"healthmates/auth-service/models"
	"healthmates/auth-service/utils"
)

const (
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type AuthService struct {
	userRepo    repository.UserRepository
	socialRepo  repository.SocialAuthRepository
	refreshRepo repository.RefreshTokenRepository
	roleRepo    repository.RoleRepository
	cfg         *config.Config
}

func NewAuthService(
	u repository.UserRepository,
	s repository.SocialAuthRepository,
	r repository.RefreshTokenRepository,
	role repository.RoleRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{userRepo: u, socialRepo: s, refreshRepo: r, roleRepo: role, cfg: cfg}
}

// Login возвращает access и refresh токены
func (s *AuthService) Login(email, password string) (accessToken, refreshToken string, err error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	if !utils.CheckPasswordHash(password, u.Password) {
		return "", "", errors.New("invalid credentials")
	}
	// Генерируем access token
	accessToken, err = utils.GenerateJWT(u.Email, u.Phone)
	if err != nil {
		return "", "", err
	}
	// Обновляем время входа
	s.userRepo.UpdateLastLogin(u.Id, time.Now())

	// Генерируем refresh token
	b := make([]byte, 32)
	rand.Read(b)
	rt := base64.URLEncoding.EncodeToString(b)
	rtModel := &models.RefreshToken{
		UserID:    u.Id,
		Token:     rt,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}
	if err := s.refreshRepo.Create(rtModel); err != nil {
		return "", "", err
	}

	return accessToken, rt, nil
}

// Refresh отзывает старый refresh и выдаёт новую пару токенов
func (s *AuthService) Refresh(oldRT string) (newAccess, newRefresh string, err error) {
	rtModel, err := s.refreshRepo.Find(oldRT)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if rtModel.Revoked || time.Now().After(rtModel.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}
	// Отменяем старый
	s.refreshRepo.Revoke(rtModel.ID)

	// Берём пользователя
	u, err := s.userRepo.FindByID(rtModel.UserID)
	if err != nil {
		return "", "", err
	}
	// Новый access
	newAccess, err = utils.GenerateJWT(u.Email, u.Phone)
	if err != nil {
		return "", "", err
	}
	// Новый refresh
	b := make([]byte, 32)
	rand.Read(b)
	nf := base64.URLEncoding.EncodeToString(b)
	newRTModel := &models.RefreshToken{
		UserID:    u.Id,
		Token:     nf,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}
	if err := s.refreshRepo.Create(newRTModel); err != nil {
		return "", "", err
	}
	return newAccess, nf, nil
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.cfg.GoogleConfig.AuthCodeURL(state)
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, state, code string) (string, error) {
	tok, err := s.cfg.GoogleConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}
	client := s.cfg.GoogleConfig.Client(ctx, tok)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var ui struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	json.NewDecoder(resp.Body).Decode(&ui)

	sa, err := s.socialRepo.FindByProvider("google", ui.Sub)
	if err != nil {
		// новый пользователь
		user := &models.User{Email: ui.Email, IsActive: true, LastLogin: time.Now()}
		if err := s.userRepo.Create(user); err != nil {
			return "", err
		}
		sa = &models.SocialAuth{UserId: user.Id, Provider: "google", ProviderId: ui.Sub}
		s.socialRepo.Create(sa)
	}
	s.socialRepo.UpdateTokens(sa.Id, tok.AccessToken, tok.RefreshToken, tok.Expiry)
	return utils.GenerateJWT(ui.Email, "")
}

// Register создаёт пользователя и сразу выдаёт пару токенов
func (s *AuthService) Register(email, phone, password string) (accessToken, refreshToken string, err error) {
	// 1) Хешируем пароль
	hash, err := utils.HashPassword(password)
	if err != nil {
		return "", "", err
	}
	role, err := s.roleRepo.FindByName("User")
	if err != nil {
		return "", "", fmt.Errorf("cannot get default role: %w", err)
	}
	// 2) Сохраняем нового пользователя
	user := &models.User{Email: email, Phone: phone, Password: hash, RoleId: role.Id, IsActive: true, CreatedDate: time.Now(), LastLogin: time.Now()}
	if err := s.userRepo.Create(user); err != nil {
		return "", "", err
	}
	// 3) Обновляем время последнего входа
	s.userRepo.UpdateLastLogin(user.Id, time.Now())

	// 4) Генерируем access-токен
	accessToken, err = utils.GenerateJWT(user.Email, user.Phone)
	if err != nil {
		return "", "", err
	}

	// 5) Генерируем refresh-токен
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	rt := base64.URLEncoding.EncodeToString(b)
	rtModel := &models.RefreshToken{
		UserID:    user.Id,
		Token:     rt,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}
	if err := s.refreshRepo.Create(rtModel); err != nil {
		return "", "", err
	}

	return accessToken, rt, nil
}
