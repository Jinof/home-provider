package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"home-provider/internal/database"
	"home-provider/internal/models"

	"github.com/google/uuid"
)

type KeyManager struct{}

var keyManager = &KeyManager{}

func NewKeyManager() *KeyManager {
	return keyManager
}

func (km *KeyManager) GenerateKey() (raw, hash, prefix string, err error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	raw = "hpk_" + hex.EncodeToString(bytes)

	hasher := sha256.New()
	hasher.Write([]byte(raw))
	hash = hex.EncodeToString(hasher.Sum(nil))

	prefix = raw[:12]
	return
}

func (km *KeyManager) Create(name string, requestLimit *int, expiresAt *time.Time) (string, string, error) {
	raw, hash, prefix, err := km.GenerateKey()
	if err != nil {
		return "", "", err
	}

	apiKey := models.APIKey{
		ID:           uuid.New().String(),
		KeyHash:      hash,
		KeyPrefix:    prefix,
		Name:         name,
		RequestLimit: requestLimit,
		ExpiresAt:    expiresAt,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return "", "", err
	}

	keys = append(keys, apiKey)

	if err := database.WriteJSON("./data/api_keys.json", keys); err != nil {
		return "", "", err
	}

	return apiKey.ID, raw, nil
}

func (km *KeyManager) Validate(rawKey string) (*models.APIKey, error) {
	if !strings.HasPrefix(rawKey, "hpk_") {
		return nil, fmt.Errorf("invalid key format")
	}

	hasher := sha256.New()
	hasher.Write([]byte(rawKey))
	hash := hex.EncodeToString(hasher.Sum(nil))

	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.KeyHash == hash && key.IsActive {
			if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("key expired")
			}
			return &key, nil
		}
	}

	return nil, fmt.Errorf("key not found")
}

func (km *KeyManager) List() ([]models.APIKey, error) {
	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return nil, err
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[j].CreatedAt.After(keys[i].CreatedAt)
	})

	return keys, nil
}

func (km *KeyManager) Revoke(id string) error {
	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return err
	}

	found := false
	for i := range keys {
		if keys[i].ID == id {
			keys[i].IsActive = false
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("key not found")
	}

	return database.WriteJSON("./data/api_keys.json", keys)
}

func (km *KeyManager) Get(id string) (*models.APIKey, error) {
	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return nil, err
	}
	for _, k := range keys {
		if k.ID == id {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("key not found")
}

func (km *KeyManager) Delete(id string) error {
	var keys []models.APIKey
	if err := database.ReadJSON("./data/api_keys.json", &keys); err != nil {
		return err
	}

	originalLen := len(keys)
	keys = filterKeysByID(keys, id)

	if len(keys) == originalLen {
		return fmt.Errorf("key not found")
	}

	return database.WriteJSON("./data/api_keys.json", keys)
}

func filterKeysByID(keys []models.APIKey, id string) []models.APIKey {
	result := make([]models.APIKey, 0, len(keys))
	for _, key := range keys {
		if key.ID != id {
			result = append(result, key)
		}
	}
	return result
}
