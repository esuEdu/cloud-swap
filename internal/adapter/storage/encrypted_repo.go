package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/esuEdu/cloud-swap/internal/domain"
)

type EncryptedStorage struct {
	password string
}

func NewEncryptedStorage(password string) *EncryptedStorage {
	return &EncryptedStorage{password: password}
}

func (e *EncryptedStorage) deriveKey() []byte {
	hash := sha256.Sum256([]byte(e.password))
	return hash[:]
}

func (e *EncryptedStorage) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.deriveKey())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (e *EncryptedStorage) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.deriveKey())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonceBytes, ciphertextBody := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonceBytes, ciphertextBody, nil)
}

func (e *EncryptedStorage) Save(creds []domain.Credential) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	encrypted, err := e.encrypt(data)
	if err != nil {
		return err
	}

	path, err := ensureStorage()
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(encrypted)), 0600)
}

func (e *EncryptedStorage) Load() ([]domain.Credential, error) {
	path, err := ensureStorage()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	encrypted, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	decrypted, err := e.decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	var creds []domain.Credential
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func (e *EncryptedStorage) Backup(path string) error {
	creds, err := e.Load()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (e *EncryptedStorage) Restore(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var creds []domain.Credential
	if err := json.Unmarshal(data, &creds); err != nil {
		return err
	}

	return e.Save(creds)
}

func (e *EncryptedStorage) BackupAll(dir string) error {
	return fmt.Errorf("encrypted storage does not support full backup")
}

func (e *EncryptedStorage) RestoreAll(dir string) error {
	return fmt.Errorf("encrypted storage does not support full restore")
}

func (e *EncryptedStorage) Delete(name string) error {
	creds, err := e.Load()
	if err != nil {
		return err
	}

	newCreds := make([]domain.Credential, 0)
	for _, c := range creds {
		if c.Name != name {
			newCreds = append(newCreds, c)
		}
	}

	return e.Save(newCreds)
}
