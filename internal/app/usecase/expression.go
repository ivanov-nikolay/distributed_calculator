package usecase

import (
	"context"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/entity"
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/repository"
)

type ExpressionUseCase struct {
	exprRepo repository.ExpressionRepository
}

func NewExpressionUseCase(exprRepo repository.ExpressionRepository) *ExpressionUseCase {
	return &ExpressionUseCase{
		exprRepo: exprRepo,
	}
}

func (uc *ExpressionUseCase) Add(ctx context.Context, userID int, expr string) (int, error) {
	return uc.exprRepo.Add(ctx, userID, expr)
}

func (uc *ExpressionUseCase) UpdateResult(ctx context.Context, id int, result string) error {
	return uc.exprRepo.UpdateResult(ctx, id, result)
}

func (uc *ExpressionUseCase) GetAll(ctx context.Context, userID int) ([]entity.Expression, error) {
	return uc.exprRepo.GetAll(ctx, userID)
}

func (uc *ExpressionUseCase) GetByID(ctx context.Context, userID, id int) (entity.Expression, error) {
	return uc.exprRepo.GetByID(ctx, userID, id)
}

func (uc *ExpressionUseCase) UpdateStatus(ctx context.Context, exprID string, status string) error {
	return uc.exprRepo.UpdateStatus(ctx, exprID, status)
}
