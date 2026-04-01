# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build Go server
go build -o server ./cmd/server

# Run Go server
./server

# Run single test
go test ./internal/services/... -run TestTagManager_Create -v

# Format Go code
gofmt -w ./internal/

# Build Vue frontend
cd web && npm install && npm run build
```

## Architecture

API proxy server routing LLM requests through configurable provider backends. Request flow:

```
Client → API Key Validation → Tag Resolution → Provider → Response
```

**Routing:**
- `/v1/chat/completions` → OpenAIHandler (OpenAI-compatible)
- `/v1/messages` → AnthropicHandler (Anthropic Messages API)
- `/admin/*` → AdminHandler (management UI)
- `/` → Vue.js web UI

**Middleware chain for API routes:** RequestLogger → APIKeyAuth → Handler

**Key packages:**
- `internal/handlers/` — HTTP handlers (admin, openai, anthropic)
- `internal/services/` — Business logic (KeyManager, ProviderManager, TagManager, UsageTracker)
- `internal/middleware/` — Auth and logging middleware
- `internal/models/` — Data structures

## Patterns

**Service singletons:**
```go
var once sync.Once
var instance *TagManager
func NewTagManager() *TagManager {
    once.Do(func() { instance = &TagManager{} })
    return instance
}
```

**Response helpers:**
```go
respondError(w, statusCode, "error_type", "message")  // Always return after
respondJSON(w, statusCode, data)
```

**Context for API key:**
```go
apiKey := middleware.GetAPIKey(r)  // Returns *models.APIKey or nil
```

## Important Notes

- Port 18427 is the default
- API keys encrypted with AES-GCM — call `services.InitCrypto()` at startup
- Logs written to `~/Library/Logs/home-provider/app-YYYY-MM-DD.log`
- Authentication supports both `Authorization: Bearer <key>` and `X-Api-Key: <key>` headers
- SSE stream errors (HTTP 200 with error in body) are detected and mapped to proper HTTP status codes

## Anti-Patterns

- ❌ Do NOT call `database.ReadJSON` at package init — call explicitly per-request
- ❌ Do NOT store handler state — handlers are `struct{}`
- ❌ Do NOT forget `return` after `respondError/respondJSON`
- ❌ Do NOT commit `.env` — use `.env.example` as template
- ❌ Do NOT use `go run` — use compiled binary

## 环境变量

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `PORT` | 18427 | Server port |
| `LOG_LEVEL` | info | debug\|info\|warn\|error |
| `ENCRYPTION_KEY` | — | 32-byte key for AES-GCM (required) |
| `DATA_DIR` | `~/.config/home-provider` | JSON data files directory |
