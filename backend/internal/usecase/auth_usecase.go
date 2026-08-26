package usecase

import (
	"context"
	"errors"

	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/hash"
	"homeapp/internal/pkg/jwt"
	"homeapp/internal/repository"
)

type AuthUsecase struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtManager *jwt.Manager) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, jwtManager: jwtManager}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	Token string
	User  *entity.User
}

func (u *AuthUsecase) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	existing, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.ErrEmailAlreadyUsed
	}

	hashed, err := hash.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: &hashed,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := u.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (u *AuthUsecase) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrInvalidCredential
		}
		return nil, err
	}

	if user.PasswordHash == nil || !hash.CheckPassword(in.Password, *user.PasswordHash) {
		return nil, apperror.ErrInvalidCredential
	}

	token, err := u.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}
