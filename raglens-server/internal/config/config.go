package config

import (
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Models   ModelConfig
}

type AppConfig struct {
	Name        string
	Environment string
}

type ServerConfig struct {
	Addr string
}

type DatabaseConfig struct {
	URL string
}

type ModelConfig struct {
	OpenAIBaseURL       string
	OpenAIAPIKey        string
	ChatModel           string
	EmbeddingModel      string
	EmbeddingDimensions int
}

func Default() Config {
	return Config{
		App: AppConfig{
			Name:        "RAGLens",
			Environment: "local",
		},
		Server: ServerConfig{
			Addr: ":8080",
		},
		Database: DatabaseConfig{
			URL: "postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable",
		},
		Models: ModelConfig{
			OpenAIBaseURL:       "https://api.openai.com/v1",
			ChatModel:           "gpt-4o-mini",
			EmbeddingModel:      "text-embedding-3-small",
			EmbeddingDimensions: 1536,
		},
	}
}

func Load() Config {
	cfg := Default()

	cfg.App.Environment = envOrDefault("RAGLENS_ENV", cfg.App.Environment)
	cfg.Server.Addr = envOrDefault("RAGLENS_SERVER_ADDR", cfg.Server.Addr)
	cfg.Database.URL = envOrDefault("RAGLENS_DATABASE_URL", cfg.Database.URL)
	cfg.Models.OpenAIBaseURL = envOrDefault("RAGLENS_OPENAI_BASE_URL", cfg.Models.OpenAIBaseURL)
	cfg.Models.OpenAIAPIKey = envOrDefault("RAGLENS_OPENAI_API_KEY", cfg.Models.OpenAIAPIKey)
	cfg.Models.ChatModel = envOrDefault("RAGLENS_CHAT_MODEL", cfg.Models.ChatModel)
	cfg.Models.EmbeddingModel = envOrDefault("RAGLENS_EMBEDDING_MODEL", cfg.Models.EmbeddingModel)
	cfg.Models.EmbeddingDimensions = intEnvOrDefault("RAGLENS_EMBEDDING_DIMENSIONS", cfg.Models.EmbeddingDimensions)

	return cfg
}

func intEnvOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
