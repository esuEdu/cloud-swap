package credential

import (
	"testing"

	"github.com/esuEdu/cloud-swap/internal/domain"
)

type mockStorage struct {
	creds []domain.Credential
}

func (m *mockStorage) Load() ([]domain.Credential, error) {
	return m.creds, nil
}

func (m *mockStorage) Save(creds []domain.Credential) error {
	m.creds = creds
	return nil
}

func (m *mockStorage) Backup(path string) error {
	return nil
}

func (m *mockStorage) Restore(path string) error {
	return nil
}

func (m *mockStorage) BackupAll(dir string) error {
	return nil
}

func (m *mockStorage) RestoreAll(dir string) error {
	return nil
}

func (m *mockStorage) Add(c domain.Credential) error {
	m.creds = append(m.creds, c)
	return nil
}

func TestAddCredential(t *testing.T) {
	storage := &mockStorage{}
	svc := NewService(storage)

	cred := domain.Credential{
		Name:      "test",
		Provider:  "aws",
		AccessKey: "AKIATEST",
		SecretKey: "testsecret",
		Region:    "us-east-1",
	}

	err := svc.Add(cred)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if len(storage.creds) != 1 {
		t.Fatalf("Expected 1 credential, got %d", len(storage.creds))
	}

	if storage.creds[0].Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", storage.creds[0].Name)
	}
}

func TestAddDuplicateCredential(t *testing.T) {
	storage := &mockStorage{
		creds: []domain.Credential{{Name: "test", Provider: "aws"}},
	}
	svc := NewService(storage)

	cred := domain.Credential{Name: "test", Provider: "aws"}
	err := svc.Add(cred)

	if err == nil {
		t.Error("Expected error for duplicate")
	}
}

func TestListCredentials(t *testing.T) {
	storage := &mockStorage{
		creds: []domain.Credential{
			{Name: "test1", Provider: "aws"},
			{Name: "test2", Provider: "gcp"},
		},
	}
	svc := NewService(storage)

	creds, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(creds) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(creds))
	}
}

func TestDeleteCredential(t *testing.T) {
	storage := &mockStorage{
		creds: []domain.Credential{
			{Name: "test1", Provider: "aws"},
			{Name: "test2", Provider: "gcp"},
		},
	}
	svc := NewService(storage)

	err := svc.Delete("test1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if len(storage.creds) != 1 {
		t.Errorf("Expected 1 credential after delete, got %d", len(storage.creds))
	}

	if storage.creds[0].Name != "test2" {
		t.Errorf("Expected remaining credential to be 'test2', got '%s'", storage.creds[0].Name)
	}
}

func TestDeleteNotFound(t *testing.T) {
	storage := &mockStorage{}
	svc := NewService(storage)

	err := svc.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent credential")
	}
}

func TestGetCredential(t *testing.T) {
	storage := &mockStorage{
		creds: []domain.Credential{
			{Name: "test", Provider: "aws", Region: "us-east-1"},
		},
	}
	svc := NewService(storage)

	cred, err := svc.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if cred.Name != "test" {
		t.Errorf("Expected 'test', got '%s'", cred.Name)
	}

	if cred.Region != "us-east-1" {
		t.Errorf("Expected 'us-east-1', got '%s'", cred.Region)
	}
}

func TestGetCredentialNotFound(t *testing.T) {
	storage := &mockStorage{}
	svc := NewService(storage)

	_, err := svc.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent credential")
	}
}
