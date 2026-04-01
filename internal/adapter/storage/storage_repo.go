package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/port"
)

type storageRepo struct {
}

func NewStorageRepo() port.Storage {
	return &storageRepo{}
}

func (repo *storageRepo) Save(creds []domain.Credential) error {

	path, err := ensureStorage()
	if err != nil {
		return err
	}

	wrapped := struct {
		Credential []domain.Credential `json:"credential"`
	}{Credential: creds}

	data, err := json.MarshalIndent(wrapped, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (repo *storageRepo) Load() ([]domain.Credential, error) {
	path, err := ensureStorage()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wrapped struct {
		Credential []domain.Credential `json:"credential"`
	}

	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}

	return wrapped.Credential, nil
}

func ensureStorage() (string, error) {
	path := storagePath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.WriteFile(path, []byte(`{"credential":[]}`), 0600)
	}

	return path, nil
}

func storagePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cloud-swap", "credentials.json")
}

func (repo *storageRepo) Backup(path string) error {
	src, err := ensureStorage()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (repo *storageRepo) Restore(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var wrapped struct {
		Credential []domain.Credential `json:"credential"`
	}

	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}

	return repo.Save(wrapped.Credential)
}

func (repo *storageRepo) BackupAll(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := repo.Backup(filepath.Join(dir, "credentials.json")); err != nil {
		return err
	}

	cfgRepo := NewConfigRepo()
	cfg, err := cfgRepo.Load()
	if err != nil {
		return err
	}

	cfgData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), cfgData, 0600); err != nil {
		return err
	}

	return nil
}

func (repo *storageRepo) RestoreAll(dir string) error {
	credsPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.json")

	if err := repo.Restore(credsPath); err != nil {
		return err
	}

	if _, err := os.Stat(cfgPath); err == nil {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return err
		}

		var cfg domain.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}

		cfgRepo := NewConfigRepo()
		if err := cfgRepo.Save(&cfg); err != nil {
			return err
		}
	}

	return nil
}
