package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func serveWithID(cfg Config, headers map[string]string) (*httptest.ResponseRecorder, *int64) {
	var got *int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	h := Middleware(next, cfg)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, got
}

// withNow подменяет часы middleware и возвращает функцию восстановления.
func withNow(now time.Time) func() {
	prev := timeNow
	timeNow = func() time.Time { return now }
	return func() { timeNow = prev }
}

// fmtID — безопасный вывод id в сообщениях ошибок.
func fmtID(id *int64) string {
	if id == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *id)
}

func TestMiddleware_Require_ValidInitData(t *testing.T) {
	defer withNow(fixtureNow)()
	rec, id := serveWithID(Config{BotToken: fixtureToken, Require: true},
		map[string]string{HeaderInitData: fixtureRaw})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if id == nil || *id != 777 {
		t.Fatalf("id = %s, want 777", fmtID(id))
	}
}

func TestMiddleware_Require_Missing(t *testing.T) {
	rec, id := serveWithID(Config{BotToken: fixtureToken, Require: true}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("code = %d, want 401", rec.Code)
	}
	if id != nil {
		t.Errorf("id = %s, want nil", fmtID(id))
	}
}

func TestMiddleware_Require_InvalidInitData(t *testing.T) {
	defer withNow(fixtureNow)()
	rec, _ := serveWithID(Config{BotToken: fixtureToken, Require: true},
		map[string]string{HeaderInitData: "auth_date=1&hash=deadbeef"})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("code = %d, want 401", rec.Code)
	}
}

func TestMiddleware_Require_IgnoresLegacyHeader(t *testing.T) {
	defer withNow(fixtureNow)()
	rec, id := serveWithID(Config{BotToken: fixtureToken, Require: true},
		map[string]string{
			HeaderInitData: fixtureRaw,
			HeaderUserID:   "999999",
		})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if id == nil || *id != 777 {
		t.Fatalf("id = %s, want 777 (из подписи, не из заголовка)", fmtID(id))
	}
}

func TestMiddleware_Require_WithoutTokenPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("ожидался panic при Require без токена")
		}
	}()
	Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), Config{Require: true})
}

func TestMiddleware_Dev_NoHeaders(t *testing.T) {
	rec, id := serveWithID(Config{}, nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("code = %d, want 204", rec.Code)
	}
	if id != nil {
		t.Errorf("id = %s, want nil", fmtID(id))
	}
}

func TestMiddleware_Dev_LegacyHeader(t *testing.T) {
	rec, id := serveWithID(Config{}, map[string]string{HeaderUserID: "42"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if id == nil || *id != 42 {
		t.Fatalf("id = %s, want 42", fmtID(id))
	}
}

func TestMiddleware_Dev_InvalidLegacyHeader(t *testing.T) {
	for _, raw := range []string{"abc", "-5", "0"} {
		rec, id := serveWithID(Config{}, map[string]string{HeaderUserID: raw})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%q: code = %d, want 400", raw, rec.Code)
		}
		if id != nil {
			t.Errorf("%q: id = %s, want nil", raw, fmtID(id))
		}
	}
}

func TestMiddleware_Dev_InitDataBeatsLegacy(t *testing.T) {
	defer withNow(fixtureNow)()
	rec, id := serveWithID(Config{BotToken: fixtureToken},
		map[string]string{HeaderInitData: fixtureRaw, HeaderUserID: "42"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if id == nil || *id != 777 {
		t.Fatalf("id = %s, want 777 из подписи", fmtID(id))
	}
}

func TestMiddleware_Dev_InvalidInitDataFallsBack(t *testing.T) {
	rec, id := serveWithID(Config{BotToken: fixtureToken},
		map[string]string{HeaderInitData: "auth_date=1&hash=deadbeef", HeaderUserID: "42"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
	if id == nil || *id != 42 {
		t.Fatalf("id = %s, want 42 (legacy)", fmtID(id))
	}
}
