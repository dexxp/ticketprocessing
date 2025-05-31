package service

import (
	"context"
	"errors"
	"ticketprocessing/internal/auth"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/repository"
	"ticketprocessing/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInternal           = errors.New("internal server error")
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	repo       repository.AuthRepository
	jwtManager *auth.JWTManager
	tokenStore *auth.RedisTokenStore
}

func NewAuthService(repo repository.AuthRepository, jwtManager *auth.JWTManager, tokenStore *auth.RedisTokenStore) AuthService {
	return &authService{
		repo:       repo,
		jwtManager: jwtManager,
		tokenStore: tokenStore,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password string) error {
	if _, err := s.repo.GetUserByEmail(email); err == nil {
		return ErrUserExists
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return ErrInternal
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return ErrInternal
	}

	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return "", ErrInternal
	}

	if err := s.tokenStore.SaveToken(ctx, token, user.ID); err != nil {
		return "", ErrInternal
	}

	return token, nil
}
