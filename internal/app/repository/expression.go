package repository

import (
	"context"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/entity"
)

type ExpressionRepository interface {
	Add(ctx context.Context, userID int, expr string) (int, error)
	UpdateResult(ctx context.Context, id int, result string) error
	GetAll(ctx context.Context, userID int) ([]entity.Expression, error)
	GetByID(ctx context.Context, userID, id int) (entity.Expression, error)
	UpdateStatus(ctx context.Context, exprID string, status string) error
}
