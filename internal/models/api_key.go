package models

import (
	"time"
)

type APIKey struct {
	ID        string `json:"id"`
	KeyHash   string `json:"key_hash"`
	KeyPrefix string `json:"key_prefix"`
	Name      string `json:"name"`

	RequestLimit *int       `json:"request_limit"`
	ExpiresAt    *time.Time `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
}
