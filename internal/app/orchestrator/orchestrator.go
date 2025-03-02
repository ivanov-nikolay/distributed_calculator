package orchestrator

import (
	"log"
	"net/http"

	"github.com/ivanov-nikolay/distributed_calculator/internal/api/http/orchestrator"
	"github.com/ivanov-nikolay/distributed_calculator/internal/config"
)

// operationTimes хранилище времени выполнения математических операций
var operationTimes = map[string]int{}

// ApplicationOrchestrator содержит конфигурацию оркестратора
type ApplicationOrchestrator struct {
	orchestrator *config.Orchestrator
}

// NewApplicationOrchestrator создает новый экземпляр ApplicationOrchestrator
func NewApplicationOrchestrator() *ApplicationOrchestrator {
	return &ApplicationOrchestrator{
		orchestrator: config.LoadConfigOrchestrator(),
	}
}

// RunApplicationOrchestrator запускает оркестратор
func (a *ApplicationOrchestrator) RunApplicationOrchestrator() {
	operationTimes["+"] = a.orchestrator.TimeAdditionMS
	operationTimes["-"] = a.orchestrator.TimeSubtractionMS
	operationTimes["*"] = a.orchestrator.TimeMultiplicationsMS
	operationTimes["/"] = a.orchestrator.TimeDivisionsMS

	http.HandleFunc("/api/v1/calculate", orchestrator.HandleCalculate)
	http.HandleFunc("/api/v1/expressions", orchestrator.HandleGetExpressions)
	http.HandleFunc("/api/v1/expressions/", orchestrator.HandleGetExpressionByID)
	http.HandleFunc("/internal/task", orchestrator.HandleTask)

	log.Println("orchestrator is running on :8080")
	log.Fatal(http.ListenAndServe(a.orchestrator.ServerPort, nil))
}
