package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/port"
)

type configRepo struct{}

func NewConfigRepo() port.ConfigStorage {
	return &configRepo{}
}

func (r *configRepo) Load() (*domain.Config, error) {
	path := configPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &domain.Config{
			DefaultRegion: "us-east-1",
			DefaultOutput: "json",
			FzfEnabled:    true,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.DefaultRegion == "" {
		cfg.DefaultRegion = "us-east-1"
	}
	if cfg.DefaultOutput == "" {
		cfg.DefaultOutput = "json"
	}

	return &cfg, nil
}

func (r *configRepo) Save(cfg *domain.Config) error {
	path, err := ensureConfigDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cloud-swap", "config.json")
}

func ensureConfigDir() (string, error) {
	path := configPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return path, nil
}
