package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"home-provider/internal/models"
)

type virtualModelTestEnv struct {
	dataDir string
}

func setupVirtualModelTestEnv(t *testing.T) *virtualModelTestEnv {
	t.Helper()
	dataDir, err := os.MkdirTemp("", "home-provider-virtual-model-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return &virtualModelTestEnv{dataDir: dataDir}
}

func (e *virtualModelTestEnv) setDataDir() {
	os.Setenv("DATA_DIR", e.dataDir)
}

func (e *virtualModelTestEnv) unsetDataDir() {
	os.Unsetenv("DATA_DIR")
}

func (e *virtualModelTestEnv) writeJSON(t *testing.T, name string, value interface{}) {
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

func (e *virtualModelTestEnv) createProviderFile(t *testing.T, providers []models.Provider) {
	e.writeJSON(t, "providers.json", providers)
}

func (e *virtualModelTestEnv) cleanup() {
	os.RemoveAll(e.dataDir)
}

func TestVirtualModelManager_Create(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider := models.Provider{
		ID:        "test-provider-1",
		Name:      "TestProvider",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider})

	vm := NewVirtualModelManager()

	tests := []struct {
		name        string
		modelName   string
		providerID  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "create valid virtual model",
			modelName:  "valid-model",
			providerID: "test-provider-1",
		},
		{
			name:        "create with invalid name format",
			modelName:   "Invalid_Name",
			providerID:  "test-provider-1",
			wantErr:     true,
			errContains: "must match pattern",
		},
		{
			name:        "create with uppercase letters",
			modelName:   "UPPERCASE",
			providerID:  "test-provider-1",
			wantErr:     true,
			errContains: "must match pattern",
		},
		{
			name:        "create with spaces",
			modelName:   "has spaces",
			providerID:  "test-provider-1",
			wantErr:     true,
			errContains: "must match pattern",
		},
		{
			name:        "create with non-existent provider",
			modelName:   "some-model",
			providerID:  "non-existent-provider",
			wantErr:     true,
			errContains: "provider not found",
		},
		{
			name:       "create duplicate name",
			modelName:  "duplicate-test",
			providerID: "test-provider-1",
		},
		{
			name:        "create duplicate name again",
			modelName:   "duplicate-test",
			providerID:  "test-provider-1",
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := vm.Create(tt.modelName, tt.providerID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if id == "" {
				t.Error("expected non-empty ID")
			}
		})
	}
}

func TestVirtualModelManager_Get(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider := models.Provider{
		ID:        "test-provider-get",
		Name:      "TestProvider",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider})

	vm := NewVirtualModelManager()
	virtualModelID, _ := vm.Create("get-test-model", "test-provider-get")

	t.Run("get existing virtual model by ID", func(t *testing.T) {
		virtualModel, err := vm.Get(virtualModelID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if virtualModel.ID != virtualModelID {
			t.Errorf("expected ID %q, got %q", virtualModelID, virtualModel.ID)
		}
		if virtualModel.Name != "get-test-model" {
			t.Errorf("expected name %q, got %q", "get-test-model", virtualModel.Name)
		}
	})

	t.Run("get non-existent virtual model", func(t *testing.T) {
		_, err := vm.Get("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent virtual model, got nil")
		}
	})
}

func TestVirtualModelManager_GetByName(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider := models.Provider{
		ID:        "test-provider-byname",
		Name:      "TestProvider",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider})

	vm := NewVirtualModelManager()
	_, _ = vm.Create("byname-test-model", "test-provider-byname")

	t.Run("get by name existing", func(t *testing.T) {
		virtualModel, err := vm.GetByName("byname-test-model")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if virtualModel == nil {
			t.Fatal("expected virtual model, got nil")
		}
		if virtualModel.Name != "byname-test-model" {
			t.Errorf("expected name %q, got %q", "byname-test-model", virtualModel.Name)
		}
	})

	t.Run("get by name non-existent", func(t *testing.T) {
		virtualModel, err := vm.GetByName("non-existent-model")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if virtualModel != nil {
			t.Errorf("expected nil virtual model, got %+v", virtualModel)
		}
	})
}

