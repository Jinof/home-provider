package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func DefaultDataDir() string {
	if dataDir := strings.TrimSpace(os.Getenv("DATA_DIR")); dataDir != "" {
		return dataDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return "./data"
	}

	return filepath.Join(homeDir, ".config", "home-provider")
}

func resolvePath(filePath string) string {
	dataDir := DefaultDataDir()

	switch {
	case filePath == "./data", filePath == "data":
		return dataDir
	case strings.HasPrefix(filePath, "./data/"):
		return filepath.Join(dataDir, strings.TrimPrefix(filePath, "./data/"))
	case strings.HasPrefix(filePath, "data/"):
		return filepath.Join(dataDir, strings.TrimPrefix(filePath, "data/"))
	default:
		return filePath
	}
}

func ReadJSON(filePath string, out interface{}) error {
	file, err := os.Open(resolvePath(filePath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(out)
}

func WriteJSON(filePath string, data interface{}) error {
	resolvedPath := resolvePath(filePath)
	tmpPath := resolvedPath + ".tmp"

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, resolvedPath)
}

func Init(dir string) error {
	if dir == "" {
		dir = DefaultDataDir()
	}
	return os.MkdirAll(dir, 0755)
}
