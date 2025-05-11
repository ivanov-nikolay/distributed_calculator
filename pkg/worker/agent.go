package worker

import (
	"log"
	"sync"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/usecase"
)

type Agent struct {
	workers    int
	calculator *usecase.CalculatorUseCase
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

func NewAgent(workers int, calculator *usecase.CalculatorUseCase) *Agent {
	return &Agent{
		workers:    workers,
		calculator: calculator,
		stopChan:   make(chan struct{}),
	}
}

func (a *Agent) worker(workerID int) {
	defer a.wg.Done()

	retryCount := 0
	maxRetries := 5

	for {
		select {
		case <-a.stopChan:
			log.Printf("Worker %d: stopping", workerID)
			return
		default:
			if err := a.calculator.ProcessTask(); err != nil {
				log.Printf("Worker %d: error: %v", workerID, err)
				retryCount++

				if retryCount >= maxRetries {
					time.Sleep(5 * time.Second)
					retryCount = 0
				}
				continue
			}
			retryCount = 0
			log.Printf("Worker %d processed task successfully", workerID)
		}
	}
}

func (a *Agent) Run() {
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker(i)
	}
	log.Printf("Agent started with %d workers", a.workers)
	a.wg.Wait()
}

func (a *Agent) Stop() {
	close(a.stopChan)
	a.wg.Wait()
	log.Println("Agent stopped gracefully")
}
