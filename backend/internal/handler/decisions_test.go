package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KevinMarklin/decision-pause/backend/internal/calculator"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
	"github.com/KevinMarklin/decision-pause/backend/internal/service"
)

const (
	testUUID  = "3b6f8a10-0000-4000-8000-000000000001"
	otherUUID = "3b6f8a10-0000-4000-8000-000000000002"
	validBody = `{"type":"loan","inputs":{"loan_amount":2000000,"loan_term_months":36,` +
		`"interest_rate_pct":20,"revenue":800000,"expenses":600000,"reserve":500000,` +
		`"revenue_growth_pct":30,"expense_growth_pct":15,"purpose":"Закупка оборудования"}}`
)

// fakeService — фейк Service: записывает вызовы и отдаёт заготовленные ответы.
type fakeService struct {
	createErr error
	getErr    error
	listErr   error
	deleteErr error
	pingErr   error
	panicPing bool

	items []model.DecisionListItem

	gotUserID *int64
	gotID     string
	gotLimit  int
	gotInput  model.Input
}

func (f *fakeService) Create(_ context.Context, userID *int64, in model.Input) (model.Decision, error) {
	f.gotUserID = userID
	f.gotInput = in
	if f.createErr != nil {
		return model.Decision{}, f.createErr
	}
	return model.Decision{
		ID:        testUUID,
		CreatedAt: time.Date(2026, 9, 25, 1, 9, 21, 0, time.UTC),
		Type:      "loan",
		Status:    "calculated",
		Inputs:    in,
		Result:    calculator.BuildResult(in),
	}, nil
}

func (f *fakeService) Get(_ context.Context, id string) (model.Decision, error) {
	f.gotID = id
	if f.getErr != nil {
		return model.Decision{}, f.getErr
	}
	return model.Decision{
		ID:     id,
		Type:   "loan",
		Inputs: model.Input{LoanAmount: 2_000_000, LoanTermMonths: 36, InterestRatePct: 20, Revenue: 800_000},
		Result: calculator.BuildResult(model.Input{LoanAmount: 2_000_000, LoanTermMonths: 36, InterestRatePct: 20, Revenue: 800_000}),
	}, nil
}

func (f *fakeService) List(_ context.Context, userID *int64, limit int) ([]model.DecisionListItem, error) {
	f.gotUserID = userID
	f.gotLimit = limit
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.items, nil
}

func (f *fakeService) Delete(_ context.Context, id string, userID *int64) error {
	f.gotID = id
	f.gotUserID = userID
	return f.deleteErr
}

func (f *fakeService) Ping(context.Context) error {
	if f.panicPing {
		panic("boom")
	}
	return f.pingErr
}

func newRouter(t *testing.T, svc Service, opts Options) http.Handler {
	t.Helper()
	return New(svc, opts).Routes()
}

