package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"home-provider/internal/models"
	"home-provider/internal/services"
)

type contextKey string

const APIKeyContextKey contextKey = "apiKey"

var apiKeyByRequest sync.Map

func APIKeyAuth(next http.Handler) http.Handler {
	keyManager := services.NewKeyManager()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var apiKey string

		// Support both Authorization: Bearer <key> and X-Api-Key: <key>
		if apiKeyHeader := r.Header.Get("X-Api-Key"); apiKeyHeader != "" {
			apiKey = apiKeyHeader
		} else if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			if apiKey == authHeader {
				slog.Warn("authentication failed: invalid authorization format",
					slog.String("path", r.URL.Path),
					slog.String("client_ip", clientIP(r)),
					slog.String("user_agent", r.UserAgent()),
					slog.String("auth_header_present", "true"),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(401)
				w.Write([]byte(`{"error":{"type":"authentication_error","message":"Invalid authorization format"}}`))
				return
			}
		} else {
			slog.Warn("authentication failed: missing authorization header",
				slog.String("path", r.URL.Path),
				slog.String("client_ip", clientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.String("auth_header_present", "false"),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			w.Write([]byte(`{"error":{"type":"authentication_error","message":"Missing authorization header"}}`))
			return
		}

		record, err := keyManager.Validate(apiKey)
		if err != nil {
			slog.Warn("authentication failed: invalid API key",
				slog.String("path", r.URL.Path),
				slog.String("client_ip", clientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.String("key_prefix", keyPrefix(apiKey)),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			w.Write([]byte(`{"error":{"type":"authentication_error","message":"Invalid API key"}}`))
			return
		}

		apiKeyByRequest.Store(r, record)

		ctx := context.WithValue(r.Context(), APIKeyContextKey, record)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

func AdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetAPIKey(r *http.Request) *models.APIKey {
	if record, ok := r.Context().Value(APIKeyContextKey).(*models.APIKey); ok {
		return record
	}
	if record, ok := apiKeyByRequest.Load(r); ok {
		return record.(*models.APIKey)
	}
	return nil
}

// clientIP extracts the client IP from the request, checking X-Forwarded-For first
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return r.RemoteAddr
}

// keyPrefix returns the first 8 characters of the API key for logging
func keyPrefix(apiKey string) string {
	if len(apiKey) < 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:8] + "..."
}
