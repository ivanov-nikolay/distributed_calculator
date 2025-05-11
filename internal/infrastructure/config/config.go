package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP           HTTPConfig
	DB             DatabaseConfig
	JWT            JWTConfig
	Task           TaskConfig
	InternalAPIKey string `envconfig:"INTERNAL_API_KEY" required:"true"`
}

type HTTPConfig struct {
	Port            string        `envconfig:"HTTP_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"15s"`
	ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"5s"`
}

type DatabaseConfig struct {
	Path         string `envconfig:"DB_PATH" default:"calculator.db"`
	MaxOpenConns int    `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int    `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
}

type JWTConfig struct {
	SecretKey      string        `envconfig:"JWT_SECRET_KEY" required:"true"`
	TokenLifetime  time.Duration `envconfig:"JWT_TOKEN_LIFETIME" default:"24h"`
	CookieName     string        `envconfig:"JWT_COOKIE_NAME" default:"auth_token"`
	CookieSecure   bool          `envconfig:"JWT_COOKIE_SECURE" default:"true"`
	CookieHTTPOnly bool          `envconfig:"JWT_COOKIE_HTTP_ONLY" default:"true"`
}

type TaskConfig struct {
	AdditionTime       time.Duration `envconfig:"TASK_ADDITION_TIME" default:"100ms"`
	SubtractionTime    time.Duration `envconfig:"TASK_SUBTRACTION_TIME" default:"100ms"`
	MultiplicationTime time.Duration `envconfig:"TASK_MULTIPLICATION_TIME" default:"100ms"`
	DivisionTime       time.Duration `envconfig:"TASK_DIVISION_TIME" default:"100ms"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	cfg.HTTP.Port = GetEnv("HTTP_PORT", "8080")

	readTimeout, err := time.ParseDuration(GetEnv("HTTP_READ_TIMEOUT", "10s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_READ_TIMEOUT: %w", err)
	}
	cfg.HTTP.ReadTimeout = readTimeout

	writeTimeout, err := time.ParseDuration(GetEnv("HTTP_WRITE_TIMEOUT", "10s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_WRITE_TIMEOUT: %w", err)
	}
	cfg.HTTP.WriteTimeout = writeTimeout

	cfg.DB.Path = GetEnv("DB_PATH", "/data/calculator.db")

	maxConns, err := parseInt(GetEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}
	cfg.DB.MaxOpenConns = maxConns

	cfg.JWT.SecretKey = GetEnv("JWT_SECRET_KEY", "your-secure-key-here")

	tokenLifetime, err := time.ParseDuration(GetEnv("JWT_TOKEN_LIFETIME", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TOKEN_LIFETIME: %w", err)
	}
	cfg.JWT.TokenLifetime = tokenLifetime

	additionTime, err := time.ParseDuration(GetEnv("TASK_ADDITION_TIME", "100ms"))
	if err != nil {
		return nil, fmt.Errorf("invalid TASK_ADDITION_TIME: %w", err)
	}
	cfg.Task.AdditionTime = additionTime

	subtractionTime, err := time.ParseDuration(GetEnv("TASK_SUBTRACTION_TIME", "100ms"))
	if err != nil {
		return nil, fmt.Errorf("invalid TASK_SUBTRACTION_TIME: %w", err)
	}
	cfg.Task.SubtractionTime = subtractionTime

	cfg.InternalAPIKey = GetEnv("INTERNAL_API_KEY", "")

	return cfg, nil
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func GetEnvAsInt(key string, fallback int) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	return fallback
}
