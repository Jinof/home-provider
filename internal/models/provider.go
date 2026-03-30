package models

import (
	"time"
)

type Provider struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	APIEndpoint   string    `json:"api_endpoint"`
	APIKeyEncrypt string    `json:"api_key_encrypted"`
	Models        string    `json:"models"` // JSON array
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
