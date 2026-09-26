package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

type Config struct {
	APIURL      string
	TenantID    string
	ClientID    string
	APIScope    string
	RedirectURI string
	TokenFile   string
}

type Manager struct {
	config Config
	store  tokenStore
}

func NewManager(config Config) *Manager {
	return &Manager{config: config, store: systemTokenStore{}}
}

func newManagerWithStore(config Config, store tokenStore) *Manager {
	return &Manager{config: config, store: store}
}

func ConfigFromEnv() (Config, error) {
	values := loadConfiguration()
	config := Config{
		APIURL:      configValue(values, "PLATFORM_API_URL"),
		TenantID:    configValue(values, "PLATFORM_TENANT_ID"),
		ClientID:    configValue(values, "PLATFORM_CLI_CLIENT_ID"),
		APIScope:    configValue(values, "PLATFORM_API_SCOPE"),
		RedirectURI: configValue(values, "PLATFORM_REDIRECT_URI"),
		TokenFile:   configValue(values, "PLATFORM_TOKEN_FILE"),
	}
	if config.RedirectURI == "" {
		config.RedirectURI = "http://localhost:8765/callback"
	}
	if config.APIURL == "" {
		config.APIURL = "http://localhost:8000"
	}
	if config.TokenFile == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return Config{}, fmt.Errorf("resolve user config directory: %w", err)
		}
		config.TokenFile = filepath.Join(configDir, "spx-platform", "token.json")
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

func (m *Manager) Login(ctx context.Context) (*oauth2.Token, error) {
	if cached, err := m.readToken(); err == nil && cached.Valid() {
		return cached, nil
	}
	verifier, err := randomString(32)
	if err != nil {
		return nil, err
	}
	state, err := randomString(24)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	parsed, err := url.Parse(m.config.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("parse redirect URI: %w", err)
	}
	server := &http.Server{}
	codeCh := make(chan string, 1)
	errorCh := make(chan error, 1)
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != parsed.Path {
			http.Error(w, "invalid callback path", http.StatusBadRequest)
			return
		}
		if callbackState := r.URL.Query().Get("state"); callbackState != state {
			http.Error(w, "invalid OAuth state", http.StatusBadRequest)
			errorCh <- errors.New("OAuth state mismatch")
			return
		}
		if callbackError := r.URL.Query().Get("error"); callbackError != "" {
			errorCh <- fmt.Errorf("Microsoft Entra login failed: %s", callbackError)
			_, _ = io.WriteString(w, "Login failed. You can close this window.")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errorCh <- errors.New("OAuth callback did not contain an authorization code")
			return
		}
		codeCh <- code
		_, _ = io.WriteString(w, "Login complete. You can close this window.")
	})
	listener, err := net.Listen("tcp", parsed.Host)
	if err != nil {
		return nil, fmt.Errorf("listen for OAuth callback: %w", err)
	}
	defer listener.Close()
	go func() { _ = server.Serve(listener) }()

	endpoint := oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/token",
	}
	oauthConfig := oauth2.Config{ClientID: m.config.ClientID, Endpoint: endpoint, RedirectURL: m.config.RedirectURI, Scopes: []string{m.config.APIScope, "openid", "profile", "offline_access"}}
	authURL := oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("code_challenge", challenge))
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Open this URL in a browser:\n%s\n", authURL)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errorCh:
		return nil, err
	case code := <-codeCh:
		token, err := oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
		if err != nil {
			return nil, fmt.Errorf("exchange authorization code: %w", err)
		}
		if err := m.writeToken(token); err != nil {
			return nil, err
		}
		return token, nil
	}
}

func (m *Manager) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := m.readToken()
	if err != nil {
		return nil, errors.New("not logged in; run platform login")
	}
	endpoint := oauth2.Endpoint{TokenURL: "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/token"}
	oauthConfig := oauth2.Config{ClientID: m.config.ClientID, Endpoint: endpoint}
	return oauth2.ReuseTokenSource(token, oauthConfig.TokenSource(ctx, token)), nil
}

func (m *Manager) Logout() error {
	var errs []error
	if err := m.store.Delete(keyringService, m.keyringUser()); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("remove keyring token: %w", err))
	}
	if err := os.Remove(m.config.TokenFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, fmt.Errorf("remove legacy token: %w", err))
	}
	return errors.Join(errs...)
}

func (m *Manager) readToken() (*oauth2.Token, error) {
	secret, err := m.store.Get(keyringService, m.keyringUser())
	if err == nil {
		return decodeToken([]byte(secret))
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("read token from OS keyring: %w", err)
	}

	legacy, err := os.ReadFile(m.config.TokenFile)
	if err != nil {
		return nil, err
	}
	token, err := decodeToken(legacy)
	if err != nil {
		return nil, err
	}
	if err := m.store.Set(keyringService, m.keyringUser(), string(legacy)); err != nil {
		return nil, fmt.Errorf("migrate token to OS keyring: %w", err)
	}
	if err := os.Remove(m.config.TokenFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove migrated token file: %w", err)
	}
	return token, nil
}

func (m *Manager) writeToken(token *oauth2.Token) error {
	b, err := json.Marshal(token)
	if err != nil {
		return err
	}
	if err := m.store.Set(keyringService, m.keyringUser(), string(b)); err != nil {
		return fmt.Errorf("save token to OS keyring: %w", err)
	}
	return nil
}

func (m *Manager) keyringUser() string {
	return m.config.TenantID + ":" + m.config.ClientID
}

func decodeToken(data []byte) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func randomString(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(target string) error {
	command := "xdg-open"
	if runtime.GOOS == "darwin" {
		command = "open"
	}
	if runtime.GOOS == "windows" {
		command = "rundll32"
	}
	args := []string{target}
	if runtime.GOOS == "windows" {
		args = []string{"url.dll,FileProtocolHandler", target}
	}
	return exec.Command(command, args...).Start()
}
