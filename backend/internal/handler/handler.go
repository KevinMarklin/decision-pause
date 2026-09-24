package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

type Handler struct {
	svc *service.DecisionService
}

func New(svc *service.DecisionService) *Handler {
	return &Handler{svc: svc}
}

// Routes — маршруты API (Go 1.22+ паттерны с path-параметрами).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /api/decisions", h.createDecision)
	mux.HandleFunc("GET /api/decisions", h.listDecisions)
	mux.HandleFunc("GET /api/decisions/{id}", h.getDecision)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
