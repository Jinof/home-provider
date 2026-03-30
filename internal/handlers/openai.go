package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"home-provider/internal/middleware"
	"home-provider/internal/models"
	"home-provider/internal/services"
)

type OpenAIHandler struct{}

func NewOpenAIHandler() *OpenAIHandler {
	return &OpenAIHandler{}
}

type ChatCompletionRequest struct {
	Model       string               `json:"model"`
	Messages    []models.ChatMessage `json:"messages"`
	Stream      bool                 `json:"stream,omitempty"`
	Temperature float64              `json:"temperature,omitempty"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	TopP        float64              `json:"top_p,omitempty"`
}

func (h *OpenAIHandler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	apiKeyRecord := middleware.GetAPIKey(r)
	if apiKeyRecord == nil {
		respondErrorWithDetails(w, 401, "authentication_error",
			"API key is missing or invalid",
			"Provide a valid API key in the Authorization header: Bearer <your-api-key>",
			nil)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, 400, "invalid_request_error", "Failed to read request body")
		return
	}

	slog.Debug("OpenAI request received", "raw_body", string(bodyBytes))

	var req ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		slog.Error("Failed to parse OpenAI request body", "raw_body", string(bodyBytes), "error", err)
		respondErrorWithDetails(w, 400, "invalid_request_error",
			"Failed to parse request body",
			"Ensure the request body is valid JSON with required fields: model, messages",
			map[string]interface{}{"parse_error": err.Error()})
		return
	}

	resolver, err := ResolveProvider(r, req.Model, services.NewProviderManager(), services.NewTagManager())
	if err != nil {
		slog.Warn("Provider resolution failed", "model", req.Model, "error", err)
		respondErrorWithDetails(w, 400, "invalid_request_error",
			"Model/tag not found: "+req.Model,
			"Check available tags in the Tags page. Common tags: 'latest', 'default'",
			map[string]interface{}{"requested_model": req.Model})
		return
	}
	resolvedModel, err := services.NewProviderManager().ResolveModel(resolver.Provider, req.Model)
	if err != nil {
		slog.Warn("Model resolution failed", "model", req.Model, "error", err)
		respondError(w, 400, "invalid_request_error", err.Error())
		return
	}
	req.Model = resolvedModel

	apiKey, err := services.NewProviderManager().GetDecryptedKey(resolver.Provider)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to decrypt API key")
		return
	}

	start := time.Now()

	reqBody, err := json.Marshal(req)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to marshal request")
		return
	}
	upstreamURL := resolver.Provider.APIEndpoint + "/v1/chat/completions"

	req2, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(reqBody))
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to create upstream request")
		return
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+apiKey)
	SetProviderHeaders(resolver.Provider, req2)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req2)
	if err != nil {
		slog.Error("Upstream request failed", "provider", resolver.Provider.Name, "error", err)
		respondError(w, 502, "upstream_error", err.Error())
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)
		io.Copy(w, resp.Body)
		return
	}

	latency := int(time.Since(start).Milliseconds())
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	resp.Body.Close()
	if err != nil {
		slog.Error("Failed to read response body", "provider", resolver.Provider.Name, "error", err)
		respondError(w, 502, "upstream_error", "Failed to read response body")
		return
	}

	usage := ParseTokenUsage(body, true)

	services.NewUsageTracker().Log(services.UsageRecord{
		APIKeyID:     apiKeyRecord.ID,
		Provider:     resolver.Provider.Name,
		Model:        req.Model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		LatencyMs:    latency,
		StatusCode:   resp.StatusCode,
	})

	tagName := resolver.TagName()
	LogRequest(start, apiKeyRecord, r.Method, r.URL.Path, resp.StatusCode, req.Model, tagName, resolver.Provider.Name)

	if resp.StatusCode != 200 {
		providerErr := ParseProviderError(body, resolver.Provider.Name)
		suggestion := GetErrorSuggestion(resolver.Provider.Name, providerErr.ErrorType, providerErr.Reason)

		errDetails := map[string]interface{}{
			"provider":        resolver.Provider.Name,
			"provider_status": resp.StatusCode,
			"provider_error":  providerErr.ErrorType,
			"reason":          providerErr.Reason,
		}

		if tagName != "" {
			errDetails["tag"] = tagName
		}
		if req.Model != "" {
			errDetails["model"] = req.Model
		}

		slog.Error("Upstream provider error",
			slog.String("provider", resolver.Provider.Name),
			slog.Int("status", resp.StatusCode),
			slog.String("error_type", providerErr.ErrorType),
			slog.String("reason", providerErr.Reason),
		)

		respondErrorWithDetails(w, 502, "upstream_error", providerErr.Message, suggestion, errDetails)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (h *OpenAIHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	apiKeyRecord := middleware.GetAPIKey(r)
	if apiKeyRecord == nil {
		respondErrorWithDetails(w, 401, "authentication_error",
			"API key is missing or invalid",
			"Provide a valid API key in the Authorization header: Bearer <your-api-key>",
			nil)
		return
	}

	models, err := services.NewProviderManager().ListModels()
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	respondJSON(w, 200, map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}
