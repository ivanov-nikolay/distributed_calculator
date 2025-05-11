package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/domain"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/calculation"
)

type CalculatorUseCase struct {
	repo domain.TaskRepository
}

func NewCalculatorUseCase(repo domain.TaskRepository) *CalculatorUseCase {
	return &CalculatorUseCase{repo: repo}
}

func (uc *CalculatorUseCase) ProcessTask() error {
	task, err := uc.repo.FetchTask()
	if err != nil {
		return fmt.Errorf("fetch task failed: %w", err)
	}
	log.Printf("fetched task: %v", task)

	if task.ID == "" || task.Operation == "" {
		log.Println("Received empty task, skipping")
		time.Sleep(1 * time.Second)
		return nil
	}

	if !calculation.SupportedOps[task.Operation] {
		log.Printf("Skipping invalid task (ID: %s, Op: %s\n)", task.ID, task.Operation)
		return uc.repo.ReportInvalidTask(task.ID, task.Operation)
	}

	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	result, err := calculation.Compute(task.Operation, task.Arg1, task.Arg2)
	if err != nil {
		return fmt.Errorf("compute failed: %w", err)
	}

	log.Printf("compute result: %v", result)

	return uc.repo.SubmitResult(domain.TaskResult{
		ID:     task.ID,
		Result: result,
	})
}
