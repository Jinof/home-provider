# internal/services

## OVERVIEW

Service layer for home-provider: manages API keys, provider configs, usage tracking, and encryption.

## STRUCTURE

```
internal/services/
├── key_manager.go       # API key generation, validation, revocation
├── provider_manager.go  # Provider configuration CRUD
├── usage_tracker.go     # API usage metrics collection
├── crypto.go            # AES-GCM encryption/decryption
```

## WHERE TO LOOK

**API key operations**: `key_manager.go`

- Generate `hpk_` prefixed keys, SHA-256 hash storage
- Validate, revoke, list keys
- Persists to `data/api_keys.json`

**Provider configs**: `provider_manager.go`

- Create, read, update, delete provider configurations
- Persists to `data/providers.json`

**Usage metrics**: `usage_tracker.go`

- Track input/output tokens, latency, status codes per request

**Encryption**: `crypto.go`

- `Encrypt()` / `Decrypt()` via AES-GCM
- Call `InitCrypto()` during app startup before use

## CONVENTIONS

**Singleton pattern**: Every service uses a package-level singleton.

```go
var keyManager = &KeyManager{}
func NewKeyManager() *KeyManager { return keyManager }
```

Always use `NewXxxManager()` to get the instance. Never instantiate directly.

**JSON persistence**: All services use `database.ReadJSON` / `database.WriteJSON`.
Data files live in `data/`:

- `data/api_keys.json`
- `data/providers.json`
- `data/usage.json`

**Key prefix**: API keys are prefixed with `hpk_` (e.g., `hpk_abc123...`).

**Crypto initialization**: `crypto.go` requires explicit `InitCrypto()` call at startup. Do not call `Encrypt()`/`Decrypt()` before initialization.

## ANTI-PATTERNS

**Do not call crypto before InitCrypto()**: Will panic. Ensure initialization happens in `main.go` or equivalent entry point before any request handling.

**Do not bypass NewXxxManager()**: Direct struct instantiation skips the singleton and can cause inconsistent state.

**Do not hardcode `data/` paths**: Use the `database` package functions which handle path resolution.
