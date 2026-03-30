# Home Provider

API proxy server that routes LLM requests (OpenAI-compatible + Anthropic) through configurable provider backends, with API key management and usage tracking.

## Features

- **OpenAI-compatible API** — `POST /v1/chat/completions`
- **Anthropic Messages API** — `POST /v1/messages`
- **Tag-based Model Routing** — Use unified tag names (e.g., `latest`) across machines and agents
- **API Key Management** — Generate, manage, and track API keys with AES-GCM encryption
- **Usage Tracking** — Per-key and global statistics with time series data
- **Multi-language Support** — English and Chinese (i18n)

## Quick Start

### 1. Build the Server

```bash
go build -o server ./cmd/server
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and set `ENCRYPTION_KEY` (32-byte key for AES-GCM encryption).

### 3. Run

```bash
./server
```

### 4. Access Web UI

Open http://localhost:18427

## Usage

### 1. Add a Provider

Go to **Providers** tab → Add Provider:

- **Name**: e.g., `MiniMax`
- **Endpoint**: e.g., `https://api.minimax.io`
- **API Key**: Your provider's API key
- **Models**: e.g., `["MiniMax-M2.7-highspeed"]`

### 2. Create a Tag

Go to **Tags** tab → Create Tag:

- **Name**: e.g., `latest`
- **Provider**: Select the provider you added

### 3. Generate an API Key

Go to **API Keys** tab → Generate Key

### 4. Make Requests

```bash
# Anthropic Messages API
curl -X POST http://127.0.0.1:18427/v1/messages \
  -H "Authorization: Bearer hpk_xxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "latest",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 1024
  }'

# OpenAI-compatible API
curl -X POST http://127.0.0.1:18427/v1/chat/completions \
  -H "Authorization: Bearer hpk_xxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "latest",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## How It Works

```
Client → API Key → Tag → Provider → AI Response
```

1. Client sends request with API key and tag name (e.g., `latest`)
2. API key is validated
3. Tag is resolved to a specific provider and model
4. Request is forwarded to the provider
5. Response is returned to client

## Environment Variables

| Variable         | Default | Description                        |
| ---------------- | ------- | ---------------------------------- |
| `PORT`           | 18427   | Server port                        |
| `ENCRYPTION_KEY` | —       | 32-byte key for AES-GCM (required) |
| `LOG_LEVEL`      | info    | debug\|info\|warn\|error           |
| `DATA_DIR`       | ./data  | Data directory                     |

## Development

```bash
# Build frontend
cd web && npm install && npm run build

# Run tests
go test ./...

# Format code
gofmt -w ./internal/
```
