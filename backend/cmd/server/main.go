package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KevinMarklin/decision-pause/backend/internal/auth"
	"github.com/KevinMarklin/decision-pause/backend/internal/bot"
	"github.com/KevinMarklin/decision-pause/backend/internal/config"
	"github.com/KevinMarklin/decision-pause/backend/internal/handler"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := repository.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := repository.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := handler.New(
		service.New(repository.New(pool)),
		handler.Options{
			Auth: auth.Config{
				BotToken: cfg.MaxBotToken,
				Require:  cfg.RequireInitData,
			},
			CORS: cfg.CorsOrigins,
		},
	)

	if cfg.MaxBotToken != "" {
		go func() {
			if err := bot.Run(ctx, cfg.MaxBotToken); err != nil && ctx.Err() == nil {
				log.Printf("max bot stopped: %v", err)
			}
		}()
	} else {
		log.Printf("max bot: MAX_BOT_TOKEN is not set, bot is disabled")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("anti-impulse api listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
