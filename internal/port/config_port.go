package port

import "github.com/esuEdu/cloud-swap/internal/domain"

type ConfigStorage interface {
	Load() (*domain.Config, error)
	Save(*domain.Config) error
}
