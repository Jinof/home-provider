package models

import "time"

type VirtualModel struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ProviderID string    `json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
