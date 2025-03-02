package agent

import (
	"log"
	"sync"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/config"
	"github.com/ivanov-nikolay/distributed_calculator/internal/transport/agent"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/calculator"
)

// ApplicationAgent содержит конфигурацию агента
type ApplicationAgent struct {
	config *config.Agent
}

// NewApplicationAgent создает новый объект ApplicationAgent
func NewApplicationAgent() *ApplicationAgent {
	return &ApplicationAgent{
		config: config.LoadConfigAgent(),
	}
}

// RunApplicationAgent запускает агента
func (a *ApplicationAgent) RunApplicationAgent() {
	var wg sync.WaitGroup
	for i := 0; i < a.config.ComputingPower; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				task, err := agent.FetchTask()
				if err != nil {
					log.Println("error fetching task:", err)
					time.Sleep(1 * time.Second)
					continue
				}

				result := calculator.ComputeTask(task)
				if err := agent.SendResult(task.ID, result); err != nil {
					log.Println("error sending result:", err)
				}
			}
		}()
	}
	wg.Wait()
}
