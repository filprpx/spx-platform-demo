package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadsUserConfigOutsideProject(t *testing.T) {
	clearConfigEnvironment(t)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	writeConfigFile(t, filepath.Join(configHome, "spx", "config.env"), "PLATFORM_API_URL=http://user-config.test\nPLATFORM_TENANT_ID=user-tenant\nPLATFORM_CLI_CLIENT_ID=user-client\nPLATFORM_API_SCOPE=api://user/scope\nPLATFORM_REDIRECT_URI=http://localhost:8765/callback\n")
	changeDirectory(t, t.TempDir())

	config, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.APIURL != "http://user-config.test" || config.TenantID != "user-tenant" {
		t.Fatalf("unexpected user config: %+v", config)
	}
}

func TestConfigPrecedenceIsEnvironmentThenProjectThenUser(t *testing.T) {
	clearConfigEnvironment(t)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	writeConfigFile(t, filepath.Join(configHome, "spx", "config.env"), "PLATFORM_TENANT_ID=user-tenant\nPLATFORM_CLI_CLIENT_ID=user-client\nPLATFORM_API_SCOPE=user-scope\n")
	project := t.TempDir()
	writeConfigFile(t, filepath.Join(project, ".env"), "PLATFORM_TENANT_ID=project-tenant\nPLATFORM_CLI_CLIENT_ID=project-client\nPLATFORM_API_SCOPE=project-scope\n")
	child := filepath.Join(project, "cli")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	changeDirectory(t, child)
	t.Setenv("PLATFORM_CLI_CLIENT_ID", "environment-client")

	config, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.TenantID != "project-tenant" || config.ClientID != "environment-client" || config.APIScope != "project-scope" {
		t.Fatalf("unexpected precedence result: %+v", config)
	}
}

func TestConfigReportsMissingRequiredValues(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	changeDirectory(t, t.TempDir())

	if _, err := FromEnv(); err == nil {
		t.Fatal("expected missing configuration error")
	}
}

func clearConfigEnvironment(t *testing.T) {
	for _, key := range keys {
		value, present := os.LookupEnv(key)
		_ = os.Unsetenv(key)
		t.Cleanup(func() {
			if present {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func writeConfigFile(t *testing.T, filename, contents string) {
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func changeDirectory(t *testing.T, directory string) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
}
