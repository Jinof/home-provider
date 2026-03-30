# OVERVIEW

Middleware for request authentication and request logging via slog.

## STRUCTURE

- `auth.go` (69 lines): Bearer token validation, API key context propagation, admin auth stub.
- `logging.go` (66 lines): Request logging via slog JSON handler.

## WHERE TO LOOK

- **API key validation**: Apply `APIKeyAuth` to routes requiring Bearer token auth.
- **Admin endpoints**: `AdminAuth` stub exists but passes through all requests unchanged.
- **Request logging**: `RequestLogger` logs admin and non-inference requests via slog.
- **API key from context**: Call `GetAPIKey(r)` to retrieve validated `*models.APIKey` from request context.

## LOGGING

**Log Output**: All logs are written to slog in JSON format.

**Log Directory**: `~/Library/Logs/home-provider/` (macOS) or `$LOG_DIR` environment variable.

**Log File**: `app-YYYY-MM-DD.log` with daily rotation.

**Log Format** (slog JSON):

```json
{"time":"...","level":"INFO","msg":"request","type":"inference","method":"POST","path":"/messages","status":200,"latency":...,"key_prefix":"hpk_xxx","model":"kimi-k2.5","tag":"latest","provider":"Kimi"}
```

**Log Fields**:
| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `inference` or `admin` |
| `method` | string | HTTP method |
| `path` | string | Request path |
| `status` | int | HTTP status code |
| `latency` | duration | Request latency |
| `key_prefix` | string | API key prefix (not full key) |
| `model` | string | Model name (inference only) |
| `tag` | string | Tag name (inference only) |
| `provider` | string | Provider name (inference only) |

## CONVENTIONS

**Context key pattern:**

```go
type contextKey string
const APIKeyContextKey contextKey = "apiKey"
```

**401 JSON error format:**

```json
{ "error": { "type": "authentication_error", "message": "..." } }
```

**Auth failure**: Set `Content-Type: application/json`, write header 401, use `http.Error` with JSON body.

## ANTI-PATTERNS

- Do not use `context.WithValue` with a plain string key; define a `contextKey` type.
- Do not hardcode 401 response bodies; the JSON structure is part of the API contract.
- Do not store large objects in log entries; logs are written to disk.
