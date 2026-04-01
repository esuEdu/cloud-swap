package tui

import (
	"github.com/esuEdu/cloud-swap/internal/domain"
)

type (
	Page string
	Msg  string
)

const (
	PageList          Page = "list"
	PageAdd           Page = "add"
	PageEdit          Page = "edit"
	PageConfirmDelete Page = "confirm_delete"
)

type Model struct {
	Page          Page
	Credentials   []domain.Credential
	SelectedIndex int
	FormField     int
	Name          string
	Provider      string
	AccessKey     string
	SecretKey     string
	Region        string
	SessionTok    string
	ExpiresAt     string

	// GCP
	ProjectID string

	// Azure
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string

	// AWS SSO
	SSOStartURL  string
	SSOAccountID string
	SSORoleName  string

	DeleteTarget string
	Message      string
	Err          error
}

func InitialModel() Model {
	return Model{
		Page:          PageList,
		SelectedIndex: 0,
		Provider:      "aws",
		Region:        "us-east-1",
	}
}

type TickMsg struct{}
type RefreshMsg struct{}
type ErrorMsg error
type LoadedMsg []domain.Credential
