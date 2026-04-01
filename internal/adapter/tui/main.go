package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/esuEdu/cloud-swap/internal/adapter/storage"
	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/usercase/credential"
)

var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
)

const (
	stepSelectProvider int = iota
	stepFillForm
)

var providers = []struct {
	id   string
	name string
	desc string
}{
	{"aws", "AWS", "Access Key + Secret Key"},
	{"aws-sso", "AWS SSO", "AWS Single Sign-On"},
	{"gcp", "GCP", "Google Cloud Platform"},
	{"azure", "Azure", "Microsoft Azure"},
}

type model struct {
	creds            []domain.Credential
	step             int
	selectedProvider int

	nameInput       textinput.Model
	providerInput   textinput.Model
	regionInput     textinput.Model
	accessKeyInput  textinput.Model
	secretKeyInput  textinput.Model
	projectInput    textinput.Model
	tenantInput     textinput.Model
	clientIDInput   textinput.Model
	clientSecInput  textinput.Model
	subIDInput      textinput.Model
	ssoURLInput     textinput.Model
	ssoAccountInput textinput.Model
	ssoRoleInput    textinput.Model

	message string
	err     error
	focus   int
}

func RunTUI() error {
	s := storage.NewStorageRepo()
	svc := credential.NewService(s)
	creds, err := svc.List()
	if err != nil {
		return err
	}

	m := model{
		creds:            creds,
		step:             stepSelectProvider,
		selectedProvider: 0,
	}

	m.initInputs()

	p := tea.NewProgram(&m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case refreshMsg:
		m.creds = msg.Creds
		m.step = stepSelectProvider
		m.message = msg.Message
		m.err = nil
		return m, nil

	case errorMsg:
		m.err = error(msg)
		return m, nil

	case successMsg:
		m.message = string(msg)
		return m, nil

	case tea.KeyMsg:
		return m.updateKey(msg)
	}

	return m, nil
}

func (m *model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "r":
		return m.refresh()

	case "d":
		return m.deleteSelected()

	case "v":
		return m.validateSelected()
	}

	if m.step == stepSelectProvider {
		switch msg.String() {
		case "up", "k":
			m.selectedProvider = (m.selectedProvider - 1 + len(providers)) % len(providers)
		case "down", "j":
			m.selectedProvider = (m.selectedProvider + 1) % len(providers)
		case "enter":
			m.providerInput.SetValue(providers[m.selectedProvider].id)
			m.step = stepFillForm
			m.focus = 0
			m.nameInput.Focus()
		}
		return m, nil
	}

	// stepFillForm
	switch msg.String() {
	case "esc":
		m.step = stepSelectProvider
		m.message = ""
		m.err = nil
	case "enter":
		return m, m.saveCredential()
	case "tab":
		m.focus = (m.focus + 1) % m.fieldCount()
	case "shift+tab":
		m.focus = (m.focus - 1 + m.fieldCount()) % m.fieldCount()
	}

	provider := m.providerInput.Value()
	switch provider {
	case "aws":
		switch m.focus {
		case 0:
			m.nameInput, _ = m.nameInput.Update(msg)
		case 1:
			m.regionInput, _ = m.regionInput.Update(msg)
		case 2:
			m.accessKeyInput, _ = m.accessKeyInput.Update(msg)
		case 3:
			m.secretKeyInput, _ = m.secretKeyInput.Update(msg)
		}
	case "aws-sso":
		switch m.focus {
		case 0:
			m.nameInput, _ = m.nameInput.Update(msg)
		case 1:
			m.regionInput, _ = m.regionInput.Update(msg)
		case 2:
			m.ssoURLInput, _ = m.ssoURLInput.Update(msg)
		case 3:
			m.ssoAccountInput, _ = m.ssoAccountInput.Update(msg)
		case 4:
			m.ssoRoleInput, _ = m.ssoRoleInput.Update(msg)
		}
	case "gcp":
		switch m.focus {
		case 0:
			m.nameInput, _ = m.nameInput.Update(msg)
		case 1:
			m.regionInput, _ = m.regionInput.Update(msg)
		case 2:
			m.projectInput, _ = m.projectInput.Update(msg)
		case 3:
			m.accessKeyInput, _ = m.accessKeyInput.Update(msg)
		}
	case "azure":
		switch m.focus {
		case 0:
			m.nameInput, _ = m.nameInput.Update(msg)
		case 1:
			m.regionInput, _ = m.regionInput.Update(msg)
		case 2:
			m.tenantInput, _ = m.tenantInput.Update(msg)
		case 3:
			m.clientIDInput, _ = m.clientIDInput.Update(msg)
		case 4:
			m.clientSecInput, _ = m.clientSecInput.Update(msg)
		case 5:
			m.subIDInput, _ = m.subIDInput.Update(msg)
		}
	}

	return m, nil
}

