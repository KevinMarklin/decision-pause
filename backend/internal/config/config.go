package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	MaxBotToken string

	// RequireInitData — prod-режим: каждый запрос обязан нести валидную
	// подпись window.WebApp.initData, иначе 401.
	RequireInitData bool

	// CorsOrigins — allowlist origin'ов для Mini App (пусто = CORS выключен).
	CorsOrigins []string
}

func Load() (Config, error) {
	loadDotEnv()
	cfg := Config{
		Port:            env("PORT", "8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://anti:anti@localhost:5432/antipulse?sslmode=disable"),
		MaxBotToken:     env("MAX_BOT_TOKEN", ""),
		RequireInitData: envBool("MAX_REQUIRE_INIT_DATA", false),
		CorsOrigins:     splitList(env("CORS_ORIGINS", "")),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate — конфигурация, без которой запуск бессмыслен.
func (c Config) Validate() error {
	if c.RequireInitData && c.MaxBotToken == "" {
		return fmt.Errorf("MAX_REQUIRE_INIT_DATA=true требует MAX_BOT_TOKEN: " +
			"без токена бота подпись initData проверить невозможно")
	}
	return nil
}

// loadDotEnv — загрузка .env, если файл есть (для прод-окружения переменные
// приходят извне, поэтому ошибки чтения не считаем фатальными).
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

// envBool — булев env: 1/true/yes/on и 0/false/no/off, иначе fallback.
func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

// splitList — "a, b ,c" → ["a","b","c"].
func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
