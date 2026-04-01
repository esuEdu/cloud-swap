package domain

type Credential struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	SessionTok string `json:"session_token,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	Region     string `json:"region"`
	Output     string `json:"output"`

	// Multi-cloud
	TenantID       string `json:"tenant_id,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`

	// AWS SSO
	SSOStartURL    string `json:"sso_start_url,omitempty"`
	SSOAccountID   string `json:"sso_account_id,omitempty"`
	SSORoleName    string `json:"sso_role_name,omitempty"`
	SSOAccountName string `json:"sso_account_name,omitempty"`
}

type ProviderType string

const (
	ProviderAWS   ProviderType = "aws"
	ProviderGCP   ProviderType = "gcp"
	ProviderAzure ProviderType = "azure"
)

func (c *Credential) IsAWS() bool {
	return c.Provider == "aws"
}

func (c *Credential) IsGCP() bool {
	return c.Provider == "gcp"
}

func (c *Credential) IsAzure() bool {
	return c.Provider == "azure"
}

func (c *Credential) IsAWSSSO() bool {
	return c.Provider == "aws-sso"
}

func (c *Credential) IsSSO() bool {
	return c.SSOStartURL != "" && c.SSOAccountID != ""
}
