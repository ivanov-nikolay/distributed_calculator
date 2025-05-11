package repository

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (int, error)
	GetID(ctx context.Context, login string) (int, error)
	Authenticate(ctx context.Context, login, password string) (bool, error)
}
