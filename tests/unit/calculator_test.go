package unit

import (
	"fmt"
	"testing"

	"github.com/ivanov-nikolay/distributed_calculator/internal/models"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/calculator"
	"github.com/stretchr/testify/assert"
)

func TestComputeTask(t *testing.T) {
	tests := []struct {
		task     models.Task
		expected float64
	}{
		{models.Task{"1", 2, 3, "+", 0}, 5},
		{models.Task{"2", 5, 2, "-", 0}, 3},
		{models.Task{"3", 4, 3, "*", 0}, 12},
		{models.Task{"4", 10, 2, "/", 0}, 5},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%0.2f %s %0.2f", tt.task.Arg1, tt.task.Operation, tt.task.Arg2), func(t *testing.T) {
			result := calculator.ComputeTask(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}
