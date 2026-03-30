package middleware

import (
	"context"
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
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			http.Error(w, `{"error":{"type":"authentication_error","message":"Missing authorization header"}}`, 401)
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == authHeader {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			http.Error(w, `{"error":{"type":"authentication_error","message":"Invalid authorization format"}}`, 401)
			return
		}

		record, err := keyManager.Validate(apiKey)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			http.Error(w, `{"error":{"type":"authentication_error","message":"Invalid API key"}}`, 401)
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
