package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Orchestrator структура, содержащая конфигурационные параметры оркестратора
type Orchestrator struct {
	ServerPort            string
	TimeAdditionMS        int
	TimeSubtractionMS     int
	TimeMultiplicationsMS int
	TimeDivisionsMS       int
}

// Agent структура, содержащая конфигурационные параметры агента
type Agent struct {
	ComputingPower int
}

// ServerPort конфигурация севера
type ServerPort struct {
	ServerPort string
}

// LoadConfigOrchestrator загружает параметры для запуска сервера
func LoadConfigOrchestrator() *Orchestrator {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
		return nil
	}

	port, exists := os.LookupEnv("SERVER_PORT")
	if !exists {
		port = ":8080"
	}
	timeAdditionMS, exists := os.LookupEnv("TIME_ADDITION_MS")
	if !exists {
		timeAdditionMS = "1000"
	}
	timeSubtructionMS, exists := os.LookupEnv("TIME_SUBTRACTION_MS")
	if !exists {
		timeSubtructionMS = "1000"
	}
	timeMultiplicationsMS, exists := os.LookupEnv("TIME_MULTIPLIER_MS")
	if !exists {
		timeMultiplicationsMS = "2000"
	}
	timeDivisionsMS, exists := os.LookupEnv("TIME_DIVISION_MS")
	if !exists {
		timeDivisionsMS = "2000"
	}

	timeAddition, err := strconv.ParseInt(timeAdditionMS, 10, 64)
	if err != nil {
		log.Fatalf("error parsing TIME_ADDITION_MS: %v", err)
	}
	timeSubtraction, err := strconv.ParseInt(timeSubtructionMS, 10, 64)
	if err != nil {
		log.Fatalf("error parsing TIME_SUBTRACTION_MS: %v", err)
	}
	timeMultiplications, err := strconv.ParseInt(timeMultiplicationsMS, 10, 64)
	if err != nil {
		log.Fatalf("error parsing TIME_MULTIPLIER_MS: %v", err)
	}
	timeDivisions, err := strconv.ParseInt(timeDivisionsMS, 10, 64)
	if err != nil {
		log.Fatalf("error parsing TIME_DIVISION_MS: %v", err)
	}

	return &Orchestrator{
		ServerPort:            port,
		TimeAdditionMS:        int(timeAddition),
		TimeSubtractionMS:     int(timeSubtraction),
		TimeMultiplicationsMS: int(timeMultiplications),
		TimeDivisionsMS:       int(timeDivisions),
	}
}

// LoadConfigAgent загружает параметры для запуска сервера
func LoadConfigAgent() *Agent {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	computingPower, exists := os.LookupEnv("COMPUTING_POWER")
	if !exists {
		computingPower = "4"
	}

	computingPowerInt, err := strconv.Atoi(computingPower)
	if err != nil {
		log.Fatalf("error parsing COMPUTING_POWER: %v", err)
	}

	return &Agent{
		ComputingPower: computingPowerInt,
	}
}

// LoadServerPort загружает конфигурацию сервера
func LoadServerPort() *ServerPort {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	port, exists := os.LookupEnv("SERVER_PORT")
	if !exists {
		port = ":8080"
	}
	return &ServerPort{
		ServerPort: port,
	}
}
