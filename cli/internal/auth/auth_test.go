package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestLogoutRemovesCachedTokenAndPreservesConfiguration(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token.json")
	configFile := filepath.Join(t.TempDir(), "config.env")
	if err := os.WriteFile(tokenFile, []byte(`{"access_token":"token"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte("PLATFORM_TENANT_ID=tenant\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := newManagerWithStore(Config{TokenFile: tokenFile}, newMemoryTokenStore()).Logout(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tokenFile); !os.IsNotExist(err) {
		t.Fatalf("expected token to be removed, stat error: %v", err)
	}
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("expected configuration to remain: %v", err)
	}
}

func TestLogoutIsIdempotentWhenTokenIsMissing(t *testing.T) {
	err := newManagerWithStore(Config{TokenFile: filepath.Join(t.TempDir(), "missing-token.json")}, newMemoryTokenStore()).Logout()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogoutReportsTokenRemovalFailure(t *testing.T) {
	tokenDirectory := filepath.Join(t.TempDir(), "token.json")
	if err := os.Mkdir(tokenDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tokenDirectory, "contents"), []byte("token"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := newManagerWithStore(Config{TokenFile: tokenDirectory}, newMemoryTokenStore()).Logout(); err == nil {
		t.Fatal("expected token removal failure")
	}
}

func TestLegacyTokenIsMigratedToKeyringAndFileIsRemoved(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token.json")
	legacyToken := `{"access_token":"legacy-token","token_type":"Bearer"}`
	if err := os.WriteFile(tokenFile, []byte(legacyToken), 0o600); err != nil {
		t.Fatal(err)
	}
	store := newMemoryTokenStore()

	token, err := newManagerWithStore(Config{TenantID: "tenant", ClientID: "client", TokenFile: tokenFile}, store).readToken()
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "legacy-token" {
		t.Fatalf("unexpected token: %+v", token)
	}
	if _, err := os.Stat(tokenFile); !os.IsNotExist(err) {
		t.Fatalf("expected legacy token file to be removed, stat error: %v", err)
	}
	if got, err := store.Get(keyringService, "tenant:client"); err != nil || got != legacyToken {
		t.Fatalf("expected migrated token, got %q, error %v", got, err)
	}
}

type memoryTokenStore struct {
	values map[string]string
}

func newMemoryTokenStore() *memoryTokenStore {
	return &memoryTokenStore{values: make(map[string]string)}
}

func (s *memoryTokenStore) Get(_, user string) (string, error) {
	value, ok := s.values[user]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return value, nil
}

func (s *memoryTokenStore) Set(_, user, password string) error {
	s.values[user] = password
	return nil
}

func (s *memoryTokenStore) Delete(_, user string) error {
	if _, ok := s.values[user]; !ok {
		return keyring.ErrNotFound
	}
	delete(s.values, user)
	return nil
}

var _ tokenStore = (*memoryTokenStore)(nil)
