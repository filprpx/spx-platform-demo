package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"spx/internal/platform"
)

const (
	wizardIdentityStep = iota
	wizardComputeStep
	wizardNetworkStep
	wizardScalingStep
	wizardReviewStep
	defaultApplicationName = "payments-api"
	defaultOwningTeam      = "finance-engineering"
	defaultContainerPort   = 8080
)

var (
	applicationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,62}$`)
	teamIdentifierPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,127}$`)
	wizardOptionStyle      = lipgloss.NewStyle().PaddingLeft(2)
	wizardSelectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	wizardMutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	wizardErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type createWizardModel struct {
	request      platform.CreateApplicationRequest
	step         int
	field        int
	option       int
	reviewChoice int
	err          string
	canceled     bool
	confirmed    bool
	nameInput    textinput.Model
	teamInput    textinput.Model
	portInput    textinput.Model
}

func newCreateWizardModel(initial platform.CreateApplicationRequest) createWizardModel {
	nameValue := initial.Name
	teamValue := initial.OwningTeam
	portValue := initial.ContainerPort
	if initial.Name == "" {
		initial.Name = defaultApplicationName
	}
	if initial.OwningTeam == "" {
		initial.OwningTeam = defaultOwningTeam
	}
	if initial.ComputeSize == "" {
		initial.ComputeSize = "small"
	}
	if initial.ContainerPort == 0 {
		initial.ContainerPort = defaultContainerPort
	}
	if initial.Ingress == "" {
		initial.Ingress = "external"
	}
	if initial.MaxReplicas == 0 {
		initial.MaxReplicas = 1
	}

	nameInput := textinput.New()
	nameInput.Prompt = ""
	nameInput.Placeholder = "payments-api (default)"
	nameInput.SetValue(nameValue)
	teamInput := textinput.New()
	teamInput.Prompt = ""
	teamInput.Placeholder = "finance-engineering (default)"
	teamInput.SetValue(teamValue)
	portInput := textinput.New()
	portInput.Prompt = ""
	portInput.Placeholder = "8080 (default)"
	if portValue != 0 {
		portInput.SetValue(strconv.Itoa(portValue))
	}

	model := createWizardModel{
		request:   initial,
		nameInput: nameInput,
		teamInput: teamInput,
		portInput: portInput,
	}
	if initial.Ingress == "internal" {
		model.option = 1
	}
	model.prepareStep()
	return model
}

func runCreateWizard(ctx context.Context, in io.Reader, out io.Writer, initial platform.CreateApplicationRequest) (platform.CreateApplicationRequest, error) {
	program := tea.NewProgram(newCreateWizardModel(initial), tea.WithContext(ctx), tea.WithInput(in), tea.WithOutput(out))
	finalModel, err := program.Run()
	if err != nil {
		return platform.CreateApplicationRequest{}, fmt.Errorf("run application wizard: %w", err)
	}
	model := finalModel.(createWizardModel)
	if model.canceled {
		return platform.CreateApplicationRequest{}, errors.New("application creation cancelled")
	}
	if !model.confirmed {
		return platform.CreateApplicationRequest{}, errors.New("application creation was not confirmed")
	}
	return model.request, nil
}

func (m createWizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m createWizardModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		case "esc":
			if m.step == wizardIdentityStep {
				m.canceled = true
				return m, tea.Quit
			}
			m.step--
			m.err = ""
			m.prepareStep()
			return m, nil
		case "b":
			canGoBack := m.step == wizardNetworkStep || m.step == wizardScalingStep || m.step == wizardReviewStep || (m.step == wizardComputeStep && m.field == 0)
			if canGoBack {
				m.step--
				m.err = ""
				m.prepareStep()
				return m, nil
			}
		}
	}

	switch m.step {
	case wizardIdentityStep:
		return m.updateIdentity(message)
	case wizardComputeStep:
		return m.updateCompute(message)
	case wizardNetworkStep:
		return m.updateNetwork(message)
	case wizardScalingStep:
		return m.updateScaling(message)
	case wizardReviewStep:
		return m.updateReview(message)
	default:
		return m, nil
	}
}

func (m createWizardModel) updateIdentity(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab", "shift+tab":
			m.syncIdentityField()
			m.field = 1 - m.field
			m.prepareStep()
			return m, nil
		case "enter":
			if m.field == 0 {
				m.request.Name = strings.TrimSpace(m.nameInput.Value())
				if m.request.Name == "" {
					m.request.Name = defaultApplicationName
				}
				if err := validateApplicationName(m.request.Name); err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.field = 1
				m.prepareStep()
				return m, nil
			}
			m.request.OwningTeam = strings.TrimSpace(m.teamInput.Value())
			if m.request.OwningTeam == "" {
				m.request.OwningTeam = defaultOwningTeam
			}
			if err := validateOwningTeam(m.request.OwningTeam); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.step++
			m.field = 0
			m.err = ""
			m.prepareStep()
			return m, nil
		}
	}
	var command tea.Cmd
	if m.field == 0 {
		m.nameInput, command = m.nameInput.Update(message)
	} else {
		m.teamInput, command = m.teamInput.Update(message)
	}
	return m, command
}

