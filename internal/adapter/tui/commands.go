package tui

import (
	"github.com/esuEdu/cloud-swap/internal/adapter/storage"
	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/usercase/credential"
)

func LoadCredentials() func() ([]domain.Credential, error) {
	return func() ([]domain.Credential, error) {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)
		creds, err := svc.List()
		if err != nil {
			return nil, err
		}
		return creds, nil
	}
}

func AddCredential(cred domain.Credential) func() ([]domain.Credential, error) {
	return func() ([]domain.Credential, error) {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)

		err := svc.Add(cred)
		if err != nil {
			return nil, err
		}

		creds, err := svc.List()
		if err != nil {
			return nil, err
		}
		return creds, nil
	}
}

func DeleteCredential(name string) func() ([]domain.Credential, error) {
	return func() ([]domain.Credential, error) {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)

		err := svc.Delete(name)
		if err != nil {
			return nil, err
		}

		creds, err := svc.List()
		if err != nil {
			return nil, err
		}
		return creds, nil
	}
}

func UseCredential(name string) func() (string, error) {
	return func() (string, error) {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)

		err := svc.Use(name)
		if err != nil {
			return "", err
		}
		return "Credential activated: " + name, nil
	}
}

func ValidateCredential(name string) func() (string, error) {
	return func() (string, error) {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)

		cred, err := svc.Get(name)
		if err != nil {
			return "", err
		}

		err = svc.Validate(cred)
		if err != nil {
			return "", err
		}
		return "Credential is valid!", nil
	}
}
