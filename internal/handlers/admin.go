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
		Name        string `json:"name"`
		APIEndpoint string `json:"api_endpoint"`
		APIKey      string `json:"api_key"`
		Models      string `json:"models"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	if body.Name == "" {
		respondError(w, 400, "validation_error", "name is required")
		return
	}
	if body.APIEndpoint == "" {
		respondError(w, 400, "validation_error", "api_endpoint is required")
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

	id, err := services.NewProviderManager().Create(body.Name, body.APIEndpoint, body.APIKey, body.Models)
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

	if err := services.NewProviderManager().Update(id, body); err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	provider, err := services.NewProviderManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve updated provider")
		return
	}
	respondJSON(w, 200, provider)
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
	Time      string `json:"time"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Type      string `json:"type"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Latency   int64  `json:"latency"`
	KeyPrefix string `json:"key_prefix"`
	Model     string `json:"model"`
	Tag       string `json:"tag"`
	Provider  string `json:"provider"`
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

func (h *AdminHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
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

	id, err := services.NewTagManager().Create(body.Name, body.ProviderID)
	if err != nil {
		if err.Error() == "provider not found" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "tag with this name already exists" {
			respondError(w, 409, "conflict_error", err.Error())
			return
		}
		if err.Error() == "tag name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	tag, err := services.NewTagManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve created tag")
		return
	}
	respondJSON(w, 201, tag)
}

func (h *AdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := services.NewTagManager().List()
	if err != nil {
		respondError(w, 500, "internal_error", err.Error())
		return
	}
	respondJSON(w, 200, tags)
}

func (h *AdminHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tag, err := services.NewTagManager().Get(id)
	if err != nil {
		respondError(w, 404, "not_found_error", "tag not found")
		return
	}
	respondJSON(w, 200, tag)
}

func (h *AdminHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, 400, "invalid_request_error", "Invalid request body")
		return
	}

	err := services.NewTagManager().Update(id, body)
	if err != nil {
		if err.Error() == "tag not found" {
			respondError(w, 404, "not_found_error", err.Error())
			return
		}

		if err.Error() == "provider not found" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "tag name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$" {
			respondError(w, 400, "validation_error", err.Error())
			return
		}
		if err.Error() == "tag with this name already exists" {
			respondError(w, 409, "conflict_error", err.Error())
			return
		}
		respondError(w, 500, "internal_error", err.Error())
		return
	}

	tag, err := services.NewTagManager().Get(id)
	if err != nil {
		respondError(w, 500, "internal_error", "Failed to retrieve updated tag")
		return
	}
	respondJSON(w, 200, tag)
}

func (h *AdminHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := services.NewTagManager().Delete(id)
	if err != nil {
		if err.Error() == "tag not found" {
			respondError(w, 404, "not_found_error", "tag not found")
			return
		}
		if err.Error() == "cannot delete default tag" {
			respondError(w, 400, "validation_error", "cannot delete default tag")
			return
		}
		respondError(w, 500, "internal_error", "Failed to delete tag")
		return
	}
	w.WriteHeader(204)
}