func (m *model) fieldCount() int {
	provider := m.providerInput.Value()
	switch provider {
	case "aws":
		return 4
	case "aws-sso":
		return 5
	case "gcp":
		return 4
	case "azure":
		return 6
	default:
		return 4
	}
}

func (m *model) refresh() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)
		creds, _ := svc.List()
		return refreshMsg{Creds: creds}
	}
}

func (m *model) deleteSelected() (tea.Model, tea.Cmd) {
	if len(m.creds) == 0 {
		return m, nil
	}

	idx := 0
	name := m.creds[idx].Name
	return m, func() tea.Msg {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)
		svc.Delete(name)
		creds, _ := svc.List()
		return refreshMsg{Creds: creds, Message: "Deleted: " + name}
	}
}

func (m *model) validateSelected() (tea.Model, tea.Cmd) {
	if len(m.creds) == 0 {
		return m, nil
	}

	idx := 0
	name := m.creds[idx].Name
	return m, func() tea.Msg {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)
		cred, _ := svc.Get(name)
		if cred == nil {
			return errorMsg(fmt.Errorf("credential not found"))
		}
		err := svc.Validate(cred)
		if err != nil {
			return errorMsg(err)
		}
		return successMsg("Credential is valid!")
	}
}

func (m *model) saveCredential() tea.Cmd {
	provider := m.providerInput.Value()

	return func() tea.Msg {
		s := storage.NewStorageRepo()
		svc := credential.NewService(s)

		cred := domain.Credential{
			Name:     m.nameInput.Value(),
			Provider: provider,
			Region:   m.regionInput.Value(),
		}

		switch provider {
		case "aws":
			cred.AccessKey = m.accessKeyInput.Value()
			cred.SecretKey = m.secretKeyInput.Value()
		case "aws-sso":
			cred.SSOStartURL = m.ssoURLInput.Value()
			cred.SSOAccountID = m.ssoAccountInput.Value()
			cred.SSORoleName = m.ssoRoleInput.Value()
		case "gcp":
			cred.ProjectID = m.projectInput.Value()
			cred.AccessKey = m.accessKeyInput.Value()
		case "azure":
			cred.TenantID = m.tenantInput.Value()
			cred.ClientID = m.clientIDInput.Value()
			cred.ClientSecret = m.clientSecInput.Value()
			cred.SubscriptionID = m.subIDInput.Value()
		}

		err := svc.Add(cred)
		if err != nil {
			return errorMsg(err)
		}

		creds, err := svc.List()
		if err != nil {
			return errorMsg(err)
		}
		return refreshMsg{Creds: creds, Message: "Credential added!"}
	}
}

func (m *model) View() string {
	if m.step == stepSelectProvider {
		return m.viewProviderSelect()
	}
	return m.viewAddForm()
}

func (m *model) viewProviderSelect() string {
	s := titleStyle.Render("\n  Cloud Swap - Add Credential") + "\n\n"
	s += dimStyle.Render("  Select a provider with arrow keys, press Enter:\n\n")

	for i, p := range providers {
		prefix := "    "
		if i == m.selectedProvider {
			prefix = "  ▶ "
			s += selectedItemStyle.Render(fmt.Sprintf("%s%s", prefix, p.name))
		} else {
			s += itemStyle.Render(fmt.Sprintf("%s%s", prefix, p.name))
		}
		s += dimStyle.Render(fmt.Sprintf(" - %s\n", p.desc))
	}

	s += "\n" + helpStyle.Render("  [enter]select  [r]refresh  [q]quit")

	if m.message != "" {
		s += "\n" + successStyle.Render("  "+m.message)
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("  Error: %v", m.err))
	}

	if len(m.creds) > 0 {
		s += "\n\n" + titleStyle.Render("  Saved Credentials:")
		for _, c := range m.creds {
			s += fmt.Sprintf("\n    • %s (%s) - %s", c.Name, c.Provider, c.Region)
		}
	}

	return s
}

