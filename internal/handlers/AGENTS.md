# Handlers

## OVERVIEW

HTTP request handlers for the home-provider API, proxying OpenAI and Anthropic endpoints with admin key/provider management.

## STRUCTURE

- `admin.go` — AdminHandler: CRUD for providers and keys, usage stats, log viewing
- `anthropic.go` — AnthropicHandler: POST /messages (Anthropic Messages API proxy)
- `openai.go` — OpenAIHandler: POST /chat/completions, GET /models (OpenAI-compatible API proxy)
- `response.go` — `respondJSON`, `respondError` response helpers
- `shared.go` — `resolveProvider` helper for provider resolution by model or header

## WHERE TO LOOK

| Route                           | Handler          | Method                |
| ------------------------------- | ---------------- | --------------------- |
| `GET /admin/providers`          | AdminHandler     | ListProviders         |
| `POST /admin/providers`         | AdminHandler     | CreateProvider        |
| `GET /admin/providers/:id`      | AdminHandler     | (via ProviderManager) |
| `PUT /admin/providers/:id`      | AdminHandler     | UpdateProvider        |
| `DELETE /admin/providers/:id`   | AdminHandler     | DeleteProvider        |
| `GET /admin/keys`               | AdminHandler     | ListKeys              |
| `POST /admin/keys`              | AdminHandler     | CreateKey             |
| `DELETE /admin/keys/:id`        | AdminHandler     | DeleteKey             |
| `POST /admin/keys/:id/provider` | AdminHandler     | SetActiveProvider     |
| `GET /admin/usage`              | AdminHandler     | GetUsage              |
| `GET /admin/logs`               | AdminHandler     | GetLogs               |
| `POST /messages`                | AnthropicHandler | Messages              |
| `POST /chat/completions`        | OpenAIHandler    | ChatCompletions       |
| `GET /models`                   | OpenAIHandler    | Models                |

## CONVENTIONS

- **Response helpers**: `respondError(w, code, "type", "message")` for errors, `respondJSON(w, code, data)` for success
- **Manager instantiation**: `services.NewXManager()` — create fresh instance per request, do not store in handler struct
- **Handler struct pattern**: `type XHandler struct{}` with `func NewXHandler() *XHandler` constructor returning `&XHandler{}`
- **Handler method signature**: `func (h *XHandler) Method(w http.ResponseWriter, r *http.Request)`
- **API key access**: `middleware.GetAPIKey(r)` returns `*models.APIKey` or nil
- **Request body decoding**: `json.NewDecoder(r.Body).Decode(&var)` with validation immediately after
- **Path parameters**: `r.PathValue("id")` (Go 1.22+)

## ANTI-PATTERNS

- Do not store `*services.XManager` as a field on the handler struct; always call `services.NewXManager()` inline
- Do not call `respondError` without `return` — error responses must halt handler execution
- Do not pass raw user input (e.g., `r.Body`, path values) to service methods without validation
- Do not use `fmt.Print` or `log.Print`; use `slog` with the configured level
- Do not create one-off goroutines within handlers without proper context cancellation handling
