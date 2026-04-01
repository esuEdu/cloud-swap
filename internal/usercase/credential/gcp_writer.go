package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esuEdu/cloud-swap/internal/domain"
)

type GCPCredentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func writeGCPCredentials(c domain.Credential) error {
	home := os.Getenv("HOME")
	configDir := filepath.Join(home, ".config", "gcloud")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create gcloud config dir: %w", err)
	}

	var creds map[string]interface{}
	if err := json.Unmarshal([]byte(c.AccessKey), &creds); err != nil {
		creds = map[string]interface{}{
			"type":         "service_account",
			"project_id":   c.ProjectID,
			"private_key":  c.AccessKey,
			"client_email": c.SecretKey,
		}
	}

	credBytes, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	credFile := filepath.Join(configDir, fmt.Sprintf("cloud-swap-%s.json", c.Name))
	if err := os.WriteFile(credFile, credBytes, 0600); err != nil {
		return fmt.Errorf("failed to write gcloud credentials: %w", err)
	}

	gcloudPath := filepath.Join(configDir, "application_default_credentials.json")
	if err := os.WriteFile(gcloudPath, credBytes, 0600); err != nil {
		return fmt.Errorf("failed to write ADC: %w", err)
	}

	exportCmd := fmt.Sprintf(`export GOOGLE_PROJECT=%s
export GOOGLE_APPLICATION_CREDENTIALS=%s`, c.ProjectID, credFile)
	fmt.Println(exportCmd)

	return nil
}

func GetGCPCredentialsPath(name string) string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "gcloud", fmt.Sprintf("cloud-swap-%s.json", name))
}

func ActivateGCP(c domain.Credential) error {
	if err := writeGCPCredentials(c); err != nil {
		return err
	}

	fmt.Printf("GCP credential '%s' is now active!\n", c.Name)
	fmt.Printf("Run: export GOOGLE_PROJECT=%s\n", c.ProjectID)
	return nil
}

func ExportGCP(c domain.Credential) []string {
	credPath := GetGCPCredentialsPath(c.Name)
	return []string{
		fmt.Sprintf("export GOOGLE_PROJECT=%s", c.ProjectID),
		fmt.Sprintf("export GOOGLE_APPLICATION_CREDENTIALS=%s", credPath),
	}
}

func IsGCPKey(key string) bool {
	return strings.Contains(key, "private_key") || strings.Contains(key, "client_email")
}
