package models

import (
	"time"
)

type APIType string

const (
	APITypeOpenAIOnly    APIType = "openai_only"
	APITypeAnthropicOnly APIType = "anthropic_only"
	APITypeBoth          APIType = "both"
)

type Provider struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	APIType           APIType   `json:"api_type"`
	OpenAIEndpoint    string    `json:"openai_endpoint"`
	AnthropicEndpoint string    `json:"anthropic_endpoint"`
	APIKeyEncrypt     string    `json:"api_key_encrypted"`
	Models            string    `json:"models"` // JSON array
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
