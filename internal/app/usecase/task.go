package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/entity"
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/repository"
)

type TaskUseCase struct {
	exprRepo    repository.ExpressionRepository
	taskStore   map[string]*entity.Task
	taskQueue   []*entity.Task
	mu          sync.Mutex
	taskCounter int64
	config      *Config
}

type Config struct {
	TimeAddition        int
	TimeSubtraction     int
	TimeMultiplications int
	TimeDivisions       int
}

func NewTaskUseCase(exprRepo repository.ExpressionRepository, config *Config) *TaskUseCase {
	return &TaskUseCase{
		exprRepo:  exprRepo,
		taskStore: make(map[string]*entity.Task),
		taskQueue: make([]*entity.Task, 0),
		config:    config,
	}
}

func (uc *TaskUseCase) GetNextTask(ctx context.Context) (*entity.Task, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if len(uc.taskQueue) == 0 {
		return nil, ErrNoTasksAvailable
	}

	task := uc.taskQueue[0]
	uc.taskQueue = uc.taskQueue[1:]

	if err := uc.exprRepo.UpdateStatus(ctx, task.ExprID, "in_progress"); err != nil {
		return nil, err
	}

	return task, nil
}

func (uc *TaskUseCase) SubmitTaskResult(ctx context.Context, userID int, taskID string, result float64) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	task, ok := uc.taskStore[taskID]
	if !ok {
		return ErrTaskNotFound
	}

	task.Node.IsLeaf = true
	task.Node.Value = result
	delete(uc.taskStore, taskID)

	exprID, err := strconv.Atoi(task.ExprID)
	if err != nil {
		return fmt.Errorf("invalid expression ID: %w", err)
	}

	expr, err := uc.exprRepo.GetByID(ctx, userID, exprID)
	if err != nil {
		return err
	}

	if isExpressionComplete(expr.AST) {
		res := fmt.Sprintf("%f", result)
		if err := uc.exprRepo.UpdateResult(ctx, exprID, res); err != nil {
			return err
		}
		if err := uc.exprRepo.UpdateStatus(ctx, task.ExprID, "completed"); err != nil {
			return err
		}
		log.Printf("Expression %s completed with result: %f", task.ExprID, result)
	} else {
		expr, err := uc.exprRepo.GetByID(ctx, userID, exprID)
		if err != nil {
			return err
		}
		uc.ScheduleTasks(ctx, &expr)
	}

	return nil
}

func (uc *TaskUseCase) ScheduleTasks(ctx context.Context, expr *entity.Expression) {
	var traverse func(node *entity.ASTNode) bool
	traverse = func(node *entity.ASTNode) bool {
		if node == nil || node.IsLeaf {
			return true
		}

		leftReady := traverse(node.Left)
		rightReady := traverse(node.Right)

		if leftReady && rightReady && !node.TaskScheduled {
			opTime, err := uc.getOperationTime(node.Operator)
			if err != nil {
				log.Printf("Error getting operation time: %v", err)
				return false
			}

			uc.taskCounter++
			taskID := fmt.Sprintf("%d", uc.taskCounter)

			var arg1, arg2 float64
			if node.Left != nil && node.Left.IsLeaf {
				arg1 = node.Left.Value
			}
			if node.Right != nil && node.Right.IsLeaf {
				arg2 = node.Right.Value
			}

			task := &entity.Task{
				ID:            taskID,
				ExprID:        expr.ID,
				Arg1:          arg1,
				Arg2:          arg2,
				Operation:     node.Operator,
				OperationTime: opTime,
				Node:          node,
			}

			node.TaskScheduled = true
			uc.taskStore[taskID] = task
			uc.taskQueue = append(uc.taskQueue, task)

			log.Printf("Scheduled task: %s %v %s %v = ? (Time: %dms)",
				task.Arg1, task.Operation, task.Arg2, task.OperationTime)

			return false
		}

		return false
	}

	for {
		scheduled := traverse(expr.AST)
		if scheduled {
			break
		}
	}
}

func isExpressionComplete(node *entity.ASTNode) bool {
	if node == nil {
		return true
	}

	if node.IsLeaf {
		return isExpressionComplete(node.Left) && isExpressionComplete(node.Right)
	}

	return false
}

var (
	ErrNoTasksAvailable = errors.New("no tasks available")
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidOperator  = errors.New("invalid operator")
)

func printAST(node *entity.ASTNode, indent string) {
	if node == nil {
		return
	}

	if node.IsLeaf {
		log.Printf("%sValue: %.2f", indent, node.Value)
	} else {
		log.Printf("%sOperator: %s (Scheduled: %v)", indent, node.Operator, node.TaskScheduled)
		printAST(node.Left, indent+"  ")
		printAST(node.Right, indent+"  ")
	}
}

func (uc *TaskUseCase) getOperationTime(op string) (int, error) {
	switch op {
	case "*", "/":
		return uc.config.TimeMultiplications, nil
	case "+", "-":
		return uc.config.TimeAddition, nil
	default:
		return 0, ErrInvalidOperator
	}
}
