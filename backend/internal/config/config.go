package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	MaxBotToken string
}

func Load() Config {
	loadDotEnv()
	return Config{
		Port:        env("PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://anti:anti@localhost:5432/antipulse?sslmode=disable"),
		MaxBotToken: env("MAX_BOT_TOKEN", ""),
	}
}

// loadDotEnv — читает .env, если он есть (лучший случай, ошибки не критичны).
func loadDotEnv() {
	for _, path := range []string{".env", "backend/.env"} {
		if err := godotenv.Load(path); err == nil {
			return
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
