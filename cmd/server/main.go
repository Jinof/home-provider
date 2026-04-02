package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"home-provider/internal/handlers"
	"home-provider/internal/middleware"
	"home-provider/internal/services"
	"home-provider/internal/web"

	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	godotenv.Load()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		logDir = filepath.Join(homeDir, "Library", "Logs", "home-provider")
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	logFilePath := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	slog.SetDefault(slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: level})))

	if err := services.InitCrypto(); err != nil {
		slog.Error("Failed to init crypto", "error", err)
		os.Exit(1)
	}

	if err := services.NewVirtualModelManager().EnsureDefaultVirtualModel(""); err != nil {
		slog.Warn("Failed to ensure default virtual model", "error", err)
	}

	anthropicHandler := handlers.NewAnthropicHandler()
	openaiHandler := handlers.NewOpenAIHandler()
	adminHandler := handlers.NewAdminHandler()

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /messages", anthropicHandler.Messages)
	apiMux.HandleFunc("POST /chat/completions", openaiHandler.ChatCompletions)
	apiMux.HandleFunc("GET /models", openaiHandler.ListModels)
	apiMuxWithAuth := middleware.RequestLogger(middleware.APIKeyAuth(apiMux))

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /providers", adminHandler.ListProviders)
	adminMux.HandleFunc("POST /providers", adminHandler.CreateProvider)
	adminMux.HandleFunc("PUT /providers/{id}", adminHandler.UpdateProvider)
	adminMux.HandleFunc("DELETE /providers/{id}", adminHandler.DeleteProvider)
	adminMux.HandleFunc("GET /keys", adminHandler.ListKeys)
	adminMux.HandleFunc("POST /keys", adminHandler.CreateKey)
	adminMux.HandleFunc("DELETE /keys/{id}", adminHandler.DeleteKey)
	adminMux.HandleFunc("POST /virtual-models", adminHandler.CreateVirtualModel)
	adminMux.HandleFunc("GET /virtual-models", adminHandler.ListVirtualModels)
	adminMux.HandleFunc("GET /virtual-models/{id}", adminHandler.GetVirtualModel)
	adminMux.HandleFunc("PUT /virtual-models/{id}", adminHandler.UpdateVirtualModel)
	adminMux.HandleFunc("DELETE /virtual-models/{id}", adminHandler.DeleteVirtualModel)
	adminMux.HandleFunc("GET /usage", adminHandler.GetUsage)
	adminMux.HandleFunc("GET /logs", adminHandler.GetLogs)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/health" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		if path == "/" && r.Method == "GET" {
			web.Dashboard(w, r)
			return
		}

		if strings.HasPrefix(path, "/assets/") {
			http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/dist/assets"))).ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(path, "/v1") {
			r.URL.Path = strings.TrimPrefix(path, "/v1")
			apiMuxWithAuth.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(path, "/admin") {
			r.URL.Path = strings.TrimPrefix(path, "/admin")
			middleware.RequestLogger(adminMux).ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, "./web/dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "18427"
	}

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: handler,
	}

	if certFile != "" && keyFile != "" {
		slog.Info("Starting home-provider with TLS/HTTP2", "port", port)
		http2.ConfigureServer(srv, &http2.Server{})
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			slog.Error("Failed to start server with TLS", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Starting home-provider with h2c (HTTP/2 over plain HTTP)", "port", port)
		http2.ConfigureServer(srv, &http2.Server{})
		h2s := &http2.Server{}
		srv2 := &http.Server{
			Addr:    "0.0.0.0:" + port,
			Handler: h2c.NewHandler(handler, h2s),
		}
		if err := srv2.ListenAndServe(); err != nil {
			slog.Error("Failed to start server with h2c", "error", err)
			os.Exit(1)
		}
	}
}
