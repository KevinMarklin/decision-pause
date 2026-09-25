package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// timeNow вынесен в переменную, чтобы тесты подменяли часы.
var timeNow = time.Now

// HeaderInitData — сырое window.WebApp.initData из мини-аппа.
const HeaderInitData = "X-Max-Init-Data"

// HeaderUserID — legacy-заголовок из initDataUnsafe (только dev-режим).
const HeaderUserID = "X-Max-User-Id"

type ctxKey struct{}

// Config — параметры middleware.
//
// Require=true (prod): запрос обязан нести валидную подпись в HeaderInitData,
// иначе 401. Legacy HeaderUserID при этом игнорируется.
//
// Require=false (dev): валидная подпись имеет приоритет, иначе берётся
// HeaderUserID, иначе анонимная история.
type Config struct {
	BotToken string
	Require  bool
}

// Middleware кладёт в контекст id пользователя (*int64, nil — аноним).
func Middleware(next http.Handler, cfg Config) http.Handler {
	if cfg.Require && cfg.BotToken == "" {
		panic("auth: MAX_REQUIRE_INIT_DATA=true требует MAX_BOT_TOKEN")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(HeaderInitData)
		if raw != "" {
			data, err := Validate(raw, cfg.BotToken, timeNow())
			if err == nil {
				id := data.UserID
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, &id)))
				return
			}
			if cfg.Require {
				slog.Warn("init data rejected", "reason", err)
				writeUnauthorized(w)
				return
			}
			slog.Warn("init data rejected, falling back to legacy header", "reason", err)
		} else if cfg.Require {
			writeUnauthorized(w)
			return
		}

		id, bad := legacyUserID(r)
		if bad {
			writeInvalidUserID(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

// UserIDFromContext возвращает id из контекста запроса (nil — аноним).
func UserIDFromContext(ctx context.Context) *int64 {
	id, _ := ctx.Value(ctxKey{}).(*int64)
	return id
}

// legacyUserID читает HeaderUserID. bad=true → заголовок есть, но испорчен.
func legacyUserID(r *http.Request) (id *int64, bad bool) {
	raw := r.Header.Get(HeaderUserID)
	if raw == "" {
		return nil, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return nil, true
	}
	return &v, false
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

func writeInvalidUserID(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"error":"invalid_user_id"}`))
}
