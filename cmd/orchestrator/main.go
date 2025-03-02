package main

import (
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/orchestrator"
)

func main() {
	o := orchestrator.NewApplicationOrchestrator()
	o.RunApplicationOrchestrator()
}
