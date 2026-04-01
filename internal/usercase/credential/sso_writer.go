package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/esuEdu/cloud-swap/internal/domain"
)

func GetSSOTokenPath(name string) string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".cloud-swap", "sso", fmt.Sprintf("%s.json", name))
}

type SSOToken struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func SSOLogin(cred *domain.Credential) error {
	dir := filepath.Dir(GetSSOTokenPath(cred.Name))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create SSO dir: %w", err)
	}

	fmt.Printf("Starting SSO login for: %s\n", cred.Name)
	fmt.Printf("SSO Start URL: %s\n", cred.SSOStartURL)

	cmd := exec.Command("aws", "sso", "login",
		"--sso-session", cred.Name,
		"--start-url", cred.SSOStartURL,
		"--region", cred.Region,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aws sso login failed: %w", err)
	}

	fmt.Println("SSO login successful!")
	fmt.Printf("Run 'cloud-swap use %s' to activate credentials\n", cred.Name)
	return nil
}

func GetSSOCredentialsUsingAWSCLI(cred *domain.Credential) (*domain.Credential, error) {
	cmd := exec.Command("aws", "sso", "get-role-credentials",
		"--account-id", cred.SSOAccountID,
		"--role-name", cred.SSORoleName,
		"--session-name", "cloud-swap-"+cred.Name,
		"--region", cred.Region,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO credentials: %w (make sure 'aws sso login' was run)", err)
	}

	var result struct {
		RoleCredentials struct {
			AccessKeyId     string `json:"accessKeyId"`
			SecretAccessKey string `json:"secretAccessKey"`
			SessionToken    string `json:"sessionToken"`
			Expiration      int64  `json:"expiration"`
		} `json:"roleCredentials"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse SSO credentials: %w", err)
	}

	expiresAt := time.UnixMilli(result.RoleCredentials.Expiration).Format(time.RFC3339)

	return &domain.Credential{
		Name:       cred.Name,
		Provider:   "aws",
		AccessKey:  result.RoleCredentials.AccessKeyId,
		SecretKey:  result.RoleCredentials.SecretAccessKey,
		SessionTok: result.RoleCredentials.SessionToken,
		ExpiresAt:  expiresAt,
		Region:     cred.Region,
		Output:     "json",
	}, nil
}

func ActivateAWSSSO(c domain.Credential) error {
	cliPath := GetSSOTokenPath(c.Name)

	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		fmt.Println("SSO token not found. Running 'aws sso login'...")
		fmt.Println("You may need to authenticate in your browser.")
		fmt.Println("")

		homeDir, _ := os.UserHomeDir()
		ssoConfigDir := filepath.Join(homeDir, ".aws", "sso", "sessions")

		ssoSessionName := c.Name
		ssoConfig := fmt.Sprintf(`[%s]
sso_start_url = %s
sso_region = %s
`, ssoSessionName, c.SSOStartURL, c.Region)

		if err := os.MkdirAll(ssoConfigDir, 0700); err == nil {
			sessionFile := filepath.Join(ssoConfigDir, fmt.Sprintf("%s.json", ssoSessionName))
			os.WriteFile(sessionFile, []byte(ssoConfig), 0600)
		}
	}

	creds, err := GetSSOCredentialsUsingAWSCLI(&c)
	if err != nil {
		fmt.Printf("Failed to get SSO credentials: %v\n", err)
		fmt.Println("Please run: cloud-swap sso-login " + c.Name)
		return nil
	}

	if err := writeAwsFiles(*creds); err != nil {
		return err
	}

	fmt.Printf("AWS SSO credential '%s' is now active!\n", c.Name)
	fmt.Printf("Credentials expire at: %s\n", creds.ExpiresAt)
	return nil
}

func ExportAWSSSO(c domain.Credential) []string {
	creds, err := GetSSOCredentialsUsingAWSCLI(&c)
	if err != nil {
		return []string{
			fmt.Sprintf("# Error getting SSO credentials: %v", err),
			"# Run: cloud-swap sso-login " + c.Name,
		}
	}

	expiresAt := creds.ExpiresAt
	return []string{
		"# AWS SSO credentials",
		fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s", creds.AccessKey),
		fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=%s", creds.SecretKey),
		fmt.Sprintf("export AWS_SESSION_TOKEN=%s", creds.SessionTok),
		fmt.Sprintf("export AWS_DEFAULT_REGION=%s", creds.Region),
		fmt.Sprintf("# Expires at: %s", expiresAt),
	}
}

func IsSSOCredentialAlmostExpired(cred *domain.Credential) bool {
	if cred.ExpiresAt == "" {
		return false
	}

	parsed, err := time.Parse(time.RFC3339, cred.ExpiresAt)
	if err != nil {
		return true
	}

	refreshThreshold := 5 * time.Minute
	return time.Now().Add(refreshThreshold).After(parsed)
}

func IsAWSCLIInstalled() bool {
	cmd := exec.Command("aws", "--version")
	return cmd.Run() == nil
}

func HasValidSSOToken(cred *domain.Credential) bool {
	if !IsAWSCLIInstalled() {
		return false
	}

	cmd := exec.Command("aws", "sso", "get-role-credentials",
		"--account-id", cred.SSOAccountID,
		"--role-name", cred.SSORoleName,
		"--session-name", "cloud-swap-check",
	)

	err := cmd.Run()
	return err == nil
}
