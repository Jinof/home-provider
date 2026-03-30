package models

import (
	"time"
)

type UsageLog struct {
	ID           string    `json:"id"`
	APIKeyID     string    `json:"api_key_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	LatencyMs    int       `json:"latency_ms"`
	StatusCode   int       `json:"status_code"`
	CreatedAt    time.Time `json:"created_at"`
}
