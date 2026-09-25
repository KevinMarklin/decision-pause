package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/KevinMarklin/decision-pause/backend/internal/auth"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

// Service — зависимости хендлеров. Интерфейс, чтобы юнит-тесты подменяли
// сервис фейком и не требовали БД.
type Service interface {
	Create(ctx context.Context, maxUserID *int64, in model.Input) (model.Decision, error)
	Get(ctx context.Context, id string) (model.Decision, error)
	List(ctx context.Context, maxUserID *int64, limit int) ([]model.DecisionListItem, error)
	Delete(ctx context.Context, id string, maxUserID *int64) error
	Ping(ctx context.Context) error
}

var _ Service = (*service.DecisionService)(nil)

// Options — конфигурация HTTP-слоя.
type Options struct {
	Auth auth.Config
	CORS []string // allowlist origin'ов, см. withCORS
}

type Handler struct {
	svc  Service
	auth auth.Config
	cors []string
}

func New(svc Service, opts Options) *Handler {
	return &Handler{svc: svc, auth: opts.Auth, cors: opts.CORS}
}

// Routes — маршруты API (Go 1.22+ паттерны с path-параметрами).
// /health публичный (балансировщик и мониторинг не носят initData),
// всё под /api/ — за auth-middleware.
// Цепочка middleware: лог → panic-recovery → security → CORS → auth → mux.
func (h *Handler) Routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("POST /api/decisions", h.createDecision)
	api.HandleFunc("GET /api/decisions", h.listDecisions)
	api.HandleFunc("GET /api/decisions/{id}", h.getDecision)
	api.HandleFunc("DELETE /api/decisions/{id}", h.deleteDecision)

	root := http.NewServeMux()
	root.HandleFunc("GET /health", h.health)
	root.Handle("/api/", auth.Middleware(api, h.auth))

	var out http.Handler = root
	out = withCORS(h.cors)(out)
	out = withSecurityHeaders(out)
	out = withRecover(out)
	out = withRequestLog(out)
	return out
}

// health — liveness + проверка доступности БД.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Ping(r.Context()); err != nil {
		slog.Warn("health: db ping failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "db": "down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
