package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"home-provider/internal/models"
	"home-provider/internal/services"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := services.NewProviderManager().List()
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}
	respondJSON(w, 200, providers)
}

func (h *AdminHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name              string `json:"name"`
		APIType           string `json:"api_type"`
		OpenAIEndpoint    string `json:"openai_endpoint"`
		AnthropicEndpoint string `json:"anthropic_endpoint"`
		APIKey            string `json:"api_key"`
		Models            string `json:"models"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	if body.Name == "" {
		respondError(w, 400, "validation_error", "name is required")
		return
	}
	if body.APIKey == "" {
		respondError(w, 400, "validation_error", "api_key is required")
		return
	}
	if body.Models == "" {
		respondError(w, 400, "validation_error", "models is required and must not be empty")
		return
	}
	if errMsg := validateProviderConfig(body.APIType, body.OpenAIEndpoint, body.AnthropicEndpoint); errMsg != "" {
		respondError(w, 400, "validation_error", errMsg)
		return
	}

	id, err := services.NewProviderManager().Create(body.Name, body.APIType, body.OpenAIEndpoint, body.AnthropicEndpoint, body.APIKey, body.Models)
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	provider, err := services.NewProviderManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve created provider")
		return
	}
	respondJSON(w, 201, provider)
}

func (h *AdminHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	if apiKey, ok := body["api_key"].(string); ok && apiKey != "" {
		encrypted, err := services.Encrypt(apiKey)
		if err != nil {
			respondError(w, 500, "internal_error", err.Error())
			return
		}
		body["api_key_encrypted"] = encrypted
		delete(body, "api_key")
	}

	provider, err := services.NewProviderManager().Get(id)
	if err != nil {
		respondError(w, 404, "not_found_error", "Provider not found")
		return
	}

	apiType := string(provider.APIType)
	if apiType == "" {
		apiType = string(models.APITypeOpenAIOnly)
	}
	openAIEndpoint := provider.OpenAIEndpoint
	anthropicEndpoint := provider.AnthropicEndpoint

	if value, ok := body["api_type"].(string); ok {
		apiType = value
	}
	if value, ok := body["openai_endpoint"].(string); ok {
		openAIEndpoint = value
	}
	if value, ok := body["anthropic_endpoint"].(string); ok {
		anthropicEndpoint = value
	}
	if errMsg := validateProviderConfig(apiType, openAIEndpoint, anthropicEndpoint); errMsg != "" {
		respondError(w, 400, "validation_error", errMsg)
		return
	}

	if err := services.NewProviderManager().Update(id, body); err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	provider, err = services.NewProviderManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve updated provider")
		return
	}
	respondJSON(w, 200, provider)
}

func validateProviderConfig(apiType, openAIEndpoint, anthropicEndpoint string) string {
	apiType = strings.TrimSpace(apiType)
	openAIEndpoint = strings.TrimSpace(openAIEndpoint)
	anthropicEndpoint = strings.TrimSpace(anthropicEndpoint)

	switch models.APIType(apiType) {
	case "":
		if openAIEndpoint == "" && anthropicEndpoint == "" {
			return "api_type is required"
		}
		return validateProviderConfig(string(models.APITypeOpenAIOnly), openAIEndpoint, anthropicEndpoint)
	case models.APITypeOpenAIOnly:
		if openAIEndpoint == "" {
			return "openai_endpoint is required for openai_only providers"
		}
	case models.APITypeAnthropicOnly:
		if anthropicEndpoint == "" {
			return "anthropic_endpoint is required for anthropic_only providers"
		}
	case models.APITypeBoth:
		if openAIEndpoint == "" {
			return "openai_endpoint is required for both providers"
		}
		if anthropicEndpoint == "" {
			return "anthropic_endpoint is required for both providers"
		}
	default:
		return "api_type must be one of: openai_only, anthropic_only, both"
	}

	return ""
}

func (h *AdminHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := services.NewProviderManager().Delete(id); err != nil {
		respondError(w, 500, "internal_error", "Failed to delete provider")
		return
	}
	w.WriteHeader(204)
}

func (h *AdminHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := services.NewKeyManager().List()
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}
	respondJSON(w, 200, keys)
}

func (h *AdminHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string     `json:"name"`
		RequestLimit *int       `json:"request_limit"`
		ExpiresAt    *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	if body.Name == "" {
		respondError(w, 400, "validation_error", "Name is required")
		return
	}

	id, rawKey, err := services.NewKeyManager().Create(body.Name, body.RequestLimit, body.ExpiresAt)
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	respondJSON(w, 201, map[string]interface{}{
		"id":      id,
		"api_key": rawKey,
		"name":    body.Name,
		"message": "API key created. Save this key - it will not be shown again.",
	})
}

func (h *AdminHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := services.NewKeyManager().Delete(id); err != nil {
		respondError(w, 500, "internal_error", "Failed to delete key")
		return
	}
	w.WriteHeader(204)
}

func (h *AdminHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "7"
	}
	days, _ := strconv.Atoi(daysStr)

	stats, err := services.NewUsageTracker().GetStats(days)
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}
	respondJSON(w, 200, stats)
}

type slogLogEntry struct {
	Time         string `json:"time"`
	Level        string `json:"level"`
	Msg          string `json:"msg"`
	Type         string `json:"type"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Status       int    `json:"status"`
	Latency      int64  `json:"latency"`
	KeyPrefix    string `json:"key_prefix"`
	Model        string `json:"model"`
	VirtualModel string `json:"virtual_model"`
	Provider     string `json:"provider"`
}