func TestVirtualModelManager_List(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider := models.Provider{
		ID:        "test-provider-list",
		Name:      "TestProvider",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider})

	vm := NewVirtualModelManager()

	_, _ = vm.Create("first-model", "test-provider-list")
	time.Sleep(10 * time.Millisecond)
	_, _ = vm.Create("second-model", "test-provider-list")
	time.Sleep(10 * time.Millisecond)
	_, _ = vm.Create("third-model", "test-provider-list")

	virtualModels, err := vm.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(virtualModels) != 3 {
		t.Fatalf("expected 3 virtual models, got %d", len(virtualModels))
	}

	if virtualModels[0].Name != "third-model" {
		t.Errorf("expected first virtual model (newest) to be %q, got %q", "third-model", virtualModels[0].Name)
	}
	if virtualModels[2].Name != "first-model" {
		t.Errorf("expected last virtual model (oldest) to be %q, got %q", "first-model", virtualModels[2].Name)
	}
}

func TestVirtualModelManager_Update(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider1 := models.Provider{
		ID:        "test-provider-update-1",
		Name:      "TestProvider1",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	provider2 := models.Provider{
		ID:        "test-provider-update-2",
		Name:      "TestProvider2",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider1, provider2})

	vm := NewVirtualModelManager()
	virtualModelID, _ := vm.Create("update-test-model", "test-provider-update-1")

	t.Run("update name", func(t *testing.T) {
		err := vm.Update(virtualModelID, map[string]interface{}{"name": "updated-model-name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		virtualModel, _ := vm.Get(virtualModelID)
		if virtualModel.Name != "updated-model-name" {
			t.Errorf("expected name %q, got %q", "updated-model-name", virtualModel.Name)
		}
	})

	t.Run("update provider_id", func(t *testing.T) {
		err := vm.Update(virtualModelID, map[string]interface{}{"provider_id": "test-provider-update-2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		virtualModel, _ := vm.Get(virtualModelID)
		if virtualModel.ProviderID != "test-provider-update-2" {
			t.Errorf("expected provider_id %q, got %q", "test-provider-update-2", virtualModel.ProviderID)
		}
	})

	t.Run("update non-existent virtual model", func(t *testing.T) {
		err := vm.Update("non-existent-id", map[string]interface{}{"name": "new-name"})
		if err == nil {
			t.Error("expected error for non-existent virtual model, got nil")
		}
	})

	t.Run("update with invalid name", func(t *testing.T) {
		err := vm.Update(virtualModelID, map[string]interface{}{"name": "Invalid_Name"})
		if err == nil {
			t.Error("expected error for invalid name, got nil")
		}
	})

	t.Run("update with duplicate name", func(t *testing.T) {
		_, _ = vm.Create("another-model", "test-provider-update-1")
		err := vm.Update(virtualModelID, map[string]interface{}{"name": "another-model"})
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})
}

func TestVirtualModelManager_Delete(t *testing.T) {
	env := setupVirtualModelTestEnv(t)
	defer env.cleanup()
	env.setDataDir()
	defer env.unsetDataDir()

	provider := models.Provider{
		ID:        "test-provider-delete",
		Name:      "TestProvider",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.createProviderFile(t, []models.Provider{provider})

	vm := NewVirtualModelManager()
	virtualModelID, _ := vm.Create("delete-test-model", "test-provider-delete")

	t.Run("delete existing virtual model", func(t *testing.T) {
		err := vm.Delete(virtualModelID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = vm.Get(virtualModelID)
		if err == nil {
			t.Error("expected error after deleting virtual model, got nil")
		}
	})

	t.Run("delete non-existent virtual model", func(t *testing.T) {
		err := vm.Delete("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent virtual model, got nil")
		}
	})
}