func do(h http.Handler, method, target, body, contentType string, headers map[string]string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// --- POST /api/decisions -----------------------------------------------------

func TestCreate_Ok(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(t, svc, Options{}), http.MethodPost, "/api/decisions", validBody,
		"application/json; charset=utf-8", map[string]string{"X-Max-User-Id": "777"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
	body := rec.Body.String()
	for _, want := range []string{`"id":"` + testUUID, `"monthly_payment":74327`, `"checklist"`} {
		if !strings.Contains(body, want) {
			t.Errorf("нет %s в ответе: %s", want, body)
		}
	}
	if svc.gotUserID == nil || *svc.gotUserID != 777 {
		t.Errorf("userID дошёл до сервиса: %v", svc.gotUserID)
	}
	if svc.gotInput.LoanAmount != 2_000_000 {
		t.Errorf("inputs не разобраны: %+v", svc.gotInput)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	svc := &fakeService{createErr: &service.ValidationError{Fields: map[string]string{
		"loan_amount": "должна быть больше 0",
	}}}
	rec := do(newRouter(t, svc, Options{}), http.MethodPost, "/api/decisions", validBody, "application/json", nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"loan_amount"`) {
		t.Errorf("нет карты полей: %s", rec.Body)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions", `{"inputs":`, "application/json", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid_json") {
		t.Errorf("body = %s", rec.Body)
	}
}

func TestCreate_TrailingJSON(t *testing.T) {
	body := validBody + ` {"inputs":{}}`
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions", body, "application/json", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400 (хвост после JSON)", rec.Code)
	}
}

func TestCreate_UnknownField(t *testing.T) {
	// Опечатка в имени поля не должна молча превращаться в нулевой расчёт.
	body := `{"inputs":{"loan_amount":2000000,"loan_term_months":36,"interest_rate":20,` +
		`"revenue":800000,"expenses":0,"reserve":0,"revenue_growth_pct":0,"expense_growth_pct":0}}`
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions", body, "application/json", nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unknown_field") || !strings.Contains(rec.Body.String(), "interest_rate") {
		t.Errorf("ожидался unknown_field для interest_rate: %s", rec.Body)
	}
}

func TestCreate_PayloadTooLarge(t *testing.T) {
	body := `{"inputs":{"purpose":"` + strings.Repeat("a", maxBodyBytes) + `"}}`
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions", body, "application/json", nil)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("code = %d, want 413", rec.Code)
	}
}

func TestCreate_UnsupportedMediaType(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions", "a=1", "text/plain", nil)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("code = %d, want 415", rec.Code)
	}
}

func TestCreate_UnsupportedType(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodPost, "/api/decisions",
		`{"type":"mortgage","inputs":{}}`, "application/json", nil)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "unsupported decision type") {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestCreate_InternalError(t *testing.T) {
	svc := &fakeService{createErr: errors.New("db down")}
	rec := do(newRouter(t, svc, Options{}), http.MethodPost, "/api/decisions", validBody, "application/json", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", rec.Code)
	}
}

// --- GET /api/decisions/{id} -------------------------------------------------

func TestGet_Ok(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(t, svc, Options{}), http.MethodGet, "/api/decisions/"+testUUID, "", "", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.gotID != testUUID {
		t.Errorf("id = %q", svc.gotID)
	}
	if !strings.Contains(rec.Body.String(), `"title"`) {
		t.Errorf("нет заголовков сценариев: %s", rec.Body)
	}
}

func TestGet_InvalidID(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodGet, "/api/decisions/not-a-uuid", "", "", nil)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_id") {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := &fakeService{getErr: repository.ErrNotFound}
	rec := do(newRouter(t, svc, Options{}), http.MethodGet, "/api/decisions/"+otherUUID, "", "", nil)
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "not_found") {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

// --- DELETE /api/decisions/{id} ---------------------------------------------

func TestDelete_Ok(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(t, svc, Options{}), http.MethodDelete, "/api/decisions/"+testUUID, "", "",
		map[string]string{"X-Max-User-Id": "777"})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 должен быть без тела: %q", rec.Body)
	}
	if svc.gotID != testUUID {
		t.Errorf("id = %q", svc.gotID)
	}
	if svc.gotUserID == nil || *svc.gotUserID != 777 {
		t.Errorf("владелец не передан: %v", svc.gotUserID)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := &fakeService{deleteErr: repository.ErrNotFound}
	rec := do(newRouter(t, svc, Options{}), http.MethodDelete, "/api/decisions/"+testUUID, "", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}

func TestDelete_InvalidID(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodDelete, "/api/decisions/xyz", "", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

// --- GET /api/decisions ------------------------------------------------------

func TestList_Ok(t *testing.T) {
	svc := &fakeService{items: []model.DecisionListItem{{
		ID: testUUID, Type: "loan", LoanAmount: 2_000_000, MonthlyPayment: 74_327,
		CashFlows: map[string]float64{"expected": 275_673},
	}}}
	rec := do(newRouter(t, svc, Options{}), http.MethodGet, "/api/decisions", "", "",
		map[string]string{"X-Max-User-Id": "777"})

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.gotLimit != 20 {
		t.Errorf("limit = %d, want 20 (default)", svc.gotLimit)
	}
	if svc.gotUserID == nil || *svc.gotUserID != 777 {
		t.Errorf("userID = %v", svc.gotUserID)
	}
	if !strings.Contains(rec.Body.String(), `"cash_flows"`) {
		t.Errorf("body = %s", rec.Body)
	}
}

func TestList_LimitClampedByValidation(t *testing.T) {
	for _, tc := range []struct {
		raw      string
		wantCode int
	}{
		{"abc", http.StatusBadRequest},
		{"0", http.StatusBadRequest},
		{"101", http.StatusBadRequest},
		{"50", http.StatusOK},
	} {
		rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodGet, "/api/decisions?limit="+tc.raw, "", "", nil)
		if rec.Code != tc.wantCode {
			t.Errorf("limit=%s: code = %d, want %d", tc.raw, rec.Code, tc.wantCode)
		}
	}
}

func TestList_InvalidUserIDHeader(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodGet, "/api/decisions", "", "",
		map[string]string{"X-Max-User-Id": "abc"})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_user_id") {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

// --- GET /health -------------------------------------------------------------

func TestHealth_Ok(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodGet, "/health", "", "", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"db":"ok"`) {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestHealth_DbDown(t *testing.T) {
	svc := &fakeService{pingErr: errors.New("connection refused")}
	rec := do(newRouter(t, svc, Options{}), http.MethodGet, "/health", "", "", nil)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"db":"down"`) {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}
