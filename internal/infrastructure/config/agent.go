package config

type AgentConfig struct {
	ComputingPower  int
	OrchestratorURL string
	AuthToken       string
	APIKey          string
}

func LoadAgentConfig() (*AgentConfig, error) {
	return &AgentConfig{
		ComputingPower:  GetEnvAsInt("COMPUTING_POWER", 1),
		OrchestratorURL: GetEnv("ORCHESTRATOR_URL", "http://localhost:8080"),
		AuthToken:       GetEnv("AUTH_TOKEN", ""),
		APIKey:          GetEnv("INTERNAL_API_KEY", ""),
	}, nil
}
