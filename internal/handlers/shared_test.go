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
	dataDir string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dataDir, err := os.MkdirTemp("", "home-provider-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return &testEnv{dataDir: dataDir}
}

func (e *testEnv) writeJSON(t *testing.T, name string, value interface{}) {
	t.Helper()
	path := filepath.Join(e.dataDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal %s: %v", name, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

func (e *testEnv) createProviderFile(t *testing.T, providers []models.Provider) {
	e.writeJSON(t, "providers.json", providers)
}

func (e *testEnv) createVirtualModelFile(t *testing.T, virtualModels []models.VirtualModel) {
	e.writeJSON(t, "virtual_models.json", virtualModels)
}

func (e *testEnv) cleanup() {
	os.RemoveAll(e.dataDir)
}

func testResolver(t *testing.T, dataDir string) (*services.ProviderManager, *services.VirtualModelManager) {
	originalDataDir := os.Getenv("DATA_DIR")
	os.Setenv("DATA_DIR", dataDir)
	t.Cleanup(func() {
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	})
	return services.NewProviderManager(), services.NewVirtualModelManager()
}

func TestResolveProvider_VirtualModelMatch(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{ID: "test-provider", Name: "TestProvider"}
	env.createProviderFile(t, []models.Provider{provider})

	virtualModel := models.VirtualModel{ID: "virtual-model-1", Name: "latest", ProviderID: "test-provider"}
	env.createVirtualModelFile(t, []models.VirtualModel{virtualModel})

	pm, vm := testResolver(t, env.dataDir)
	req := &http.Request{}
	result, err := ResolveProvider(req, "latest", pm, vm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider.ID != "test-provider" {
		t.Fatalf("expected provider ID 'test-provider', got %s", result.Provider.ID)
	}
	if result.VirtualModel == nil || result.VirtualModel.Name != "latest" {
		t.Fatalf("expected virtual model 'latest', got %+v", result.VirtualModel)
	}
}

func TestResolveProvider_VirtualModelNotFound(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{ID: "some-provider", Name: "SomeProvider"}
	env.createProviderFile(t, []models.Provider{provider})
	env.createVirtualModelFile(t, []models.VirtualModel{})

	pm, vm := testResolver(t, env.dataDir)
	req := &http.Request{}
	_, err := ResolveProvider(req, "unknown-model", pm, vm)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "virtual model not found" {
		t.Fatalf("expected error 'virtual model not found', got %v", err)
	}
}

func TestResolveProvider_OrphanVirtualModel(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{ID: "existing-provider", Name: "ExistingProvider"}
	env.createProviderFile(t, []models.Provider{provider})

	virtualModel := models.VirtualModel{ID: "orphan-virtual-model", Name: "orphan", ProviderID: "non-existent-provider"}
	env.createVirtualModelFile(t, []models.VirtualModel{virtualModel})

	pm, vm := testResolver(t, env.dataDir)
	req := &http.Request{}
	_, err := ResolveProvider(req, "orphan", pm, vm)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "provider not found for virtual model" {
		t.Fatalf("expected error 'provider not found for virtual model', got %v", err)
	}
}

func TestResolveProvider_MultipleVirtualModels(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider1 := models.Provider{ID: "provider-one", Name: "ProviderOne"}
	provider2 := models.Provider{ID: "provider-two", Name: "ProviderTwo"}
	env.createProviderFile(t, []models.Provider{provider1, provider2})

	virtualModels := []models.VirtualModel{
		{ID: "virtual-model-1", Name: "moe-3", ProviderID: "provider-one"},
		{ID: "virtual-model-2", Name: "moe-3.1", ProviderID: "provider-two"},
	}
	env.createVirtualModelFile(t, virtualModels)

	pm, vm := testResolver(t, env.dataDir)
	req := &http.Request{}
	result, err := ResolveProvider(req, "moe-3.1", pm, vm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider.ID != "provider-two" {
		t.Fatalf("expected provider ID 'provider-two', got %s", result.Provider.ID)
	}
}

func TestResolveProvider_ExactVirtualModelMatch(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	provider := models.Provider{ID: "kimi-provider", Name: "Kimi"}
	env.createProviderFile(t, []models.Provider{provider})

	virtualModel := models.VirtualModel{ID: "kimi-virtual-model", Name: "minimax-m2-7-highspeed", ProviderID: "kimi-provider"}
	env.createVirtualModelFile(t, []models.VirtualModel{virtualModel})

	pm, vm := testResolver(t, env.dataDir)
	req := &http.Request{}
	result, err := ResolveProvider(req, "minimax-m2-7-highspeed", pm, vm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider.ID != "kimi-provider" {
		t.Fatalf("expected provider ID 'kimi-provider', got %s", result.Provider.ID)
	}
}
