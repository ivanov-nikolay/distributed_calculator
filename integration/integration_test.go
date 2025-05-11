//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	deliveryhttp "github.com/ivanov-nikolay/distributed-calculator/internal/app/delivery/http"
	"github.com/ivanov-nikolay/distributed-calculator/internal/app/entity"
	"github.com/ivanov-nikolay/distributed-calculator/internal/app/usecase"
	"github.com/ivanov-nikolay/distributed-calculator/internal/infrastructure/database/sqlite"
	"github.com/ivanov-nikolay/distributed-calculator/pkg/jwt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFullFlow(t *testing.T) {
	// Настройка тестовой БД
	os.Remove("./test.db")
	db, err := sqlite.NewDB("./test.db")
	require.NoError(t, err)
	defer db.Close()

	// Инициализация репозиториев
	exprRepo := sqlite.NewExpressionRepository(db)
	userRepo := sqlite.NewUserRepository(db)

	// Создаем пользователя
	err = userRepo.Create(context.Background(), "test", "test")
	require.NoError(t, err)

	// Инициализация use cases
	jwtUtil := jwt.NewJWTUtil("secret")
	exprUC := usecase.NewExpressionUseCase(exprRepo)
	userUC := usecase.NewUserUseCase(userRepo, jwtUtil)
	taskCfg := &usecase.Config{
		TimeAddition:        10,
		TimeMultiplications: 20,
	}
	taskUC := usecase.NewTaskUseCase(exprRepo, taskCfg)

	// Создаем HTTP handler
	handler := deliveryhttp.NewHandler(exprUC, userUC, taskUC, jwtUtil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Запускаем тестовый сервер
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 1. Логинимся чтобы получить токен
	loginBody := bytes.NewBufferString(`{"login": "test", "password": "test"}`)
	resp, err := http.Post(ts.URL+"/api/v1/login", "application/json", loginBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResponse struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	require.NoError(t, err)
	require.NotEmpty(t, loginResponse.Token)

	// 2. Отправляем выражение на вычисление
	calcBody := bytes.NewBufferString(`{"expression": "2+3*4"}`)
	req, err := http.NewRequest("POST", ts.URL+"/api/v1/calculate", calcBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: loginResponse.Token})

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var calcResponse struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&calcResponse)
	require.NoError(t, err)
	require.NotEmpty(t, calcResponse.ID)

	// 3. Проверяем что выражение добавилось
	req, err = http.NewRequest("GET", ts.URL+"/api/v1/expressions", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: loginResponse.Token})

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var exprsResponse struct {
		Expressions []entity.Expression `json:"expressions"`
	}
	err = json.NewDecoder(resp.Body).Decode(&exprsResponse)
	require.NoError(t, err)
	require.Len(t, exprsResponse.Expressions, 1)
	require.Equal(t, "2+3*4", exprsResponse.Expressions[0].Expr)
}
