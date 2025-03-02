package calculator

import (
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/models"
)

// ComputeTask реализует простейший математический калькулятор
func ComputeTask(task models.Task) float64 {
	time.AfterFunc(time.Duration(task.OperationTime)*time.Millisecond, func() {})

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}
