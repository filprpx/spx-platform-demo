package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var keys = []string{
	"PLATFORM_API_URL",
	"PLATFORM_TENANT_ID",
	"PLATFORM_CLI_CLIENT_ID",
	"PLATFORM_API_SCOPE",
	"PLATFORM_REDIRECT_URI",
}

// FromEnv applies the user config, nearest project .env, and explicit process
// variables in increasing precedence order.
func FromEnv() (Config, error) {
	values := make(map[string]string)
	mergeEnvFile(values, userConfigFile())
	if projectFile := findProjectEnvFile(); projectFile != "" {
		mergeEnvFile(values, projectFile)
	}
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}

	config := Config{
		APIURL:      values["PLATFORM_API_URL"],
		TenantID:    values["PLATFORM_TENANT_ID"],
		ClientID:    values["PLATFORM_CLI_CLIENT_ID"],
		APIScope:    values["PLATFORM_API_SCOPE"],
		RedirectURI: values["PLATFORM_REDIRECT_URI"],
	}
	if config.RedirectURI == "" {
		config.RedirectURI = "http://localhost:8765/callback"
	}
	if config.APIURL == "" {
		config.APIURL = "http://localhost:8000"
	}
	missing := make([]string, 0)
	if config.TenantID == "" {
		missing = append(missing, "PLATFORM_TENANT_ID")
	}
	if config.ClientID == "" {
		missing = append(missing, "PLATFORM_CLI_CLIENT_ID")
	}
	if config.APIScope == "" {
		missing = append(missing, "PLATFORM_API_SCOPE")
	}
	if len(missing) > 0 {
		return Config{}, configError(missing)
	}
	return config, nil
}

type Config struct {
	APIURL      string
	TenantID    string
	ClientID    string
	APIScope    string
	RedirectURI string
}

func userConfigFile() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "spx", "config.env")
}

func findProjectEnvFile() string {
	directory, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(directory, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return ""
		}
		directory = parent
	}
}

func mergeEnvFile(values map[string]string, filename string) {
	if filename == "" {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || !isConfigKey(key) {
			continue
		}
		values[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
	}
}

func isConfigKey(key string) bool {
	key = strings.TrimSpace(key)
	for _, allowed := range keys {
		if key == allowed {
			return true
		}
	}
	return false
}

func unquote(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func configError(missing []string) error {
	return fmt.Errorf("missing configuration: %s", strings.Join(missing, ", "))
}
