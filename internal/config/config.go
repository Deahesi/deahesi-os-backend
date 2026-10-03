package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	LLM    LLMConfig
	Server ServerConfig
}

type LLMConfig struct {
	OpenRouterAPI string
	SystemPrompt  string
	Model         string
}

type ServerConfig struct {
	Port string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env doesnt exist")
	}

	return &Config{
		LLM: LLMConfig{
			OpenRouterAPI: getEnv("OPEN_ROUTER_API_KEY", ""),
			SystemPrompt:  getEnv("LLM_SYSTEM_PROMPT", ""),
			Model:         getEnv("LLM_MODEL", "qwen/qwen3.8-27b:free"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
	}
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}

func getEnvInt(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return def
	}

	return n
}
