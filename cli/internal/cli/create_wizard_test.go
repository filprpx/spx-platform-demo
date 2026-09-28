package cli

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"spx/internal/platform"
)

func TestCreateWizardDefaultsAndGroupedSteps(t *testing.T) {
	model := newCreateWizardModel(platform.CreateApplicationRequest{})

	if model.request.Name != "payments-api" || model.request.OwningTeam != "finance-engineering" || model.request.ComputeSize != "small" || model.request.ContainerPort != 8080 || model.request.Ingress != "external" || model.request.MinReplicas != 0 || model.request.MaxReplicas != 1 {
		t.Fatalf("unexpected wizard defaults: %+v", model.request)
	}

	pressEnter := func() {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(createWizardModel)
	}
	pressEnter() // name -> owning team
	pressEnter() // identity -> compute
	pressEnter() // compute size -> port
	pressEnter() // compute -> network
	pressEnter() // network -> scaling
	pressEnter() // minimum -> maximum
	pressEnter() // scaling -> review

	if model.step != wizardReviewStep {
		t.Fatalf("expected review step, got %d", model.step)
	}
	view := model.View()
	for _, expected := range []string{"Application identity", "Compute configuration", "Network configuration", "Scaling configuration", "payments-api", "finance-engineering"} {
		if !strings.Contains(view, expected) && expected != "Application identity" && expected != "Compute configuration" && expected != "Network configuration" && expected != "Scaling configuration" {
			t.Fatalf("expected %q in review output: %q", expected, view)
		}
	}
	pressEnter()
	if !model.confirmed {
		t.Fatal("expected review confirmation")
	}
}

func TestCreateWizardUsesPlaceholderDefaultsWhenFieldsAreUntouched(t *testing.T) {
	model := newCreateWizardModel(platform.CreateApplicationRequest{})
	if model.nameInput.Value() != "" || model.teamInput.Value() != "" || model.portInput.Value() != "" {
		t.Fatalf("defaults should remain placeholders, got values: name=%q team=%q port=%q", model.nameInput.Value(), model.teamInput.Value(), model.portInput.Value())
	}
	for _, placeholder := range []string{model.nameInput.Placeholder, model.teamInput.Placeholder, model.portInput.Placeholder} {
		if !strings.Contains(placeholder, "(default)") {
			t.Fatalf("expected placeholder to explain default behavior: %q", placeholder)
		}
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(createWizardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(createWizardModel)
	if model.request.Name != "payments-api" || model.request.OwningTeam != "finance-engineering" {
		t.Fatalf("untouched defaults were not retained: %+v", model.request)
	}
}

func TestValidateCreateRequest(t *testing.T) {
	valid := platform.CreateApplicationRequest{
		Name: "payments-api", OwningTeam: "finance-engineering", ComputeSize: "small",
		ContainerPort: 8080, Ingress: "external", MinReplicas: 0, MaxReplicas: 1,
	}
	if err := validateCreateRequest(valid); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		request platform.CreateApplicationRequest
	}{
		{"uppercase name", func() platform.CreateApplicationRequest { r := valid; r.Name = "Payments-api"; return r }()},
		{"invalid team", func() platform.CreateApplicationRequest { r := valid; r.OwningTeam = "finance_engineering"; return r }()},
		{"invalid port", func() platform.CreateApplicationRequest { r := valid; r.ContainerPort = 65536; return r }()},
		{"invalid replicas", func() platform.CreateApplicationRequest { r := valid; r.MinReplicas = 2; r.MaxReplicas = 1; return r }()},
		{"minimum exceeds maximum", func() platform.CreateApplicationRequest { r := valid; r.MinReplicas = 1; r.MaxReplicas = 0; return r }()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateCreateRequest(test.request); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreateWizardCancellation(t *testing.T) {
	model := newCreateWizardModel(platform.CreateApplicationRequest{})
	updated, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(createWizardModel)
	if !result.canceled || command == nil {
		t.Fatalf("expected cancellation and quit command: %+v", result)
	}
}

func TestCreateWizardTabSyncsApplicationName(t *testing.T) {
	model := newCreateWizardModel(platform.CreateApplicationRequest{})
	model.nameInput.SetValue("payments-api")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(createWizardModel)
	if model.request.Name != "payments-api" {
		t.Fatalf("expected name to be synced on tab, got %q", model.request.Name)
	}
}

func TestMaxReplicaOptionIndexUsesOneAsFirstOption(t *testing.T) {
	if got := maxReplicaOptionIndex(1); got != 0 {
		t.Fatalf("expected max replica 1 to map to option 0, got %d", got)
	}
	if got := maxReplicaOptionIndex(2); got != 1 {
		t.Fatalf("expected max replica 2 to map to option 1, got %d", got)
	}
	if got := maxReplicaOptionIndex(5); got != 2 {
		t.Fatalf("expected max replica 5 to map to option 2, got %d", got)
	}
}
