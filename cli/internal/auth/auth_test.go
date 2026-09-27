package auth

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
	"spx/internal/config"
)

func TestLogoutRemovesKeyringToken(t *testing.T) {
	store := newMemoryTokenStore()
	_ = store.Set(keyringService, "tenant:client", "token")
	removed, err := newManagerWithStore(testConfig(), store).Logout()
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected logout to remove a token")
	}
	if _, err := store.Get(keyringService, "tenant:client"); err != keyring.ErrNotFound {
		t.Fatalf("expected keyring token to be removed, error %v", err)
	}
}

func TestLogoutIsIdempotentWhenTokenIsMissing(t *testing.T) {
	removed, err := newManagerWithStore(testConfig(), newMemoryTokenStore()).Logout()
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Fatal("expected missing token to report already logged out")
	}
}

func TestLoginReportsBrowserFailureAndWaitsForCallback(t *testing.T) {
	manager := &Manager{
		config: testConfigWithRedirectURI(),
		store:  newMemoryTokenStore(),
		browser: BrowserLauncher{Start: func(string, ...string) error {
			return context.DeadlineExceeded
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var output bytes.Buffer

	if _, err := manager.Login(ctx, &output); err == nil {
		t.Fatal("expected canceled login to fail")
	}
	result := output.String()
	for _, expected := range []string{"SPX Login", "Starting browser authentication...", "Open this URL to sign in:", "Could not open the browser automatically:", "Waiting for the authentication callback...", "Login failed:"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in login output: %q", expected, result)
		}
	}
}

func TestLoginReportsCachedToken(t *testing.T) {
	store := newMemoryTokenStore()
	_ = store.Set(keyringService, "tenant:client", `{"access_token":"cached","token_type":"Bearer"}`)
	var output bytes.Buffer

	if _, err := newManagerWithStore(testConfig(), store).Login(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Already logged in.") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func testConfig() config.Config {
	return config.Config{TenantID: "tenant", ClientID: "client", APIScope: "scope"}
}

func testConfigWithRedirectURI() config.Config {
	config := testConfig()
	config.RedirectURI = "http://127.0.0.1:0/callback"
	return config
}

type memoryTokenStore struct {
	values map[string]string
}

func newMemoryTokenStore() *memoryTokenStore {
	return &memoryTokenStore{values: make(map[string]string)}
}

func (s *memoryTokenStore) Get(service, user string) (string, error) {
	value, ok := s.values[s.key(service, user)]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return value, nil
}

func (s *memoryTokenStore) Set(service, user, password string) error {
	s.values[s.key(service, user)] = password
	return nil
}

func (s *memoryTokenStore) Delete(service, user string) error {
	key := s.key(service, user)
	if _, ok := s.values[key]; !ok {
		return keyring.ErrNotFound
	}
	delete(s.values, key)
	return nil
}

func (s *memoryTokenStore) key(service, user string) string {
	return strings.Join([]string{service, user}, "\x00")
}

var _ tokenStore = (*memoryTokenStore)(nil)
