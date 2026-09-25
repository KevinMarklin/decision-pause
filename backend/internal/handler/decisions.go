package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KevinMarklin/decision-pause/backend/internal/auth"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

// maxBodyBytes — потолок на тело запроса (анкета на порядки меньше).
const maxBodyBytes = 64 << 10

// uuidPattern — форма id, который выдаёт Postgres (gen_random_uuid).
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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

// protoError — ошибка разбора запроса, уже переведённая в HTTP-код.
type protoError struct {
	status int
	code   string
	field  string // опционально: имя поля для ответа
}

func writeProtoError(w http.ResponseWriter, e *protoError) {
	if e.field != "" {
		writeJSON(w, e.status, map[string]string{"error": e.code, "field": e.field})
		return
	}
	writeError(w, e.status, e.code)
}

// decodeBody — JSON строго: только application/json, лимит размера,
// без неизвестных полей (опечатка в имени поля молча дала бы нулевой расчёт),
// без данных после первого значения.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) *protoError {
	if ct := r.Header.Get("Content-Type"); ct != "" &&
		!strings.HasPrefix(strings.ToLower(ct), "application/json") {
		return &protoError{status: http.StatusUnsupportedMediaType, code: "unsupported_media_type"}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return &protoError{status: http.StatusRequestEntityTooLarge, code: "payload_too_large"}
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return &protoError{status: http.StatusBadRequest, code: "invalid_json", field: typeErr.Field}
		}
		if name, ok := unknownFieldName(err); ok {
			return &protoError{status: http.StatusBadRequest, code: "unknown_field", field: name}
		}
		return &protoError{status: http.StatusBadRequest, code: "invalid_json"}
	}

	if dec.More() {
		return &protoError{status: http.StatusBadRequest, code: "invalid_json"}
	}
	return nil
}

// unknownFieldName достаёт имя поля из ошибки encoding/json.
// Тип ошибки пакет не экспортирует, опираемся на канонический префикс.
func unknownFieldName(err error) (string, bool) {
	const prefix = `json: unknown field "`
	msg := err.Error()
	if !strings.HasPrefix(msg, prefix) || !strings.HasSuffix(msg, `"`) {
		return "", false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(msg, prefix), `"`)
	return name, name != ""
}

func userID(r *http.Request) *int64 {
	return auth.UserIDFromContext(r.Context())
}

func (h *Handler) createDecision(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if e := decodeBody(w, r, &req); e != nil {
		writeProtoError(w, e)
		return
	}
	if req.Type != "" && req.Type != "loan" {
		writeError(w, http.StatusBadRequest, "unsupported decision type")
		return
	}

	d, err := h.svc.Create(r.Context(), userID(r), req.Inputs)
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
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}

	d, err := h.svc.Get(r.Context(), id)
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

// deleteDecision — снимок расчёта удалить можно, изменить нельзя (см. PUT → нет).
func (h *Handler) deleteDecision(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID(r)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("delete decision", "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listDecisions(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100 {
			writeError(w, http.StatusBadRequest, "invalid_limit")
			return
		}
		limit = v
	}

	items, err := h.svc.List(r.Context(), userID(r), limit)
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