func (m createWizardModel) updateCompute(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "down":
			if m.field == 0 {
				m.option = 1 - m.option
				m.request.ComputeSize = []string{"small", "medium"}[m.option]
				return m, nil
			}
		case "tab", "shift+tab":
			m.syncComputeField()
			m.field = 1 - m.field
			m.prepareStep()
			return m, nil
		case "enter":
			if m.field == 0 {
				m.field = 1
				m.prepareStep()
				return m, nil
			}
			portValue := strings.TrimSpace(m.portInput.Value())
			if portValue == "" {
				portValue = strconv.Itoa(defaultContainerPort)
			}
			port, err := strconv.Atoi(portValue)
			if err != nil || port < 1 || port > 65535 {
				m.err = "container port must be a number from 1 to 65535"
				return m, nil
			}
			m.request.ContainerPort = port
			m.step++
			m.field = 0
			m.err = ""
			m.prepareStep()
			return m, nil
		}
	}
	if m.field == 1 {
		var command tea.Cmd
		m.portInput, command = m.portInput.Update(message)
		return m, command
	}
	return m, nil
}

func (m createWizardModel) updateNetwork(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "down":
			m.option = 1 - m.option
			m.request.Ingress = []string{"external", "internal"}[m.option]
		case "enter":
			m.step++
			m.field = 0
			m.option = minReplicaOptionIndex(m.request.MinReplicas)
			m.err = ""
			m.prepareStep()
		}
	}
	return m, nil
}

func (m createWizardModel) updateScaling(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "down":
			if m.field == 0 {
				m.option = 1 - m.option
				m.request.MinReplicas = []int{0, 1}[m.option]
			} else {
				if key.String() == "up" {
					m.option = (m.option + 2) % 3
				} else {
					m.option = (m.option + 1) % 3
				}
				m.request.MaxReplicas = []int{1, 2, 5}[m.option]
			}
		case "tab", "shift+tab":
			m.field = 1 - m.field
			m.option = maxReplicaOptionIndex(m.request.MaxReplicas)
			if m.field == 0 {
				m.option = minReplicaOptionIndex(m.request.MinReplicas)
			}
		case "enter":
			if m.field == 0 {
				m.field = 1
				m.option = maxReplicaOptionIndex(m.request.MaxReplicas)
				return m, nil
			}
			if m.request.MinReplicas > m.request.MaxReplicas {
				m.err = "minimum replicas cannot exceed maximum replicas"
				return m, nil
			}
			m.step++
			m.reviewChoice = 0
			m.err = ""
		}
	}
	return m, nil
}

