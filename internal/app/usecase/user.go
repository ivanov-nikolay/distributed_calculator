package usecase

import (
	"context"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/repository"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	exprRepo repository.UserRepository
	jwtUtil  *jwt.JWTUtil
}

func NewUserUseCase(exprRepo repository.UserRepository, jwtUtil *jwt.JWTUtil) *UserUseCase {
	return &UserUseCase{
		exprRepo: exprRepo,
		jwtUtil:  jwtUtil,
	}
}

func (uc *UserUseCase) Create(ctx context.Context, login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return uc.exprRepo.Create(ctx, login, string(hashedPassword))
	//return uc.exprRepo.Create(ctx, login, password)
}

func (uc *UserUseCase) Login(ctx context.Context, login, password string) (int, error) {
	//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//if err != nil {
	//	return "", err
	//}
	return uc.exprRepo.Login(ctx, login, password)
	//return uc.exprRepo.Login(ctx, login, string(hashedPassword))
}

func (uc *UserUseCase) GetID(ctx context.Context, login string) (int, error) {
	return uc.exprRepo.GetID(ctx, login)
}

func (uc *UserUseCase) Authenticate(ctx context.Context, login, password string) (bool, error) {
	return uc.exprRepo.Authenticate(ctx, login, password)
}
