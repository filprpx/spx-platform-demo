package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var configKeys = []string{
	"PLATFORM_API_URL",
	"PLATFORM_TENANT_ID",
	"PLATFORM_CLI_CLIENT_ID",
	"PLATFORM_API_SCOPE",
	"PLATFORM_REDIRECT_URI",
	"PLATFORM_TOKEN_FILE",
}

// loadConfiguration applies user config, then the nearest project .env, then
// explicit process variables. This lets an installed CLI work outside the
// repository while still allowing a checkout-specific configuration override.
func loadConfiguration() map[string]string {
	values := make(map[string]string)
	mergeEnvFile(values, userConfigFile())
	if projectFile := findProjectEnvFile(); projectFile != "" {
		mergeEnvFile(values, projectFile)
	}
	for _, key := range configKeys {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}
	return values
}

func userConfigFile() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "spx-platform", "config.env")
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
	for _, allowed := range configKeys {
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

func configValue(values map[string]string, key string) string { return values[key] }

func configError(missing []string) error {
	return fmt.Errorf("missing configuration: %s", strings.Join(missing, ", "))
}
