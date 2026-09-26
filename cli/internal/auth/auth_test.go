package auth

import (
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
	"spx/internal/config"
)

func TestLogoutRemovesKeyringToken(t *testing.T) {
	store := newMemoryTokenStore()
	_ = store.Set(keyringService, "tenant:client", "token")
	if err := newManagerWithStore(testConfig(), store).Logout(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(keyringService, "tenant:client"); err != keyring.ErrNotFound {
		t.Fatalf("expected keyring token to be removed, error %v", err)
	}
}

func TestLogoutIsIdempotentWhenTokenIsMissing(t *testing.T) {
	err := newManagerWithStore(testConfig(), newMemoryTokenStore()).Logout()
	if err != nil {
		t.Fatal(err)
	}
}

func testConfig() config.Config {
	return config.Config{TenantID: "tenant", ClientID: "client", APIScope: "scope"}
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