func (h *AdminHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}

	filterLevel := q.Get("level")
	filterKeyPrefix := q.Get("key_prefix")
	filterPath := q.Get("path")
	filterModel := q.Get("model")
	var filterStatus int
	if s := q.Get("status"); s != "" {
		filterStatus, _ = strconv.Atoi(s)
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		logDir = filepath.Join(homeDir, "Library", "Logs", "home-provider")
	}

	logFile := filepath.Join(logDir, "app-"+time.Now().Format("2006-01-02")+".log")

	var allLogs []slogLogEntry
	file, err := os.Open(logFile)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var entry slogLogEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				continue
			}
			if entry.Type != "inference" && entry.Type != "admin" {
				continue
			}
			allLogs = append(allLogs, entry)
		}
		file.Close()
	}

	var filtered []slogLogEntry
	for i := len(allLogs) - 1; i >= 0; i-- {
		e := allLogs[i]
		if filterLevel != "" && strings.ToUpper(e.Level) != strings.ToUpper(filterLevel) {
			continue
		}
		if filterKeyPrefix != "" && !strings.Contains(e.KeyPrefix, filterKeyPrefix) {
			continue
		}
		if filterPath != "" && !strings.Contains(e.Path, filterPath) {
			continue
		}
		if filterModel != "" && !strings.Contains(e.Model, filterModel) {
			continue
		}
		if filterStatus != 0 && e.Status != filterStatus {
			continue
		}
		filtered = append(filtered, e)
	}

	total := len(filtered)

	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	result := filtered[offset:end]

	respondJSON(w, 200, map[string]interface{}{
		"logs":   result,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

func (h *AdminHandler) CreateVirtualModel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		ProviderID string `json:"provider_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	if body.Name == "" {
		respondError(w, 400, "validation_error", "name is required")
		return
	}
	if body.ProviderID == "" {
		respondError(w, 400, "validation_error", "provider_id is required")
		return
	}

	id, err := services.NewVirtualModelManager().Create(body.Name, body.ProviderID)
	if err != nil {
		if err.Error() == "provider not found" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "virtual model with this name already exists" {
			respondError(w, 409, "conflict_error", err.Error())
			return
		}
		if err.Error() == "virtual model name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	virtualModel, err := services.NewVirtualModelManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve created virtual model")
		return
	}
	respondJSON(w, 201, virtualModel)
}

func (h *AdminHandler) ListVirtualModels(w http.ResponseWriter, r *http.Request) {
	virtualModels, err := services.NewVirtualModelManager().List()
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}
	respondJSON(w, 200, virtualModels)
}

func (h *AdminHandler) GetVirtualModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	virtualModel, err := services.NewVirtualModelManager().Get(id)
	if err != nil {
		respondError(w, 404, "not_found_error", "virtual model not found")
		return
	}
	respondJSON(w, 200, virtualModel)
}

func (h *AdminHandler) UpdateVirtualModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	err := services.NewVirtualModelManager().Update(id, body)
	if err != nil {
		if err.Error() == "virtual model not found" {
			respondError(w, 404, "not_found_error", err.Error())
			return
		}

		if err.Error() == "provider not found" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "virtual model name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "virtual model with this name already exists" {
			respondError(w, 409, "conflict_error", err.Error())
			return
		}
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	virtualModel, err := services.NewVirtualModelManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve updated virtual model")
		return
	}
	respondJSON(w, 200, virtualModel)
}

func (h *AdminHandler) DeleteVirtualModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := services.NewVirtualModelManager().Delete(id)
	if err != nil {
		if err.Error() == "virtual model not found" {
			respondError(w, 404, "not_found_error", "virtual model not found")
			return
		}
		if err.Error() == "cannot delete default virtual model" {
			respondError(w, 400, "validation_error", "cannot delete default virtual model")
			return
		}
		respondError(w, 500, "internal_error", "Failed to delete virtual model")
		return
	}
	w.WriteHeader(204)
}
