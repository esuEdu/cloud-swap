package credential

import (
	"fmt"
	"time"

	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/port"
)

type Service struct {
	Storage port.Storage
}

func NewService(s port.Storage) *Service {
	return &Service{
		Storage: s,
	}
}

func (s *Service) Add(c domain.Credential) error {
	creds, err := s.Storage.Load()
	if err != nil {
		return err
	}

	for _, existing := range creds {
		if existing.Name == c.Name {
			return fmt.Errorf("credential already exists: %s", c.Name)
		}
	}

	creds = append(creds, c)

	if err := s.Storage.Save(creds); err != nil {
		return err
	}

	return nil
}

func (s *Service) List() ([]domain.Credential, error) {
	return s.Storage.Load()
}

func (s *Service) Get(name string) (*domain.Credential, error) {
	creds, err := s.Storage.Load()
	if err != nil {
		return nil, err
	}

	for _, c := range creds {
		if c.Name == name {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("credential not found: %s", name)
}

func (s *Service) Update(name string, updated domain.Credential) error {
	creds, err := s.Storage.Load()
	if err != nil {
		return err
	}

	found := false
	for i, c := range creds {
		if c.Name == name {
			creds[i] = updated
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("credential not found: %s", name)
	}

	return s.Storage.Save(creds)
}

func (s *Service) Delete(name string) error {
	creds, err := s.Storage.Load()
	if err != nil {
		return err
	}

	newCreds := make([]domain.Credential, 0, len(creds))
	found := false

	for _, c := range creds {
		if c.Name != name {
			newCreds = append(newCreds, c)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("credential not found: %s", name)
	}

	return s.Storage.Save(newCreds)
}

func (s *Service) Use(name string) error {
	creds, err := s.Storage.Load()
	if err != nil {
		return err
	}

	var selected *domain.Credential
	for _, c := range creds {
		if c.Name == name {
			selected = &c
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("credential not found: %s", name)
	}

	switch selected.Provider {
	case "gcp":
		return ActivateGCP(*selected)
	case "azure":
		return ActivateAzure(*selected)
	case "aws-sso":
		return ActivateAWSSSO(*selected)
	case "aws", "":
		return writeAwsFiles(*selected)
	default:
		return fmt.Errorf("unsupported provider: %s", selected.Provider)
	}
}

func (s *Service) Assume(roleArn, profile string) (*domain.Credential, error) {
	allCreds, err := s.Storage.Load()
	if err != nil {
		return nil, err
	}

	var baseCred *domain.Credential
	for _, c := range allCreds {
		if c.Name == profile {
			baseCred = &c
			break
		}
	}

	if baseCred == nil {
		return nil, fmt.Errorf("profile not found: %s", profile)
	}

	assumedCreds, err := assumeRole(roleArn, baseCred)
	if err != nil {
		return nil, err
	}

	if err := writeAwsFiles(*assumedCreds); err != nil {
		return nil, err
	}

	return assumedCreds, nil
}

func (s *Service) IsExpired(expiresAt string) bool {
	parsed, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return false
	}
	return time.Now().After(parsed)
}

func (s *Service) Validate(c *domain.Credential) error {
	return validateCredential(c)
}
