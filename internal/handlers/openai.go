package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
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

	upstreamReq := any(req)
	upstreamURL := resolver.Provider.OpenAIEndpoint
	var needTransform bool

	switch resolver.Provider.APIType {
	case models.APITypeAnthropicOnly:
		upstreamReq = buildAnthropicRequestFromOpenAI(req)
		upstreamURL = resolver.Provider.AnthropicEndpoint
		needTransform = true
	case models.APITypeOpenAIOnly, models.APITypeBoth:
		// OpenAI_Only and Both providers return OpenAI format directly.
	default:
		// Default behavior: no transformation.
	}

	reqBody, err := json.Marshal(upstreamReq)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to marshal request")
		return
	}

	slog.Debug("Upstream request",
		slog.String("provider", resolver.Provider.Name),
		slog.String("api_type", string(resolver.Provider.APIType)),
		slog.String("upstream_url", upstreamURL),
		slog.String("model", req.Model),
	)

	req2, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(reqBody))
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to create upstream request")
		return
	}
	req2.Header.Set("Content-Type", "application/json")
	SetUpstreamAuthHeaders(req2, resolver.Provider, apiKey, resolver.Provider.APIType != models.APITypeAnthropicOnly)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req2)
	if err != nil {
		slog.Error("Upstream request failed", "provider", resolver.Provider.Name, "error", err)
		respondError(w, 502, "upstream_error", err.Error())
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/event-stream"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(resp.StatusCode)
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

	usage := ParseTokenUsage(body)

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

	// Apply transformation if needed (Anthropic_Only provider's OpenAI endpoint returns Anthropic format)
	if needTransform {
		transformedResp := TransformAnthropicResponseToOpenAI(body)
		if transformedResp != nil {
			transformedBody, err := json.Marshal(transformedResp)
			if err == nil {
				body = transformedBody
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func buildAnthropicRequestFromOpenAI(req ChatCompletionRequest) map[string]interface{} {
	const defaultAnthropicMaxTokens = 1024

	messages := make([]map[string]interface{}, 0, len(req.Messages))
	systemParts := make([]string, 0)
	for _, message := range req.Messages {
		content := message.GetContent()
		if message.Role == "system" {
			if content != "" {
				systemParts = append(systemParts, content)
			}
			continue
		}

		messages = append(messages, map[string]interface{}{
			"role":    message.Role,
			"content": content,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultAnthropicMaxTokens
	}

	upstreamReq := map[string]interface{}{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": maxTokens,
		"stream":     req.Stream,
	}
	if len(systemParts) > 0 {
		upstreamReq["system"] = strings.Join(systemParts, "\n\n")
	}
	if req.Temperature != 0 {
		upstreamReq["temperature"] = req.Temperature
	}
	if req.TopP != 0 {
		upstreamReq["top_p"] = req.TopP
	}

	return upstreamReq
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
