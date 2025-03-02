package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/models"
	"github.com/ivanov-nikolay/distributed_calculator/internal/service"
)

var (
	// expressions хранилище математических выражений
	expressions = make(map[string]models.Expression)
	// tasks хранилище задач
	tasks = make(map[string]models.Task)
	// results хранилище результатов задач
	results = make(map[string]float64)
	// expressionID
	expressionID = 0
	// taskID
	taskID = 0
	// operationTimes
	operationTimes = map[string]int{}
	// expressionMutex мьютекс для синхронизации доступа к хранилищу математических выражений
	expressionMutex = &sync.Mutex{}
	// taskMutex мьютекс для синхронизации доступа к хранилищу задач
	taskMutex = &sync.Mutex{}
	// resultMutex мьютекс для синхронизации доступа к хранилищу математических выражений
	resultMutex = &sync.Mutex{}
)

// HandleCalculate обработчик http-запроса, принимает математическое выражение, возвращает ID
func HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // 405
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data", http.StatusUnprocessableEntity) // 422
		return
	}

	expressionMutex.Lock()
	expressionID++
	id := strconv.Itoa(expressionID)
	expressions[id] = models.Expression{
		ID:     id,
		Expr:   req.Expression,
		Status: models.StatusExpressionPending,
		Result: 0,
	}
	expressionMutex.Unlock()

	// Разбор математического выражения на задачи
	go parseExpressionToTasks(id, req.Expression)

	w.WriteHeader(http.StatusCreated) // 201
	err := json.NewEncoder(w).Encode(map[string]string{"id": id})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
		return
	}
}

// HandleGetExpressions обработчик http-запроса, возвращает полный список описаний математических выражений
func HandleGetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // 405
		return
	}

	expressionMutex.Lock()
	defer expressionMutex.Unlock()

	var exprList []models.Expression
	for _, expr := range expressions {
		exprList = append(exprList, expr)
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string][]models.Expression{"expressions": exprList})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
		return
	}
}

// HandleGetExpressionByID обработчик http-запроса, принимает ID, возвращает описание математического выражения
func HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // 405
		return
	}

	id := r.URL.Path[len("/api/v1/expressions/"):]
	expressionMutex.Lock()
	defer expressionMutex.Unlock()

	expr, exists := expressions[id]
	if !exists {
		http.Error(w, "expression not found", http.StatusNotFound) // 404
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	err := json.NewEncoder(w).Encode(map[string]models.Expression{"expression": expr})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // 500
		return
	}
}

// HandleTask обработчик http-запроса, отдает задачу агенту или принимает результат вычисления задачи от агента
func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // 405
	case http.MethodGet:
		taskMutex.Lock()
		defer taskMutex.Unlock()

		for _, task := range tasks {
			_ = json.NewEncoder(w).Encode(map[string]models.Task{"task": task})
			delete(tasks, task.ID)
			return
		}
		http.Error(w, "no tasks", http.StatusNotFound) // 404

	case http.MethodPost:
		var result models.TaskResult
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity) // 422
			return
		}

		// сохраняем результат вычисления математического выражения в хранилище результатов задач
		resultMutex.Lock()
		results[result.ID] = result.Result
		resultMutex.Unlock()

		w.WriteHeader(http.StatusOK)
	}
}

// parseExpressionToTasks рабирает математическое выражение на задачи
func parseExpressionToTasks(id, expr string) {
	expr = strings.ReplaceAll(expr, " ", "")

	// разделение выражения на токены
	tokens := service.TokenizeExpression(expr)

	// преобразование токенов в обратную польскую запись (Reverse Polish Notation (RPN))
	rpnTokens := service.ShuntingYard(tokens)

	// вычисление RPN и создание задач
	var stack []string
	for _, token := range rpnTokens {
		if service.IsNumber(token) {
			stack = append(stack, token)
		} else if service.IsOperator(token) {
			if len(stack) < 2 {
				log.Println("error: not enough operands for operator", token)
				return
			}

			arg2 := stack[len(stack)-1]
			arg1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			// создание задачи
			taskID++
			task := models.Task{
				ID:            strconv.Itoa(taskID),
				Arg1:          service.ParseNumber(arg1),
				Arg2:          service.ParseNumber(arg2),
				Operation:     token,
				OperationTime: operationTimes[token],
			}

			// сохранение задачи в хранилище задач
			taskMutex.Lock()
			tasks[task.ID] = task
			taskMutex.Unlock()

			// ожидание результата вычисления задачи
			for {
				resultMutex.Lock()
				result, exists := results[task.ID]
				resultMutex.Unlock()

				if exists {
					stack = append(stack, fmt.Sprintf("%f", result))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// сохранение результата вычисления математического выражения
	if len(stack) == 1 {
		result, _ := strconv.ParseFloat(stack[0], 64)
		expressionMutex.Lock()
		expr := expressions[id]
		expr.Status = models.StatusExpressionCompleted
		expr.Result = result
		expressions[id] = expr
		expressionMutex.Unlock()
	}
}
