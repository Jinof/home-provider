package models

import (
	"encoding/json"
	"strings"
)

type ChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Name    string          `json:"name,omitempty"`
}

// GetContent returns the content as a string, handling both string and array formats.
// For array content (multi-modal), extracts all text parts and joins them.
func (m ChatMessage) GetContent() string {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(m.Content, &str); err == nil {
		return str
	}

	// Try to unmarshal as array of content parts
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(m.Content, &parts); err == nil {
		var texts []string
		for _, p := range parts {
			if p.Type == "text" {
				texts = append(texts, p.Text)
			}
		}
		return strings.Join(texts, "")
	}

	return ""
}
