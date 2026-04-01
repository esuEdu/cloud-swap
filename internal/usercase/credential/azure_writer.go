package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/esuEdu/cloud-swap/internal/domain"
)

func writeAzureCredentials(c domain.Credential) error {
	home := os.Getenv("HOME")
	configDir := filepath.Join(home, ".cloud-swap", "azure")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create azure config dir: %w", err)
	}

	azureCreds := map[string]interface{}{
		"tenantId":       c.TenantID,
		"clientId":       c.ClientID,
		"clientSecret":   c.ClientSecret,
		"subscriptionId": c.SubscriptionID,
	}

	credBytes, err := json.MarshalIndent(azureCreds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal azure credentials: %w", err)
	}

	credFile := filepath.Join(configDir, fmt.Sprintf("%s.json", c.Name))
	if err := os.WriteFile(credFile, credBytes, 0600); err != nil {
		return fmt.Errorf("failed to write azure credentials: %w", err)
	}

	exportCmd := fmt.Sprintf(`export AZURE_TENANT_ID=%s
export AZURE_CLIENT_ID=%s
export AZURE_CLIENT_SECRET=%s
export AZURE_SUBSCRIPTION_ID=%s`, c.TenantID, c.ClientID, c.ClientSecret, c.SubscriptionID)
	fmt.Println(exportCmd)

	return nil
}

func GetAzureCredentialsPath(name string) string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".cloud-swap", "azure", fmt.Sprintf("%s.json", name))
}

func ActivateAzure(c domain.Credential) error {
	if err := writeAzureCredentials(c); err != nil {
		return err
	}

	fmt.Printf("Azure credential '%s' is now active!\n", c.Name)
	fmt.Printf("Run: cloud-swap use %s\n", c.Name)
	return nil
}

func ExportAzure(c domain.Credential) []string {
	return []string{
		fmt.Sprintf("export AZURE_TENANT_ID=%s", c.TenantID),
		fmt.Sprintf("export AZURE_CLIENT_ID=%s", c.ClientID),
		fmt.Sprintf("export AZURE_CLIENT_SECRET=%s", c.ClientSecret),
		fmt.Sprintf("export AZURE_SUBSCRIPTION_ID=%s", c.SubscriptionID),
	}
}
