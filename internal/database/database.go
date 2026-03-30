package database

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func resolvePath(filePath string) string {
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		return filepath.Join(dataDir, filePath)
	}
	return filePath
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
		dir = "./data"
	}
	return os.MkdirAll(dir, 0755)
}
