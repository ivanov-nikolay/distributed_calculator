package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/domain"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/calculation"
)

type TaskHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

func NewTaskHTTPClient(baseURL, authToken string) *TaskHTTPClient {
	return &TaskHTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		authToken:  authToken,
	}
}

func (c *TaskHTTPClient) FetchTask() (*domain.Task, error) {
	log.Println("🔄 Запрашиваем новую задачу у оркестратора")
	req, err := http.NewRequest("GET", c.baseURL+"/internal/task", nil)
	if err != nil {
		return nil, err
	}

	if c.authToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Ошибка при запросе задачи: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Статус ответа: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Тело ответа: %s", string(body))
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var taskResp struct {
		Task domain.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		log.Printf("Ошибка декодирования ответа: %v", err)
		return nil, err
	}

	log.Printf("Получена задача: %+v", taskResp.Task)
	return &taskResp.Task, nil
}

func (c *TaskHTTPClient) SubmitResult(result domain.TaskResult) error {
	log.Printf("📤 Отправляем результат:TaskID=%s, Result=%.2f", result.ID, result.Result)
	payload := map[string]interface{}{
		"id":     result.ID,
		"result": result.Result,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.baseURL+"/internal/task", bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to submit result, status: %d", resp.StatusCode)
	}
	return nil
}

func (c *TaskHTTPClient) ReportInvalidTask(taskID string, operation string) error {
	payload := map[string]interface{}{
		"id":        taskID,
		"operation": operation,
		"reason":    "unsupported operator",
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := c.httpClient.Post(
		c.baseURL+"/internal/task/invalid",
		"application/json",
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return calculation.ErrInvalidTask
	}
	return nil
}
