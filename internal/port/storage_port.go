package port

import "github.com/esuEdu/cloud-swap/internal/domain"

type Storage interface {
	Save(creds []domain.Credential) error
	Load() ([]domain.Credential, error)
	Backup(path string) error
	Restore(path string) error
	BackupAll(dir string) error
	RestoreAll(dir string) error
}
