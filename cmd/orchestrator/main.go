package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "github.com/ivanov-nikolay/distributed_calculator/internal/app/delivery/http"
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/usecase"
	"github.com/ivanov-nikolay/distributed_calculator/internal/infrastructure/config"
	"github.com/ivanov-nikolay/distributed_calculator/internal/infrastructure/database/sqlite"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/jwt"

	"github.com/lpernett/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sqlite.NewDB(cfg.DB.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	exprRepo := sqlite.NewExpressionRepository(db)
	userRepo := sqlite.NewUserRepository(db)

	jwtUtil := jwt.NewJWTUtil(cfg.JWT.SecretKey)

	exprUC := usecase.NewExpressionUseCase(exprRepo)
	userUC := usecase.NewUserUseCase(userRepo, jwtUtil)

	taskSfg := usecase.Config{
		TimeAddition:        int(cfg.Task.AdditionTime.Milliseconds()),
		TimeSubtraction:     int(cfg.Task.SubtractionTime.Milliseconds()),
		TimeMultiplications: int(cfg.Task.MultiplicationTime.Milliseconds()),
		TimeDivisions:       int(cfg.Task.DivisionTime.Milliseconds()),
	}

	taskUC := usecase.NewTaskUseCase(exprRepo, &taskSfg)

	handler := deliveryhttp.NewHandler(exprUC, userUC, taskUC, jwtUtil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(done)
	}()

	log.Printf("Starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe error: %v", err)
	}

	<-done
	log.Println("Server gracefully stopped")
}
