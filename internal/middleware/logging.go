package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
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
	status    int
	body      bytes.Buffer
	headerMap http.Header
	written   bool
}

func (rec *statusRecorder) Header() http.Header {
	if rec.headerMap == nil {
		rec.headerMap = make(http.Header)
	}
	return rec.headerMap
}

func (rec *statusRecorder) WriteHeader(code int) {
	if !rec.written {
		for k, v := range rec.headerMap {
			rec.ResponseWriter.Header()[k] = v
		}
		rec.written = true
	}
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.written {
		for k, v := range rec.headerMap {
			rec.ResponseWriter.Header()[k] = v
		}
		rec.written = true
	}
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

func parseBodyAsJSON(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	var result any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}
	return string(data)
}

func sanitizeHeaders(h http.Header) map[string][]string {
	sensitive := []string{"authorization", "x-api-key", "cookie", "set-cookie", "x-token"}
	result := make(map[string][]string)
	for k, v := range h {
		if slices.Contains(sensitive, strings.ToLower(k)) {
			result[k] = []string{"***REDACTED***"}
		} else {
			result[k] = v
		}
	}
	return result
}

func parseHeaders(h http.Header) map[string][]string {
	result := make(map[string][]string)
	for k, v := range h {
		result[k] = v
	}
	return result
}

func httpRequestToLog(r *http.Request, body []byte) map[string]any {
	return map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"raw_query":   r.URL.RawQuery,
		"query":       r.URL.Query(),
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"headers":     sanitizeHeaders(r.Header),
		"body":        parseBodyAsJSON(body),
	}
}

func httpResponseToLog(rec *statusRecorder) map[string]any {
	return map[string]any{
		"status":  rec.status,
		"headers": parseHeaders(rec.headerMap),
		"body":    parseBodyAsJSON(rec.body.Bytes()),
	}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logs" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		rec := &statusRecorder{ResponseWriter: w, status: 200, headerMap: make(http.Header)}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		attrs := []any{
			slog.String("type", "admin"),
			slog.Any("http_request", httpRequestToLog(r, bodyBytes)),
			slog.Any("http_response", httpResponseToLog(rec)),
			slog.Duration("latency", duration),
		}

		apiKey := GetAPIKey(r)
		if apiKey != nil {
			attrs = append(attrs, slog.String("key_prefix", apiKey.KeyPrefix))
		}

		level := LogLevel(rec.status)
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
