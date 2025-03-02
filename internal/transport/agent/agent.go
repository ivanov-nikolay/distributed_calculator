package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ivanov-nikolay/distributed_calculator/internal/config"
	"github.com/ivanov-nikolay/distributed_calculator/internal/models"
)

// FetchTask запрашивает задачу у оркестратора
func FetchTask() (models.Task, error) {
	port := config.LoadServerPort()

	resp, err := http.Get("http://localhost" + port.ServerPort + "/internal/task")
	if err != nil {
		return models.Task{}, fmt.Errorf("error fetching task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Task{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var received models.TaskReceived

	if err := json.NewDecoder(resp.Body).Decode(&received); err != nil {
		return models.Task{}, fmt.Errorf("error decoding task: %v", err)
	}

	log.Printf("received task: %+v", received.Task)
	return received.Task, nil
}

// SendResult отправляет оркестратору результат вычисления задачи
func SendResult(taskID string, result float64) error {
	port := config.LoadServerPort()

	data := models.TaskResult{
		ID:     taskID,
		Result: result,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost"+port.ServerPort+"/internal/task",
		"application/json",
		io.NopCloser(strings.NewReader(string(jsonData))))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error send result, status code: %d", resp.StatusCode)
	}

	return nil
}
