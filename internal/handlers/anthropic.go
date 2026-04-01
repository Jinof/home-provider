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

type AnthropicHandler struct{}

func NewAnthropicHandler() *AnthropicHandler {
	return &AnthropicHandler{}
}

type AnthropicMessageRequest struct {
	Model     string               `json:"model"`
	System    interface{}          `json:"system,omitempty"`
	Messages  []models.ChatMessage `json:"messages"`
	MaxTokens int                  `json:"max_tokens"`
	Stream    bool                 `json:"stream,omitempty"`
}

func (h *AnthropicHandler) Messages(w http.ResponseWriter, r *http.Request) {
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

	slog.Debug("Anthropic request received", "raw_body", string(bodyBytes))

	var req AnthropicMessageRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		slog.Error("Failed to parse Anthropic request body", "raw_body", string(bodyBytes), "error", err)
		respondErrorWithDetails(w, 400, "invalid_request_error",
			"Failed to parse request body",
			"Ensure the request body is valid JSON with required fields: model, messages, max_tokens",
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

	var upstreamURL string
	var needTransform bool
	var upstreamUsesOpenAIFormat bool

	switch resolver.Provider.APIType {
	case models.APITypeAnthropicOnly:
		upstreamURL = resolver.Provider.AnthropicEndpoint
	case models.APITypeOpenAIOnly:
		upstreamURL = resolver.Provider.OpenAIEndpoint
		upstreamUsesOpenAIFormat = true
		needTransform = true
	case models.APITypeBoth:
		upstreamURL = resolver.Provider.AnthropicEndpoint
	default:
		upstreamURL = resolver.Provider.AnthropicEndpoint
	}

	upstreamReq := h.buildUpstreamRequest(req, upstreamUsesOpenAIFormat)
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
	SetUpstreamAuthHeaders(req2, resolver.Provider, apiKey, upstreamUsesOpenAIFormat)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req2)
	if err != nil {
		slog.Error("Upstream request failed", "provider", resolver.Provider.Name, "error", err)
		respondError(w, 502, "upstream_error", err.Error())
		return
	}

	if req.Stream {
		// Read first chunk to check for error in SSE stream
		firstChunk, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		resp.Body.Close()
		if err != nil {
			slog.Error("Failed to read stream response", "provider", resolver.Provider.Name, "error", err)
			respondError(w, 502, "upstream_error", "Failed to read stream response")
			return
		}

		// Check if the stream contains an error in first chunk
		slog.Debug("Upstream stream first chunk",
			slog.String("provider", resolver.Provider.Name),
			slog.Int("status", resp.StatusCode),
			slog.String("content_type", resp.Header.Get("Content-Type")),
			slog.String("body", string(firstChunk)),
		)

		// First check HTTP status code - if not 200, treat as error regardless of body content
		if resp.StatusCode != 200 {
			slog.Warn("Upstream stream returned non-200 status",
				slog.String("provider", resolver.Provider.Name),
				slog.Int("status", resp.StatusCode),
				slog.String("body", string(firstChunk)),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			w.Write(firstChunk)
			return
		}

		// Check if the body contains an error even with 200 status
		if errBody := parseErrorBody(firstChunk); errBody != nil {
			slog.Warn("Upstream stream returned error",
				slog.String("provider", resolver.Provider.Name),
				slog.String("error_type", errBody.Type),
				slog.String("message", errBody.Message),
			)
			httpStatus := errorTypeToHTTPStatus(errBody.Type)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpStatus)
			w.Write(firstChunk)
			return
		}

		// Not an error, send as SSE with 200
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)
		w.Write(firstChunk)
		io.Copy(w, resp.Body)
		return
	}

	latency := int(time.Since(start).Milliseconds())
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	resp.Body.Close()
	if err != nil {
		slog.Error("Failed to read response body", "provider", resolver.Provider.Name, "error", err)
		respondError(w, 502, "upstream_error", "Failed to read response body")
		return
	}

	// Check if response body contains an error (even if HTTP status is 200)
	if errBody := parseErrorBody(body); errBody != nil {
		slog.Warn("Upstream returned error in body",
			slog.String("provider", resolver.Provider.Name),
			slog.Int("status", resp.StatusCode),
			slog.String("content_type", resp.Header.Get("Content-Type")),
			slog.String("body", string(body)),
			slog.String("error_type", errBody.Type),
			slog.String("message", errBody.Message),
		)
		httpStatus := errorTypeToHTTPStatus(errBody.Type)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		w.Write(body)
		return
	}

	slog.Debug("Upstream raw response",
		slog.String("provider", resolver.Provider.Name),
		slog.Int("status", resp.StatusCode),
		slog.String("content_type", resp.Header.Get("Content-Type")),
		slog.String("body", string(body)),
	)

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

	// Apply transformation if needed (OpenAI_Only provider's Anthropic endpoint returns OpenAI format)
	if needTransform {
		transformedResp := TransformOpenAIResponseToAnthropic(body, req.Model)
		if transformedResp != nil {
			transformedBody, err := json.Marshal(transformedResp)
			if err == nil {
				body = transformedBody
			}
		}
	}

	// Pass through upstream response directly without transformation
	// Copy upstream headers to downstream response
	for k, v := range resp.Header {
		if k == "Content-Length" {
			continue // Don't copy Content-Length as we're changing the body
		}
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (h *AnthropicHandler) buildUpstreamRequest(req AnthropicMessageRequest, upstreamUsesOpenAIFormat bool) map[string]interface{} {
	messages := make([]map[string]interface{}, 0)

	if req.System != nil {
		switch sys := req.System.(type) {
		case string:
			if sys != "" {
				messages = append(messages, map[string]interface{}{"role": "system", "content": sys})
			}
		case []interface{}:
			for _, block := range sys {
				if str, ok := block.(string); ok && str != "" {
					messages = append(messages, map[string]interface{}{"role": "system", "content": str})
				}
			}
		default:
			if str, ok := req.System.(string); ok && str != "" {
				messages = append(messages, map[string]interface{}{"role": "system", "content": str})
			}
		}
	}

	for _, m := range req.Messages {
		if upstreamUsesOpenAIFormat {
			messages = append(messages, map[string]interface{}{"role": m.Role, "content": m.GetContent()})
		} else {
			messages = append(messages, map[string]interface{}{"role": m.Role, "content": m.Content})
		}
	}

	result := map[string]interface{}{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": req.MaxTokens,
	}
	if !upstreamUsesOpenAIFormat {
		result["stream"] = req.Stream
	}
	return result
}

type upstreamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error upstreamError `json:"error"`
}

func parseErrorBody(body []byte) *upstreamError {
	// Try JSON format: {"error": {...}}
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Error.Type != "" || errResp.Error.Message != "" {
			return &errResp.Error
		}
	}

	// Try SSE format: data: {"error": {...}}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			sseData := strings.TrimPrefix(line, "data: ")
			errResp = errorResponse{}
			if unmarshalErr := json.Unmarshal([]byte(sseData), &errResp); unmarshalErr == nil {
				if errResp.Error.Type != "" || errResp.Error.Message != "" {
					return &errResp.Error
				}
			}
		}
	}
	return nil
}

func errorTypeToHTTPStatus(errorType string) int {
	switch errorType {
	case "rate_limit_reached_error", "rate_limit_error":
		return 429
	case "authentication_error", "invalid_api_key_error":
		return 401
	case "permission_error", "forbidden_error":
		return 403
	case "not_found_error":
		return 404
	case "invalid_request_error", "bad_request_error":
		return 400
	case "overloaded_error", "service_unavailable_error":
		return 503
	default:
		return 500
	}
}
