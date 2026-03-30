package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

func parseBodyAsJSON(data []byte) any {
	var result any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}
	return string(data)
}

func HttpLogger(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Any("body", parseBodyAsJSON(bodyBytes)),
		)

		rec := &responseRecorder{ResponseWriter: w, status: 200}

		next.ServeHTTP(rec, r)

		if rec.status >= 400 {
			logger.Warn("http response error",
				slog.Int("status", rec.status),
				slog.Any("body", parseBodyAsJSON(rec.body.Bytes())),
			)
		}
	})
}
