package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"home-provider/internal/database"
)

var encryptionKey []byte

func InitCrypto() error {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr != "" {
		if len(keyStr) != 32 {
			return errors.New("encryption key must be exactly 32 bytes")
		}
		encryptionKey = []byte(keyStr)
		return nil
	}

	keyFile := filepath.Join(database.DefaultDataDir(), ".encryption_key")
	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) == 32 {
		encryptionKey = data
		return nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("failed to generate random encryption key: %w", err)
	}
	os.MkdirAll(database.DefaultDataDir(), 0755)
	os.WriteFile(keyFile, key, 0600)
	encryptionKey = key
	return nil
}

func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
