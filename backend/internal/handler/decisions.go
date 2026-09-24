package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

type decisionResponse struct {
	ID        string             `json:"id"`
	CreatedAt time.Time          `json:"created_at"`
	Type      string             `json:"type"`
	Inputs    model.Input        `json:"inputs"`
	Credit    model.CreditResult `json:"credit"`
	Scenarios []model.Scenario   `json:"scenarios"`
	Checklist []string           `json:"checklist"`
}

type listItemResponse struct {
	ID             string             `json:"id"`
	CreatedAt      time.Time          `json:"created_at"`
	Type           string             `json:"type"`
	LoanAmount     float64            `json:"loan_amount"`
	MonthlyPayment float64            `json:"monthly_payment"`
	CashFlows      map[string]float64 `json:"cash_flows"`
}

type createRequest struct {
	Inputs model.Input `json:"inputs"`
	Type   string      `json:"type"`
}

func toDecisionResponse(d model.Decision) decisionResponse {
	return decisionResponse{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		Type:      d.Type,
		Inputs:    d.Inputs,
		Credit:    d.Result.Credit,
		Scenarios: d.Result.Scenarios,
		Checklist: d.Result.Checklist,
	}
}

// maxUserID — опциональный заголовок из Mini App (window.WebApp.initDataUnsafe.user.id).
func maxUserID(r *http.Request) *int64 {
	raw := r.Header.Get("X-Max-User-Id")
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		slog.Warn("invalid X-Max-User-Id header", "value", raw)
		return nil
	}
	return &v
}

func (h *Handler) createDecision(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.Type != "" && req.Type != "loan" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported decision type"})
		return
	}

	d, err := h.svc.Create(r.Context(), maxUserID(r), req.Inputs)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error":  "validation",
				"fields": ve.Fields,
			})
			return
		}
		slog.Error("create decision", "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, toDecisionResponse(d))
}

func (h *Handler) getDecision(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get decision", "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, toDecisionResponse(d))
}

func (h *Handler) listDecisions(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	items, err := h.svc.List(r.Context(), maxUserID(r), limit)
	if err != nil {
		slog.Error("list decisions", "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}

	out := make([]listItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, listItemResponse{
			ID:             it.ID,
			CreatedAt:      it.CreatedAt,
			Type:           it.Type,
			LoanAmount:     it.LoanAmount,
			MonthlyPayment: it.MonthlyPayment,
			CashFlows:      it.CashFlows,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
