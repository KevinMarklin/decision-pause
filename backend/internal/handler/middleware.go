package handler

import (
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// statusWriter запоминает код ответа и размер для лога запросов.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// withRequestLog — структурированный лог на каждый запрос (внешний middleware,
// чтобы видеть статус, который записали inner-миддлвари).
func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"bytes", sw.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// withRecover — panic внутри обработчика не роняет соединение молча:
// пишем 500 с телом ошибки и стеком в лог.
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"path", r.URL.Path,
					"panic", rec,
					"stack", stackTrace())
				writeError(w, http.StatusInternalServerError, "internal")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withSecurityHeaders — базовая гигиена ответов.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// withCORS — allowlist origin'ов из CORS_ORIGINS.
//
// Пустой список → заголовки CORS не ставятся (dev ходит через Vite-proxy
// и в них не нуждается). Preflight OPTIONS отвечает 204 и не доходит до auth.
func withCORS(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	anyOrigin := false
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "*" {
			anyOrigin = true
		}
		if o != "" {
			allowed[o] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (anyOrigin || allowed[origin]) {
				if anyOrigin {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+initDataHeaders())
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// initDataHeaders — заголовки, которыми мини-апп передаёт идентичность.
func initDataHeaders() string {
	return "X-Max-Init-Data, X-Max-User-Id"
}

// stackTrace — стек текущей (паникующей) горутины для лога.
func stackTrace() string {
	buf := make([]byte, 4096)
	return string(buf[:runtime.Stack(buf, false)])
}
