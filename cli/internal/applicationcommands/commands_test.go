package applicationcommands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestLogoutCommandIsRegisteredAndClearsToken(t *testing.T) {
	keyring.MockInit()
	tokenFile := filepath.Join(t.TempDir(), "token.json")
	if err := os.WriteFile(tokenFile, []byte(`{"access_token":"token"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	setCLIConfigEnvironment(t, tokenFile)

	var output bytes.Buffer
	command := NewRootCommand()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"logout"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Logged out.\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
	if _, err := os.Stat(tokenFile); !os.IsNotExist(err) {
		t.Fatalf("expected token to be removed, stat error: %v", err)
	}

	command = NewRootCommand()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"logout"})
	if err := command.Execute(); err != nil {
		t.Fatalf("logout should be idempotent: %v", err)
	}
}

func TestHelpListsLogout(t *testing.T) {
	var output bytes.Buffer
	command := NewRootCommand()
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

func setCLIConfigEnvironment(t *testing.T, tokenFile string) {
	t.Helper()
	t.Setenv("PLATFORM_TENANT_ID", "tenant")
	t.Setenv("PLATFORM_CLI_CLIENT_ID", "client")
	t.Setenv("PLATFORM_API_SCOPE", "scope")
	t.Setenv("PLATFORM_TOKEN_FILE", tokenFile)
}
