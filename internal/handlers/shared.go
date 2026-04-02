package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"home-provider/internal/middleware"
	"home-provider/internal/models"
	"home-provider/internal/services"
)

type ProviderResolver struct {
	Provider     *models.Provider
	VirtualModel *models.VirtualModel
}

func (r *ProviderResolver) VirtualModelName() string {
	if r.VirtualModel != nil {
		return r.VirtualModel.Name
	}
	return ""
}

const (
	ProviderKimi    = "Kimi"
	ProviderMiniMax = "MiniMax"
)

func ResolveProvider(r *http.Request, model string, pm *services.ProviderManager, vm *services.VirtualModelManager) (*ProviderResolver, error) {
	virtualModel, err := vm.GetByName(model)
	if err != nil || virtualModel == nil {
		return nil, errors.New("virtual model not found")
	}
	provider, err := pm.Get(virtualModel.ProviderID)
	if err != nil || provider == nil {
		return nil, errors.New("provider not found for virtual model")
	}
	return &ProviderResolver{Provider: provider, VirtualModel: virtualModel}, nil
}

func UsesBearerAuthForAnthropicEndpoint(provider *models.Provider) bool {
	return provider.Name == ProviderKimi || provider.Name == ProviderMiniMax
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func SetProviderHeaders(provider *models.Provider, req *http.Request) {
	if provider.Name == ProviderKimi {
		req.Header.Set("User-Agent", "KimiCLI/1.3")
	}
}

func SetUpstreamAuthHeaders(req *http.Request, provider *models.Provider, apiKey string, useOpenAIAuth bool) {
	if useOpenAIAuth {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		SetProviderHeaders(provider, req)
		return
	}

	if UsesBearerAuthForAnthropicEndpoint(provider) {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		SetProviderHeaders(provider, req)
	} else {
		req.Header.Set("x-api-key", apiKey)
	}
	req.Header.Set("anthropic-version", "2023-06-01")
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

func ParseTokenUsage(body []byte) TokenUsage {
	var usage TokenUsage
	var anthropicParsed struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if json.Unmarshal(body, &anthropicParsed) == nil &&
		(anthropicParsed.Usage.InputTokens != 0 || anthropicParsed.Usage.OutputTokens != 0) {
		usage.InputTokens = anthropicParsed.Usage.InputTokens
		usage.OutputTokens = anthropicParsed.Usage.OutputTokens
		return usage
	}

	var openAIParsed struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if json.Unmarshal(body, &openAIParsed) == nil {
		usage.InputTokens = openAIParsed.Usage.PromptTokens
		usage.OutputTokens = openAIParsed.Usage.CompletionTokens
	}

	return usage
}

func TransformOpenAIResponseToAnthropic(openAIBody []byte, model string) map[string]interface{} {
	var openAIResp openAIResponse

	if err := json.Unmarshal(openAIBody, &openAIResp); err != nil {
		slog.Warn("Failed to parse OpenAI response for transformation", "error", err)
		return nil
	}

	if len(openAIResp.Choices) == 0 {
		return nil
	}

	content := openAIResp.Choices[0].Message.Content
	stopReason := "end_turn"
	if openAIResp.Choices[0].FinishReason == "length" {
		stopReason = "max_tokens"
	}

	return map[string]interface{}{
		"id":            openAIResp.ID,
		"type":          "message",
		"role":          "assistant",
		"content":       []map[string]interface{}{{"type": "text", "text": content}},
		"model":         model,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  openAIResp.Usage.PromptTokens,
			"output_tokens": openAIResp.Usage.CompletionTokens,
		},
	}
}

func TransformAnthropicResponseToOpenAI(anthropicBody []byte) map[string]interface{} {
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(anthropicBody, &anthropicResp); err != nil {
		slog.Warn("Failed to parse Anthropic response for transformation", "error", err)
		return nil
	}

	content := ""
	if len(anthropicResp.Content) > 0 && anthropicResp.Content[0].Type == "text" {
		content = anthropicResp.Content[0].Text
	}

	finishReason := "stop"
	if anthropicResp.StopReason == "max_tokens" {
		finishReason = "length"
	}

	return map[string]interface{}{
		"id":      anthropicResp.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   anthropicResp.Model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": finishReason,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     anthropicResp.Usage.InputTokens,
			"completion_tokens": anthropicResp.Usage.OutputTokens,
			"total_tokens":      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}
}

type ProviderErrorInfo struct {
	ErrorType string
	Message   string
	Reason    string
}

func ParseProviderError(body []byte, providerName string) ProviderErrorInfo {
	var providerErr struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if json.Unmarshal(body, &providerErr) == nil && providerErr.Error.Type != "" {
		return ProviderErrorInfo{
			ErrorType: providerErr.Error.Type,
			Message:   providerErr.Error.Message,
			Reason:    providerErr.Error.Message,
		}
	}

	return ProviderErrorInfo{
		ErrorType: "upstream_error",
		Message:   "Provider returned an error",
		Reason:    string(body),
	}
}

func GetErrorSuggestion(providerName, errorType, reason string) string {
	switch errorType {
	case "access_terminated_error":
		return "This provider requires requests from recognized coding agents. Check if your User-Agent is properly set."
	case "invalid_request_error":
		return "Check your request format and parameters. Ensure the model name is valid for this provider."
	case "authentication_error":
		return "Verify your API key is correct and has not expired."
	case "rate_limit_error":
		return "Rate limit exceeded. Wait and retry, or contact the provider to increase your quota."
	case "model_not_found_error", "not_found_error":
		return "The model may not be available on this provider. Check available models or try a different virtual model."
	default:
		if providerName == "Kimi" && errorType == "upstream_error" {
			return "Kimi requires specific User-Agent. Ensure you're using KimiCLI/1.3 or similar."
		}
		return "Check provider status and try again. If persists, contact provider support."
	}
}

func LogRequest(start time.Time, apiKeyRecord *models.APIKey, method, path string, status int, model, virtualModelName, provider string) {
	attrs := []any{
		slog.String("type", "inference"),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("latency", time.Since(start)),
		slog.String("key_prefix", apiKeyRecord.KeyPrefix),
		slog.String("model", model),
		slog.String("virtual_model", virtualModelName),
		slog.String("provider", provider),
	}

	level := middleware.LogLevel(status)
	if level == "ERROR" {
		slog.Error("request failed", attrs...)
	} else if level == "WARN" {
		slog.Warn("request error", attrs...)
	} else {
		slog.Info("request", attrs...)
	}
}
