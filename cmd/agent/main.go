package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ivanov-nikolay/distributed_calculator/internal/infrastructure/config"
	"github.com/ivanov-nikolay/distributed_calculator/internal/usecase"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/http"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/worker"
)

func main() {
	computingPower := config.GetEnvAsInt("COMPUTING_POWER", 1)
	orchestratorURL := config.GetEnv("ORCHESTRATOR_URL", "http://localhost:8080")
	authToken := config.GetEnv("AUTH_TOKEN", "")

	taskClient := http.NewTaskHTTPClient(orchestratorURL, authToken)
	calculator := usecase.NewCalculatorUseCase(taskClient)
	agent := worker.NewAgent(computingPower, calculator)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		agent.Stop()
		os.Exit(0)
	}()

	agent.Run()
}
