package services

import (
	"errors"
	"regexp"
	"sort"
	"sync"
	"time"

	"home-provider/internal/database"
	"home-provider/internal/models"

	"github.com/google/uuid"
)

var virtualModelManagerOnce sync.Once
var virtualModelManagerInstance *VirtualModelManager

const DefaultVirtualModelName = "default"
const virtualModelsPath = "./data/virtual_models.json"

var virtualModelNameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type VirtualModelManager struct{}

func NewVirtualModelManager() *VirtualModelManager {
	virtualModelManagerOnce.Do(func() {
		virtualModelManagerInstance = &VirtualModelManager{}
	})
	return virtualModelManagerInstance
}

func (vm *VirtualModelManager) Create(name, providerID string) (string, error) {
	if !virtualModelNameRegex.MatchString(name) {
		return "", errors.New("virtual model name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$")
	}

	providerMgr := NewProviderManager()
	if _, err := providerMgr.Get(providerID); err != nil {
		return "", errors.New("provider not found")
	}

	virtualModels, err := vm.load()
	if err != nil {
		return "", err
	}
	for _, virtualModel := range virtualModels {
		if virtualModel.Name == name {
			return "", errors.New("virtual model with this name already exists")
		}
	}

	virtualModel := models.VirtualModel{
		ID:         uuid.New().String(),
		Name:       name,
		ProviderID: providerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	virtualModels = append(virtualModels, virtualModel)
	if err := vm.save(virtualModels); err != nil {
		return "", err
	}

	return virtualModel.ID, nil
}

func (vm *VirtualModelManager) Get(id string) (*models.VirtualModel, error) {
	virtualModels, err := vm.load()
	if err != nil {
		return nil, err
	}
	for _, virtualModel := range virtualModels {
		if virtualModel.ID == id {
			return &virtualModel, nil
		}
	}
	return nil, errors.New("virtual model not found")
}

func (vm *VirtualModelManager) GetByName(name string) (*models.VirtualModel, error) {
	virtualModels, err := vm.load()
	if err != nil {
		return nil, err
	}
	for _, virtualModel := range virtualModels {
		if virtualModel.Name == name {
			return &virtualModel, nil
		}
	}
	return nil, nil
}

func (vm *VirtualModelManager) EnsureDefaultVirtualModel(providerID string) error {
	existing, err := vm.GetByName(DefaultVirtualModelName)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	if providerID == "" {
		providers, err := NewProviderManager().List()
		if err == nil && len(providers) > 0 {
			providerID = providers[0].ID
		}
	}

	if providerID != "" {
		providerMgr := NewProviderManager()
		if _, err := providerMgr.Get(providerID); err != nil {
			return err
		}
	}

	_, err = vm.Create(DefaultVirtualModelName, providerID)
	if err != nil && err.Error() == "virtual model with this name already exists" {
		return nil
	}
	return err
}

func (vm *VirtualModelManager) List() ([]models.VirtualModel, error) {
	virtualModels, err := vm.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(virtualModels, func(i, j int) bool {
		return virtualModels[i].CreatedAt.After(virtualModels[j].CreatedAt)
	})
	return virtualModels, nil
}

func (vm *VirtualModelManager) Update(id string, updates map[string]interface{}) error {
	virtualModels, err := vm.load()
	if err != nil {
		return err
	}

	found := false
	for i, virtualModel := range virtualModels {
		if virtualModel.ID != id {
			continue
		}

		found = true
		if name, ok := updates["name"].(string); ok {
			if !virtualModelNameRegex.MatchString(name) {
				return errors.New("virtual model name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$")
			}
			for _, existing := range virtualModels {
				if existing.Name == name && existing.ID != id {
					return errors.New("virtual model with this name already exists")
				}
			}
			virtualModels[i].Name = name
		}
		if providerID, ok := updates["provider_id"].(string); ok {
			providerMgr := NewProviderManager()
			if _, err := providerMgr.Get(providerID); err != nil {
				return errors.New("provider not found")
			}
			virtualModels[i].ProviderID = providerID
		}
		virtualModels[i].UpdatedAt = time.Now()
		break
	}
	if !found {
		return errors.New("virtual model not found")
	}

	return vm.save(virtualModels)
}

func (vm *VirtualModelManager) Delete(id string) error {
	virtualModels, err := vm.load()
	if err != nil {
		return err
	}
	for _, virtualModel := range virtualModels {
		if virtualModel.ID == id && virtualModel.Name == DefaultVirtualModelName {
			return errors.New("cannot delete default virtual model")
		}
	}

	filtered := virtualModels[:0]
	for _, virtualModel := range virtualModels {
		if virtualModel.ID != id {
			filtered = append(filtered, virtualModel)
		}
	}
	if len(filtered) == len(virtualModels) {
		return errors.New("virtual model not found")
	}

	return vm.save(filtered)
}

func (vm *VirtualModelManager) load() ([]models.VirtualModel, error) {
	var virtualModels []models.VirtualModel
	if err := database.ReadJSON(virtualModelsPath, &virtualModels); err != nil {
		return nil, err
	}
	return virtualModels, nil
}

func (vm *VirtualModelManager) save(virtualModels []models.VirtualModel) error {
	return database.WriteJSON(virtualModelsPath, virtualModels)
}
