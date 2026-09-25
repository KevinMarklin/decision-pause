package handler

import (
	"net/http"
	"strings"
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/auth"
)

func TestRecover_PanicBecomes500(t *testing.T) {
	svc := &fakeService{panicPing: true}
	rec := do(newRouter(t, svc, Options{}), http.MethodGet, "/health", "", "", nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"internal"`) {
		t.Errorf("тело после паники: %s", rec.Body)
	}
}

func TestSecurityHeaders_Nosniff(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, Options{}), http.MethodGet, "/health", "", "", nil)
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("нет X-Content-Type-Options")
	}
}

func TestCORS_PreflightAllowed(t *testing.T) {
	router := newRouter(t, &fakeService{}, Options{CORS: []string{"https://app.example.com"}})
	rec := do(router, http.MethodOptions, "/api/decisions", "", "", map[string]string{
		"Origin":                        "https://app.example.com",
		"Access-Control-Request-Method": "POST",
	})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO = %q", got)
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), auth.HeaderInitData) {
		t.Errorf("нет %s в Allowed-Headers: %q",
			auth.HeaderInitData, rec.Header().Get("Access-Control-Allow-Headers"))
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("нет Allowed-Methods")
	}
}

func TestCORS_PreflightWithoutAllowlist(t *testing.T) {
	router := newRouter(t, &fakeService{}, Options{})
	rec := do(router, http.MethodOptions, "/api/decisions", "", "",
		map[string]string{"Origin": "https://app.example.com", "Access-Control-Request-Method": "POST"})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("без CORS_ORIGINS заголовок ставиться не должен")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	router := newRouter(t, &fakeService{}, Options{CORS: []string{"https://app.example.com"}})
	rec := do(router, http.MethodGet, "/health", "", "", map[string]string{"Origin": "https://evil.example.com"})

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("чужой origin должен быть отклонён")
	}
}

func TestAuth_RequireMode401(t *testing.T) {
	router := newRouter(t, &fakeService{}, Options{
		Auth: auth.Config{BotToken: "123456:token", Require: true},
	})
	rec := do(router, http.MethodGet, "/api/decisions", "", "", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Errorf("body = %s", rec.Body)
	}
}

func TestAuth_RequireModeInvalidSignatureRejected(t *testing.T) {
	router := newRouter(t, &fakeService{}, Options{
		Auth: auth.Config{BotToken: "123456:token", Require: true},
	})
	rec := do(router, http.MethodGet, "/api/decisions", "", "", map[string]string{
		auth.HeaderInitData: "auth_date=1&hash=deadbeef",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401 (невалидная подпись)", rec.Code)
	}
}

func TestHealth_PublicInRequireMode(t *testing.T) {
	// /health проверяют балансировщик и мониторинг — они не шлют initData.
	router := newRouter(t, &fakeService{}, Options{
		Auth: auth.Config{BotToken: "123456:token", Require: true},
	})
	rec := do(router, http.MethodGet, "/health", "", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 в прод-режиме", rec.Code)
	}
}