func (m *model) viewAddForm() string {
	provider := m.providerInput.Value()
	s := titleStyle.Render(fmt.Sprintf("\n  Add %s Credential", provider)) + "\n\n"

	s += helpStyle.Render("  [tab]next field  [enter]save  [esc]back\n\n")

	switch provider {
	case "aws":
		s += m.renderField(0, "Name", m.nameInput, "my-profile")
		s += m.renderField(1, "Region", m.regionInput, "us-east-1")
		s += m.renderField(2, "Access Key", m.accessKeyInput, "AKIA...")
		s += m.renderField(3, "Secret Key", m.secretKeyInput, "***")

	case "aws-sso":
		s += m.renderField(0, "Name", m.nameInput, "my-sso-profile")
		s += m.renderField(1, "Region", m.regionInput, "us-east-1")
		s += m.renderField(2, "SSO Start URL", m.ssoURLInput, "https://d-xxx.awsapps.com/start")
		s += m.renderField(3, "Account ID", m.ssoAccountInput, "123456789012")
		s += m.renderField(4, "Role Name", m.ssoRoleInput, "AdministratorAccess")

	case "gcp":
		s += m.renderField(0, "Name", m.nameInput, "my-gcp-profile")
		s += m.renderField(1, "Region", m.regionInput, "us-central1")
		s += m.renderField(2, "Project ID", m.projectInput, "my-project-id")
		s += m.renderField(3, "Credentials JSON", m.accessKeyInput, "{...}")

	case "azure":
		s += m.renderField(0, "Name", m.nameInput, "my-azure-profile")
		s += m.renderField(1, "Region", m.regionInput, "eastus")
		s += m.renderField(2, "Tenant ID", m.tenantInput, "tenant-id")
		s += m.renderField(3, "Client ID", m.clientIDInput, "client-id")
		s += m.renderField(4, "Client Secret", m.clientSecInput, "***")
		s += m.renderField(5, "Subscription ID", m.subIDInput, "sub-id")
	}

	if m.message != "" {
		s += "\n" + successStyle.Render("  "+m.message)
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("  Error: %v", m.err))
	}

	return s
}

func (m *model) renderField(focus int, label string, input textinput.Model, placeholder string) string {
	prefix := "    "
	if m.focus == focus {
		prefix = "  ▶ "
	}

	value := input.Value()
	if value == "" {
		value = dimStyle.Render(placeholder)
	}

	return fmt.Sprintf("%s%s: %s\n", prefix, label, value)
}

func (m *model) initInputs() {
	m.nameInput = textinput.New()
	m.nameInput.Placeholder = "profile-name"

	m.regionInput = textinput.New()
	m.regionInput.Placeholder = "us-east-1"

	m.providerInput = textinput.New()
	m.providerInput.Placeholder = "aws"

	m.accessKeyInput = textinput.New()
	m.accessKeyInput.Placeholder = "AKIA..."

	m.secretKeyInput = textinput.New()
	m.secretKeyInput.Placeholder = "secret"
	m.secretKeyInput.EchoMode = textinput.EchoPassword

	m.projectInput = textinput.New()
	m.projectInput.Placeholder = "project-id"

	m.tenantInput = textinput.New()
	m.tenantInput.Placeholder = "tenant-id"

	m.clientIDInput = textinput.New()
	m.clientIDInput.Placeholder = "client-id"

	m.clientSecInput = textinput.New()
	m.clientSecInput.Placeholder = "client-secret"
	m.clientSecInput.EchoMode = textinput.EchoPassword

	m.subIDInput = textinput.New()
	m.subIDInput.Placeholder = "subscription-id"

	m.ssoURLInput = textinput.New()
	m.ssoURLInput.Placeholder = "https://d-xxx.awsapps.com/start"

	m.ssoAccountInput = textinput.New()
	m.ssoAccountInput.Placeholder = "123456789012"

	m.ssoRoleInput = textinput.New()
	m.ssoRoleInput.Placeholder = "AdministratorAccess"

	m.nameInput.Focus()
}

type refreshMsg struct {
	Creds   []domain.Credential
	Message string
}

type errorMsg error
type successMsg string