func (m createWizardModel) updateReview(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "down", "tab":
			m.reviewChoice = 1 - m.reviewChoice
		case "enter":
			if m.reviewChoice == 0 {
				m.confirmed = true
			} else {
				m.canceled = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *createWizardModel) prepareStep() {
	m.nameInput.Blur()
	m.teamInput.Blur()
	m.portInput.Blur()
	if m.step == wizardIdentityStep {
		if m.field == 0 {
			m.nameInput.Focus()
		} else {
			m.teamInput.Focus()
		}
	}
	if m.step == wizardComputeStep && m.field == 1 {
		m.portInput.Focus()
	}
	if m.step == wizardComputeStep && m.field == 0 {
		m.option = map[string]int{"small": 0, "medium": 1}[m.request.ComputeSize]
	}
	if m.step == wizardNetworkStep {
		m.option = map[string]int{"external": 0, "internal": 1}[m.request.Ingress]
	}
}

func (m *createWizardModel) syncIdentityField() {
	if m.field == 0 {
		m.request.Name = strings.TrimSpace(m.nameInput.Value())
		if m.request.Name == "" {
			m.request.Name = defaultApplicationName
		}
	} else {
		m.request.OwningTeam = strings.TrimSpace(m.teamInput.Value())
		if m.request.OwningTeam == "" {
			m.request.OwningTeam = defaultOwningTeam
		}
	}
}

func (m *createWizardModel) syncComputeField() {
	if m.field == 1 {
		portValue := strings.TrimSpace(m.portInput.Value())
		if portValue == "" {
			portValue = strconv.Itoa(defaultContainerPort)
		}
		if port, err := strconv.Atoi(portValue); err == nil {
			m.request.ContainerPort = port
		}
	}
}

func (m createWizardModel) View() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s\n\n", titleStyle.Render("Create application"))
	fmt.Fprintf(&builder, "%s\n", wizardMutedStyle.Render(fmt.Sprintf("Step %d of 5", m.step+1)))

	switch m.step {
	case wizardIdentityStep:
		fmt.Fprintln(&builder, titleStyle.Render("Application identity"))
		wizardField(&builder, "Application name", m.nameInput.View(), m.field == 0)
		wizardField(&builder, "Owning team", m.teamInput.View(), m.field == 1)
	case wizardComputeStep:
		fmt.Fprintln(&builder, titleStyle.Render("Compute configuration"))
		wizardOption(&builder, "Compute size", []string{"Small — 0.25 vCPU, 0.5 GiB memory", "Medium — 0.5 vCPU, 1 GiB memory"}, m.option, m.field == 0)
		wizardField(&builder, "Container port", m.portInput.View(), m.field == 1)
	case wizardNetworkStep:
		fmt.Fprintln(&builder, titleStyle.Render("Network configuration"))
		wizardOption(&builder, "Network visibility", []string{"Public", "Internal"}, m.option, true)
	case wizardScalingStep:
		fmt.Fprintln(&builder, titleStyle.Render("Scaling configuration"))
		wizardOption(&builder, "Minimum replicas", []string{"0", "1"}, minReplicaOptionIndex(m.request.MinReplicas), m.field == 0)
		wizardOption(&builder, "Maximum replicas", []string{"1", "2", "5"}, maxReplicaOptionIndex(m.request.MaxReplicas), m.field == 1)
	case wizardReviewStep:
		fmt.Fprintln(&builder, titleStyle.Render("Review application"))
		writeReviewLine(&builder, "Name", m.request.Name)
		writeReviewLine(&builder, "Owning team", m.request.OwningTeam)
		writeReviewLine(&builder, "Compute size", m.request.ComputeSize)
		writeReviewLine(&builder, "Container port", strconv.Itoa(m.request.ContainerPort))
		writeReviewLine(&builder, "Ingress", m.request.Ingress)
		writeReviewLine(&builder, "Minimum replicas", strconv.Itoa(m.request.MinReplicas))
		writeReviewLine(&builder, "Maximum replicas", strconv.Itoa(m.request.MaxReplicas))
		wizardOption(&builder, "Confirm", []string{"Create application", "Cancel"}, m.reviewChoice, true)
	}

	if m.err != "" {
		fmt.Fprintf(&builder, "\n%s\n", wizardErrorStyle.Render(m.err))
	}
	fmt.Fprintf(&builder, "\n%s\n", wizardMutedStyle.Render("Enter: continue   Tab: next field   Esc/B: back   Ctrl+C: cancel"))
	return builder.String()
}

func wizardField(builder *strings.Builder, label, value string, active bool) {
	marker := "  "
	if active {
		marker = "> "
	}
	fmt.Fprintf(builder, "%s%s\n    %s\n", marker, labelStyle.Render(label), wizardOptionStyle.Render(value))
}

func wizardOption(builder *strings.Builder, label string, options []string, selected int, active bool) {
	fmt.Fprintf(builder, "%s\n", labelStyle.Render(label))
	for index, option := range options {
		marker := "  "
		if index == selected && active {
			marker = "› "
		}
		if index == selected && active {
			option = wizardSelectedStyle.Render(option)
		}
		fmt.Fprintf(builder, "%s%s\n", marker, wizardOptionStyle.Render(option))
	}
}

func writeReviewLine(builder *strings.Builder, label, value string) {
	fmt.Fprintf(builder, "  %s %s\n", labelStyle.Render(label+":"), value)
}

func minReplicaOptionIndex(value int) int {
	if value == 1 {
		return 1
	}
	return 0
}

func maxReplicaOptionIndex(value int) int {
	switch value {
	case 2:
		return 1
	case 5:
		return 2
	default:
		return 0
	}
}

func validateCreateRequest(request platform.CreateApplicationRequest) error {
	if err := validateApplicationName(request.Name); err != nil {
		return err
	}
	if err := validateOwningTeam(request.OwningTeam); err != nil {
		return err
	}
	if request.ComputeSize != "small" && request.ComputeSize != "medium" {
		return errors.New("compute size must be small or medium")
	}
	if request.ContainerPort < 1 || request.ContainerPort > 65535 {
		return errors.New("container port must be between 1 and 65535")
	}
	if request.Ingress != "external" && request.Ingress != "internal" {
		return errors.New("ingress must be external or internal")
	}
	if request.MinReplicas != 0 && request.MinReplicas != 1 {
		return errors.New("minimum replicas must be 0 or 1")
	}
	if request.MaxReplicas != 1 && request.MaxReplicas != 2 && request.MaxReplicas != 5 {
		return errors.New("maximum replicas must be 1, 2, or 5")
	}
	if request.MinReplicas > request.MaxReplicas {
		return errors.New("minimum replicas cannot exceed maximum replicas")
	}
	return nil
}

func validateApplicationName(value string) error {
	if !applicationNamePattern.MatchString(value) {
		return errors.New("application name must be a 3-63 character lowercase slug starting with a letter")
	}
	return nil
}

func validateOwningTeam(value string) error {
	if !teamIdentifierPattern.MatchString(value) {
		return errors.New("owning team must be a 2-128 character lowercase identifier")
	}
	return nil
}
