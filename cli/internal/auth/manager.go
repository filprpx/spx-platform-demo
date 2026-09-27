package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"spx/internal/config"
)

var (
	loginTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	loginInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	loginURLStyle   = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("39"))
	loginGoodStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	loginBadStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	loginSpinStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type Manager struct {
	config  config.Config
	store   tokenStore
	browser BrowserLauncher
}

func NewManager(config config.Config) *Manager {
	return &Manager{config: config, store: systemTokenStore{}, browser: NewBrowserLauncher()}
}

func newManagerWithStore(config config.Config, store tokenStore) *Manager {
	return &Manager{config: config, store: store, browser: NewBrowserLauncher()}
}

func (m *Manager) Login(ctx context.Context, out io.Writer) (*oauth2.Token, error) {
	if cached, err := m.readToken(); err == nil && cached.Valid() {
		fmt.Fprintln(out, loginGoodStyle.Render("✓ Already logged in."))
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

	callback, err := NewCallbackServer(m.config.RedirectURI, state)
	if err != nil {
		return nil, err
	}
	if err := callback.Start(); err != nil {
		return nil, err
	}
	defer callback.Close(context.Background())

	endpoint := oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/token",
	}
	oauthConfig := oauth2.Config{ClientID: m.config.ClientID, Endpoint: endpoint, RedirectURL: m.config.RedirectURI, Scopes: []string{m.config.APIScope, "openid", "profile", "offline_access"}}
	authURL := oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("code_challenge", challenge))
	fmt.Fprintln(out, loginTitleStyle.Render("SPX Login"))
	fmt.Fprintln(out, loginInfoStyle.Render("Starting browser authentication..."))
	fmt.Fprintln(out, "Open this URL to sign in:")
	fmt.Fprintln(out, loginURLStyle.Render(authURL))
	if err := m.browser.Open(authURL); err != nil {
		fmt.Fprintf(out, "%s Could not open the browser automatically: %v\n", loginBadStyle.Render("!"), err)
	} else {
		fmt.Fprintln(out, loginGoodStyle.Render("✓ Browser opened."))
	}
	fmt.Fprintf(out, "%s %s\n", loginSpinStyle.Render(spinner.MiniDot.Frames[0]), loginInfoStyle.Render("Waiting for the authentication callback..."))

	code, err := callback.Wait(ctx)
	if err != nil {
		fmt.Fprintf(out, "%s Login failed: %v\n", loginBadStyle.Render("✗"), err)
		return nil, fmt.Errorf("login failed while waiting for callback: %w", err)
	}
	token, err := oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		fmt.Fprintf(out, "%s Login failed: %v\n", loginBadStyle.Render("✗"), err)
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	if err := m.writeToken(token); err != nil {
		fmt.Fprintf(out, "%s Login failed: %v\n", loginBadStyle.Render("✗"), err)
		return nil, err
	}
	fmt.Fprintln(out, loginGoodStyle.Render("✓ Successfully logged in."))
	return token, nil
}

func (m *Manager) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := m.readToken()
	if err != nil {
		return nil, errors.New("not logged in; run spx login")
	}
	endpoint := oauth2.Endpoint{TokenURL: "https://login.microsoftonline.com/" + m.config.TenantID + "/oauth2/v2.0/token"}
	oauthConfig := oauth2.Config{ClientID: m.config.ClientID, Endpoint: endpoint}
	return oauth2.ReuseTokenSource(token, oauthConfig.TokenSource(ctx, token)), nil
}

func (m *Manager) Logout() (bool, error) {
	err := m.store.Delete(keyringService, m.keyringUser())
	if err == nil {
		return true, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("remove keyring token: %w", err)
}

func (m *Manager) readToken() (*oauth2.Token, error) {
	secret, err := m.store.Get(keyringService, m.keyringUser())
	if err == nil {
		return decodeToken([]byte(secret))
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("read token from OS keyring: %w", err)
	}
	return nil, err
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
