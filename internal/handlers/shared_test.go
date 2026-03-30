package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"home-provider/internal/models"
	"home-provider/internal/services"
)

type testEnv struct {
	tempDir string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "home-provider-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return &testEnv{tempDir: tempDir}
}

func (e *testEnv) createProviderFile(t *testing.T, providers []models.Provider) {
	t.Helper()
	providersPath := filepath.Join(e.tempDir, "data", "providers.json")
	if err := os.MkdirAll(filepath.Dir(providersPath), 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	data, err := json.Marshal(providers)
	if err != nil {
		t.Fatalf("failed to marshal providers: %v", err)
	}
	if err := os.WriteFile(providersPath, data, 0644); err != nil {
		t.Fatalf("failed to write providers file: %v", err)
	}
}

func (e *testEnv) createTagFile(t *testing.T, tags []models.Tag) {
	t.Helper()
	tagsPath := filepath.Join(e.tempDir, "data", "tags.json")
	if err := os.MkdirAll(filepath.Dir(tagsPath), 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	data, err := json.Marshal(tags)
	if err != nil {
		t.Fatalf("failed to marshal tags: %v", err)
	}
	if err := os.WriteFile(tagsPath, data, 0644); err != nil {
		t.Fatalf("failed to write tags file: %v", err)
	}
}

func (e *testEnv) cleanup() {
	os.RemoveAll(e.tempDir)
}

func testResolver(t *testing.T, tempDir string) (*services.ProviderManager, *services.TagManager) {
	originalDataDir := os.Getenv("DATA_DIR")
	os.Setenv("DATA_DIR", tempDir)
	t.Cleanup(func() {
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	})
	return services.NewProviderManager(), services.NewTagManager()
}

func TestResolveProvider_TagMatch(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{
		ID:   "test-provider",
		Name: "TestProvider",
	}
	env.createProviderFile(t, []models.Provider{provider})

	tag := models.Tag{
		ID:         "tag-1",
		Name:       "latest",
		ProviderID: "test-provider",
	}
	env.createTagFile(t, []models.Tag{tag})

	pm, tm := testResolver(t, env.tempDir)
	apiKey := &models.APIKey{}

	req := &http.Request{}
	result, err := resolveProvider(req, "latest", apiKey, pm, tm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "test-provider" {
		t.Fatalf("expected provider ID 'test-provider', got %s", result.ID)
	}
}

func TestResolveProvider_TagNotFound(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{
		ID:   "some-provider",
		Name: "SomeProvider",
	}
	env.createProviderFile(t, []models.Provider{provider})

	env.createTagFile(t, []models.Tag{})

	pm, tm := testResolver(t, env.tempDir)
	apiKey := &models.APIKey{}

	req := &http.Request{}
	_, err := resolveProvider(req, "unknown-model", apiKey, pm, tm)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "model not found in any tag" {
		t.Fatalf("expected error 'model not found in any tag', got %v", err)
	}
}

func TestResolveProvider_OrphanTag(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{
		ID:   "existing-provider",
		Name: "ExistingProvider",
	}
	env.createProviderFile(t, []models.Provider{provider})

	tag := models.Tag{
		ID:         "orphan-tag",
		Name:       "orphan",
		ProviderID: "non-existent-provider",
	}
	env.createTagFile(t, []models.Tag{tag})

	pm, tm := testResolver(t, env.tempDir)
	apiKey := &models.APIKey{}

	req := &http.Request{}
	_, err := resolveProvider(req, "orphan", apiKey, pm, tm)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "tag routing error: provider not found" {
		t.Fatalf("expected error 'tag routing error: provider not found', got %v", err)
	}
}

func TestResolveProvider_MultipleTags(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider1 := models.Provider{
		ID:   "provider-one",
		Name: "ProviderOne",
	}
	provider2 := models.Provider{
		ID:   "provider-two",
		Name: "ProviderTwo",
	}
	env.createProviderFile(t, []models.Provider{provider1, provider2})

	tags := []models.Tag{
		{ID: "tag-1", Name: "moe-3", ProviderID: "provider-one"},
		{ID: "tag-2", Name: "moe-3.1", ProviderID: "provider-two"},
	}
	env.createTagFile(t, tags)

	pm, tm := testResolver(t, env.tempDir)
	apiKey := &models.APIKey{}

	req := &http.Request{}
	result, err := resolveProvider(req, "moe-3.1", apiKey, pm, tm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "provider-two" {
		t.Fatalf("expected provider ID 'provider-two', got %s", result.ID)
	}
}

func TestResolveProvider_ExactTagMatch(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{
		ID:   "kimi-provider",
		Name: "Kimi",
	}
	env.createProviderFile(t, []models.Provider{provider})

	tag := models.Tag{
		ID:         "kimi-tag",
		Name:       "MiniMax-M2.7-highspeed",
		ProviderID: "kimi-provider",
	}
	env.createTagFile(t, []models.Tag{tag})

	pm, tm := testResolver(t, env.tempDir)
	apiKey := &models.APIKey{}

	req := &http.Request{}
	result, err := resolveProvider(req, "MiniMax-M2.7-highspeed", apiKey, pm, tm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "kimi-provider" {
		t.Fatalf("expected provider ID 'kimi-provider', got %s", result.ID)
	}
}
