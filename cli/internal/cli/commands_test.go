package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"spx/internal/auth"
	"spx/internal/config"
	"spx/internal/platform"
)

func TestLogoutCommandIsRegisteredAndClearsToken(t *testing.T) {
	keyring.MockInit()
	if err := keyring.Set("spx", "tenant:client", "token"); err != nil {
		t.Fatal(err)
	}
	setCLIConfigEnvironment(t)

	var output bytes.Buffer
	command := newTestCommand()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"logout"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Logged out successfully.\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
	if _, err := keyring.Get("spx", "tenant:client"); err != keyring.ErrNotFound {
		t.Fatalf("expected keyring token to be removed, error: %v", err)
	}

	command = newTestCommand()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"logout"})
	if err := command.Execute(); err != nil {
		t.Fatalf("logout should be idempotent: %v", err)
	}
	if !strings.HasSuffix(output.String(), "Already logged out.\n") {
		t.Fatalf("expected already-logged-out output, got %q", output.String())
	}
}

func TestHelpListsLogout(t *testing.T) {
	var output bytes.Buffer
	command := newTestCommand()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "logout") {
		t.Fatalf("help did not list logout: %s", output.String())
	}
}

func TestCreateCommandRejectsLegacyPositionalName(t *testing.T) {
	command := newCreateCommand(Dependencies{})
	command.SetArgs([]string{"payments-api"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "positional arguments are not supported") {
		t.Fatalf("expected positional argument error, got %v", err)
	}
}

func TestCreateCommandRequiresTerminalForAutomaticWizard(t *testing.T) {
	command := newCreateCommand(Dependencies{})
	command.SetArgs(nil)

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires a terminal") {
		t.Fatalf("expected terminal guidance, got %v", err)
	}
}

func TestCreateCommandDoesNotRegisterLegacyFlags(t *testing.T) {
	command := newCreateCommand(Dependencies{})
	command.SetArgs([]string{"--type", "api"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --type") {
		t.Fatalf("expected removed flag error, got %v", err)
	}
}

func setCLIConfigEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("PLATFORM_TENANT_ID", "tenant")
	t.Setenv("PLATFORM_CLI_CLIENT_ID", "client")
	t.Setenv("PLATFORM_API_SCOPE", "scope")
}

func newTestCommand() *cobra.Command {
	return NewCommand(Dependencies{
		LoadConfig: config.FromEnv,
		NewAuth:    auth.NewManager,
		NewAPI: func(url string, source oauth2.TokenSource) *platform.Client {
			return platform.NewClient(url, source)
		},
	})
}
