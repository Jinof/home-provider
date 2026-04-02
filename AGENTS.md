# AGENTS.md — Home Provider

**Generated:** 2026-03-31
**Type:** Go (API server) + Vue.js 3 (frontend)

## Overview

API proxy server that routes LLM requests (OpenAI-compatible + Anthropic) through configurable provider backends, with API key management and usage tracking.

## Project Structure

```
home-provider/
├── cmd/server/          # Go entry point
├── internal/            # Go source (handlers, services, middleware, models)
├── web/                 # Vue.js 3 frontend (built separately via Vite)
├── data/                # JSON persistence (api_keys, providers, usage, virtual_models, logs)
├── configs/             # External configs (provider API endpoints)
└── certs/               # TLS certificates
```

---

## Commands

### Go Server (Backend)

```bash
# Build
go build -o server ./cmd/server

# Run (dev)
./server

# Run with TLS + HTTP/2
TLS_CERT=./certs/cert.pem TLS_KEY=./certs/key.pem ./server

# Run a single test
go test ./internal/services/... -run TestVirtualModelManager_Create -v

# Run tests with coverage
go test ./... -cover

# Run all tests
go test ./...

# Format code
gofmt -w ./internal/

# Vet code
go vet ./...
```

### Vue Frontend (web/)

```bash
cd web

# Dev server (http://localhost:5173)
npm run dev

# Production build
npm run build

# Preview production build locally
npm run preview
```

## Routing Map

| Path                        | Handler           | Auth | Notes                   |
| --------------------------- | ----------------- | ---- | ----------------------- |
| `GET /health`               | inline (main.go)  | ❌   | Health check            |
| `GET /`                     | `web.Dashboard`   | ❌   | Serves web UI           |
| `/assets/*`                 | `http.FileServer` | ❌   | Static assets           |
| `POST /v1/chat/completions` | OpenAIHandler     | ✅   | OpenAI-compatible proxy |
| `GET /v1/models`            | OpenAIHandler     | ✅   | Model list              |
| `POST /v1/messages`         | AnthropicHandler  | ✅   | Anthropic Messages API  |
| `/admin/providers/*`        | AdminHandler      | ❌   | Provider CRUD           |
| `/admin/keys/*`             | AdminHandler      | ❌   | API key management      |
| `/admin/virtual-models/*`   | AdminHandler      | ❌   | Virtual model CRUD      |
| `GET /admin/usage`          | AdminHandler      | ❌   | Usage stats             |
| `GET /admin/logs`           | AdminHandler      | ❌   | Log buffer query        |

---

## Go Conventions

### Package Structure

- `package handlers` — HTTP request handlers
- `package services` — Business logic, singleton managers
- `package middleware` — Auth, logging middleware
- `package models` — Data structs
- `package database` — JSON file persistence

### Handler Pattern

```go
type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
    return &AdminHandler{}
}

func (h *AdminHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
    providers, err := services.NewProviderManager().List()
    if err != nil {
        respondError(w, 500, "internal_error", err.Error())
        return
    }
    respondJSON(w, 200, providers)
}
```

### Response Helpers

```go
respondError(w, statusCode, "error_type", "message")  // Always return after
respondJSON(w, statusCode, data)
```

### Service Singleton Pattern

```go
var once sync.Once
var instance *VirtualModelManager

func NewVirtualModelManager() *VirtualModelManager {
    once.Do(func() {
        instance = &VirtualModelManager{}
    })
    return instance
}
```

### Context Key Pattern

```go
type contextKey string
const APIKeyContextKey contextKey = "apiKey"
```

### Data Persistence

- Use `database.ReadJSON(path, &dest)` and `database.WriteJSON(path, data)`
- Data files: `data/api_keys.json`, `data/providers.json`, `data/virtual_models.json`, `data/usage.json`

### Logging

- Use `slog` with JSON handler
- Never log raw API keys — only log `KeyPrefix` (first 12 chars)
- Log levels controlled via `LOG_LEVEL` env (debug|info|warn|error)

### Anti-Patterns (Go)

- ❌ Do NOT call `database.ReadJSON` at package init — call explicitly per-request
- ❌ Do NOT store handler state — handlers are `struct{}`
- ❌ Do NOT forget `return` after `respondError/respondJSON`
- ❌ Do NOT use `go run` — use compiled binary
- ❌ Do NOT commit `.env` — use `.env.example` as template

---

## Vue.js 3 Conventions (web/)

### State Management

- Single-file component (`.vue`) with `<script setup lang="ts">` and `<style scoped>`
- Reactive state via `ref()` and `reactive()`
- i18n via `vue-i18n` with `$t('key')` in templates

### i18n Files

- `web/src/locales/en.json` — English
- `web/src/locales/zh.json` — Chinese
- Always add both keys when adding new text

### Router (Vue Router 4)

```typescript
import { createRouter, createWebHistory } from 'vue-router';
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/usage' },
    { path: '/usage', component: { template: '<div></div>' } },
    // ...
  ],
});
```

### Anti-Patterns (Vue)

- ❌ Do NOT use Options API — use Composition API with `<script setup>`
- ❌ Do NOT add console.log in production code
- ❌ Do NOT hardcode strings — use i18n keys
- ❌ Do NOT modify Vue files without reading them first

---

## Key Files

| File                                      | Purpose                                             |
| ----------------------------------------- | --------------------------------------------------- |
| `cmd/server/main.go`                      | Entry point, HTTP router, TLS/HTTP2, SPA fallback   |
| `internal/handlers/admin.go`              | Admin API (providers, keys, usage, logs)            |
| `internal/handlers/openai.go`             | OpenAI-compatible proxy                             |
| `internal/handlers/anthropic.go`          | Anthropic Messages API proxy                        |
| `internal/services/key_manager.go`        | API key generation, validation, storage             |
| `internal/services/provider_manager.go`   | Provider CRUD                                       |
| `internal/services/virtual_model_manager.go` | Virtual model CRUD                               |
| `internal/services/usage_tracker.go`      | Usage logging and stats                             |
| `internal/services/crypto.go`             | AES-GCM encryption (call `InitCrypto()` at startup) |
| `internal/middleware/auth.go`             | Bearer token validation, API key context            |
| `internal/middleware/logging.go`          | Request logging via slog JSON                       |
| `internal/models/*.go`                    | Data structs (APIKey, Provider, VirtualModel, UsageLog) |
| `web/src/App.vue`                         | Main Vue component with all tabs                    |
| `web/src/locales/*.json`                  | i18n translations                                   |

---

## Environment Variables

| Variable         | Default | Description                        |
| ---------------- | ------- | ---------------------------------- |
| `PORT`           | 18427   | Server port                        |
| `LOG_LEVEL`      | info    | debug\|info\|warn\|error           |
| `ENCRYPTION_KEY` | —       | 32-byte key for AES-GCM (required) |
| `DATA_DIR`       | `~/.config/home-provider` | Directory for JSON data files |

---

## Notes

- Port **18427** is the default
- Provider API keys are encrypted with AES-GCM — `services.InitCrypto()` must run at startup
- Log buffer is in-memory ring buffer (max 500 entries)
