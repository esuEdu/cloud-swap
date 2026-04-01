package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/esuEdu/cloud-swap/internal/adapter/storage"
	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/usercase/credential"
)

type Cli struct {
	svc    *credential.Service
	cfg    *domain.Config
	silent bool
}

func NewCli(s *credential.Service) *Cli {
	cfg, _ := storage.NewConfigRepo().Load()
	return &Cli{svc: s, cfg: cfg}
}

func NewCliWithArgs(s *credential.Service, silent bool, name, accessKey, secretKey, region string) *Cli {
	c := NewCli(s)
	c.silent = silent
	if silent {
		c.silentAdd(name, accessKey, secretKey, region)
	}
	return c
}

func (c *Cli) silentAdd(name, accessKey, secretKey, region string) {
	cred := domain.Credential{
		Name:      name,
		Provider:  "aws",
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Output:    c.cfg.DefaultOutput,
	}

	if err := c.svc.Add(cred); err != nil {
		fmt.Printf("Failed to add credential: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Credential added successfully!")
}

func (c *Cli) Add(temp bool) {
	provider := promptWithDefault("provider (aws/aws-sso/gcp/azure)", "aws")

	var cred domain.Credential
	cred.Name = promptRequired("name")
	cred.Provider = provider

	switch provider {
	case "aws":
		region := promptWithDefault("region", c.cfg.DefaultRegion)
		cred.AccessKey = promptRequired("access_key")
		cred.SecretKey = promptRequired("secret_key")
		cred.Region = region
		cred.Output = c.withDefault("json", c.cfg.DefaultOutput)

		if temp {
			cred.SessionTok = prompt("session_token")
			cred.ExpiresAt = prompt("expires_at (e.g., 2025-01-27T10:30:00Z)")
		}

	case "aws-sso":
		region := promptWithDefault("region", c.cfg.DefaultRegion)
		cred.SSOStartURL = promptRequired("sso_start_url (e.g., https://d-xxx.awsapps.com/start)")
		cred.SSOAccountID = promptRequired("sso_account_id")
		cred.SSORoleName = promptRequired("sso_role_name (e.g., AdministratorAccess)")
		cred.Region = region

	case "gcp":
		cred.ProjectID = promptRequired("project_id")
		cred.AccessKey = promptRequired("credentials (JSON key content)")
		cred.SecretKey = ""
		cred.Region = promptWithDefault("region", "us-central1")

	case "azure":
		cred.TenantID = promptRequired("tenant_id")
		cred.ClientID = promptRequired("client_id")
		cred.ClientSecret = promptRequired("client_secret")
		cred.SubscriptionID = promptRequired("subscription_id")
		cred.Region = promptWithDefault("region", "eastus")

	default:
		fmt.Printf("Unknown provider: %s\n", provider)
		return
	}

	if err := c.svc.Add(cred); err != nil {
		fmt.Printf("Failed to add credential: %v\n", err)
		return
	}

	fmt.Println("Credential added successfully!")
}

func (c *Cli) Validate(name string) error {
	cred, err := c.svc.Get(name)
	if err != nil {
		return err
	}
	return c.svc.Validate(cred)
}

func (c *Cli) withDefault(value, defaultValue string) string {
	if value == "" && defaultValue != "" {
		return defaultValue
	}
	return value
}

func (c *Cli) List() {
	creds, err := c.svc.List()
	if err != nil {
		fmt.Printf("Failed to load credentials: %v\n", err)
		return
	}

	if len(creds) == 0 {
		fmt.Println("No credentials found.")
		return
	}

	for i, cred := range creds {
		expired := ""
		if cred.ExpiresAt != "" && c.svc.IsExpired(cred.ExpiresAt) {
			expired = " [EXPIRED]"
		}

		extra := ""
		if cred.Provider == "gcp" {
			extra = fmt.Sprintf(" project=%s", cred.ProjectID)
		} else if cred.Provider == "azure" {
			extra = fmt.Sprintf(" tenant=%s", cred.TenantID)
		}

		fmt.Printf("[%d] %s (%s) region=%s%s%s\n", i+1, cred.Name, cred.Provider, cred.Region, expired, extra)
	}
}

func (c *Cli) Update() {
	name := prompt("enter name to update")

	cred := domain.Credential{
		Name:       name,
		Provider:   "aws",
		AccessKey:  prompt("new access_key"),
		SecretKey:  prompt("new secret_key"),
		SessionTok: prompt("new session_token"),
		Region:     prompt("new region"),
		Output:     "json",
	}

	if err := c.svc.Update(name, cred); err != nil {
		fmt.Printf("Failed to update credential: %v\n", err)
		return
	}

	fmt.Println("Credential updated successfully!")
}

func (c *Cli) Delete() {
	name := prompt("enter name to delete")

	if err := c.svc.Delete(name); err != nil {
		fmt.Printf("Failed to delete credential: %v\n", err)
		return
	}

	fmt.Println("Credential deleted successfully!")
}

func (c *Cli) Use(name string, export bool) {
	cred, err := c.svc.Get(name)
	if err != nil {
		fmt.Printf("Failed to get credential: %v\n", err)
		return
	}

	if export {
		switch cred.Provider {
		case "gcp":
			for _, line := range credential.ExportGCP(*cred) {
				fmt.Println(line)
			}
		case "azure":
			for _, line := range credential.ExportAzure(*cred) {
				fmt.Println(line)
			}
		case "aws-sso":
			for _, line := range credential.ExportAWSSSO(*cred) {
				fmt.Println(line)
			}
		case "aws", "":
			fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", cred.AccessKey)
			fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", cred.SecretKey)
			if cred.SessionTok != "" {
				fmt.Printf("export AWS_SESSION_TOKEN=%s\n", cred.SessionTok)
			}
			fmt.Printf("export AWS_DEFAULT_REGION=%s\n", cred.Region)
		}
		return
	}

	if err := c.svc.Use(name); err != nil {
		fmt.Printf("Failed to activate credential: %v\n", err)
		return
	}

	fmt.Printf("AWS credential '%s' is now active!\n", name)
}

func (c *Cli) Assume(roleArn, profile string) {
	creds, err := c.svc.Assume(roleArn, profile)
	if err != nil {
		fmt.Printf("Failed to assume role: %v\n", err)
		return
	}

	fmt.Printf("Successfully assumed role: %s\n", roleArn)
	fmt.Printf("Credential '%s' is now active (expires: %s)\n", creds.Name, creds.ExpiresAt)
}

func (c *Cli) Interactive() {
	creds, err := c.svc.List()
	if err != nil {
		fmt.Printf("Failed to load credentials: %v\n", err)
		return
	}

	if len(creds) == 0 {
		fmt.Println("No credentials found. Run 'cloud-swap add' first.")
		return
	}

	names := make([]string, len(creds))
	for i, cred := range creds {
		names[i] = cred.Name
	}

	selected := c.fuzzySelect(names)
	if selected == "" {
		return
	}

	c.Use(selected, false)
}

func (c *Cli) fuzzySelect(options []string) string {
	cmd, err := exec.LookPath("fzf")
	if err != nil {
		return c.menuSelect(options)
	}

	input := strings.Join(options, "\n") + "\n"
	proc := exec.Command(cmd)
	proc.Stdin = strings.NewReader(input)
	output, err := proc.Output()
	if err != nil {
		return ""
	}

	selected := strings.TrimSpace(string(output))
	for _, opt := range options {
		if opt == selected {
			return selected
		}
	}

	return ""
}

func (c *Cli) menuSelect(options []string) string {
	fmt.Println("Select a credential:")
	for i, opt := range options {
		fmt.Printf("%d. %s\n", i+1, opt)
	}
	fmt.Print("Enter number: ")

	var idx int
	if _, err := fmt.Scanln(&idx); err != nil || idx < 1 || idx > len(options) {
		return ""
	}

	return options[idx-1]
}

func prompt(label string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(label + ": ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}

func promptRequired(label string) string {
	for {
		value := prompt(label)
		if value != "" {
			return value
		}
		fmt.Printf("Error: %s is required\n", label)
	}
}

func promptWithDefault(label, defaultValue string) string {
	value := prompt(label)
	if value == "" && defaultValue != "" {
		return defaultValue
	}
	return value
}
