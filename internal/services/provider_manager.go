package services

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"home-provider/internal/database"
	"home-provider/internal/models"

	"github.com/google/uuid"
)

type ProviderManager struct{}

var providerMgr = &ProviderManager{}

func NewProviderManager() *ProviderManager {
	return providerMgr
}

func (pm *ProviderManager) Create(name, apiEndpoint, apiKey string, model string) (string, error) {
	encryptedKey, err := Encrypt(apiKey)
	if err != nil {
		return "", err
	}

	provider := models.Provider{
		ID:            uuid.New().String(),
		Name:          name,
		APIEndpoint:   apiEndpoint,
		APIKeyEncrypt: encryptedKey,
		Models:        "\"" + model + "\"",
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return "", err
	}
	providers = append(providers, provider)
	if err := database.WriteJSON("./data/providers.json", providers); err != nil {
		return "", err
	}

	return provider.ID, nil
}

func (pm *ProviderManager) Get(id string) (*models.Provider, error) {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return nil, err
	}
	for _, p := range providers {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("record not found")
}

func (pm *ProviderManager) GetByName(name string) (*models.Provider, error) {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return nil, err
	}
	for _, p := range providers {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, errors.New("record not found")
}

func (pm *ProviderManager) List() ([]models.Provider, error) {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return nil, err
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].CreatedAt.After(providers[j].CreatedAt)
	})
	return providers, nil
}

func (pm *ProviderManager) ListActive() ([]models.Provider, error) {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return nil, err
	}
	var active []models.Provider
	for _, p := range providers {
		if p.IsActive {
			active = append(active, p)
		}
	}
	return active, nil
}

func (pm *ProviderManager) Update(id string, updates map[string]interface{}) error {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return err
	}
	found := false
	for i, p := range providers {
		if p.ID == id {
			found = true
			if name, ok := updates["name"].(string); ok {
				providers[i].Name = name
			}
			if apiEndpoint, ok := updates["api_endpoint"].(string); ok {
				providers[i].APIEndpoint = apiEndpoint
			}
			if apiKeyEncrypt, ok := updates["api_key_encrypt"].(string); ok {
				providers[i].APIKeyEncrypt = apiKeyEncrypt
			}
			if rawModels, ok := updates["models"]; ok {
				var modelsStr string
				switch v := rawModels.(type) {
				case string:
					modelsStr = "\"" + v + "\""
				case []interface{}:
					bs, _ := json.Marshal(v)
					modelsStr = string(bs)
				}
				if modelsStr != "" {
					providers[i].Models = modelsStr
				}
			}
			if isActive, ok := updates["is_active"].(bool); ok {
				providers[i].IsActive = isActive
			}
			providers[i].UpdatedAt = time.Now()
			break
		}
	}
	if !found {
		return errors.New("record not found")
	}
	return database.WriteJSON("./data/providers.json", providers)
}

func (pm *ProviderManager) Delete(id string) error {
	var providers []models.Provider
	if err := database.ReadJSON("./data/providers.json", &providers); err != nil {
		return err
	}
	filtered := providers[:0]
	for _, p := range providers {
		if p.ID != id {
			filtered = append(filtered, p)
		}
	}
	return database.WriteJSON("./data/providers.json", filtered)
}

func (pm *ProviderManager) GetDecryptedKey(provider *models.Provider) (string, error) {
	return Decrypt(provider.APIKeyEncrypt)
}

func (pm *ProviderManager) ListModels() ([]map[string]string, error) {
	providers, err := pm.ListActive()
	if err != nil {
		return nil, err
	}

	var result []map[string]string
	for _, p := range providers {
		result = append(result, map[string]string{
			"id":       p.ID + ":latest",
			"provider": p.Name,
		})
	}
	return result, nil
}

func (pm *ProviderManager) ResolveModel(provider *models.Provider, rawModel string) (string, error) {
	modelName := rawModel
	if strings.HasPrefix(rawModel, provider.ID+":") {
		modelName = rawModel[len(provider.ID)+1:]
	} else if strings.Contains(rawModel, ":") {
		return "", errors.New("model does not match provider")
	}
	if modelName == "latest" {
		var model string
		if err := json.Unmarshal([]byte(provider.Models), &model); err != nil {
			return "", err
		}
		if model == "" {
			return "", errors.New("no models configured for provider")
		}
		return model, nil
	}
	return modelName, nil
}
