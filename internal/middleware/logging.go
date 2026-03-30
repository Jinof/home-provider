package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func LogLevel(status int) string {
	if status >= 500 {
		return "ERROR"
	}
	if status >= 400 {
		return "WARN"
	}
	return "INFO"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logs" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		status := rec.status

		attrs := []any{
			slog.String("type", "admin"),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", duration),
		}

		apiKey := GetAPIKey(r)
		if apiKey != nil {
			attrs = append(attrs, slog.String("key_prefix", apiKey.KeyPrefix))
		}

		level := LogLevel(status)
		if level == "ERROR" {
			slog.Error("request failed", attrs...)
		} else if level == "WARN" {
			slog.Warn("request error", attrs...)
		} else {
			slog.Info("request", attrs...)
		}

		apiKeyByRequest.Delete(r)
	})
}
