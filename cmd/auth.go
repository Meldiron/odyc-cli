package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Common OAuth error sentinels used by the auth commands.
var (
	errExpired      = errors.New("the sign-in request expired before it was confirmed")
	errDenied       = errors.New("the sign-in request was denied")
	errUnauthorized = errors.New("the access token was rejected")
)

// errOAuth builds an error from an OAuth error code and optional description.
func errOAuth(code, description string) error {
	if description != "" {
		return fmt.Errorf("%s: %s", code, description)
	}
	return fmt.Errorf("%s", code)
}

// errFromBody builds an error when a response body could not be parsed as JSON.
func errFromBody(status int, body []byte) error {
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 200 {
		snippet = snippet[:200]
	}
	if snippet == "" {
		return fmt.Errorf("unexpected response (status %d)", status)
	}
	return fmt.Errorf("unexpected response (status %d): %s", status, snippet)
}

// OAuth2.1 / OIDC configuration for the Odyc Play authorization server.
const (
	oauthIssuer   = "https://fra.cloud.appwrite.io/v1/oauth2/odyc-play"
	oauthClientID = "6a231e92000c0ef54658"
	oauthScope    = "openid profile email games.create"
)

// authorizationDetails builds the RFC 9396 Rich Authorization Request (RAR) the
// CLI asks for at sign-in: permission to write code to games. When gameID is
// set, the grant is narrowed to that specific game — the Odyc code-update
// endpoint authorizes per game by matching this detail's identifier (and type
// and actions) against the access token.
func authorizationDetails(gameID string) string {
	detail := map[string]any{
		"type":    "game",
		"actions": []string{"code.write"},
	}
	if gameID != "" {
		detail["identifier"] = gameID
	}

	data, err := json.Marshal([]map[string]any{detail})
	if err != nil {
		return `[{"type":"game","actions":["code.write"]}]`
	}
	return string(data)
}

// odycAPIBase is the base URL of the Odyc Play app, which hosts the public,
// OAuth2-protected API used by the `create` and `deploy` commands, as well as
// the playable game URLs.
const odycAPIBase = "https://odyc.appwrite.network"

// OIDCConfig holds the subset of the OpenID Connect discovery document we use.
type OIDCConfig struct {
	Issuer                      string `json:"issuer"`
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	UserinfoEndpoint            string `json:"userinfo_endpoint"`
	RevocationEndpoint          string `json:"revocation_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

// fetchOIDCConfig loads the discovery document. The device authorization
// endpoint is not advertised by the server, so we derive it from the issuer.
func fetchOIDCConfig() (*OIDCConfig, error) {
	wellKnown := strings.TrimRight(oauthIssuer, "/") + "/.well-known/openid-configuration"

	resp, err := http.Get(wellKnown)
	if err != nil {
		return nil, fmt.Errorf("failed to reach authorization server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authorization server returned status %d", resp.StatusCode)
	}

	var cfg OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse discovery document: %w", err)
	}

	if cfg.DeviceAuthorizationEndpoint == "" {
		cfg.DeviceAuthorizationEndpoint = strings.TrimRight(oauthIssuer, "/") + "/device_authorization"
	}

	return &cfg, nil
}

// Tokens is the credential bundle persisted on disk after a successful login.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	// ExpiresAt is the unix timestamp at which the access token expires.
	ExpiresAt int64 `json:"expires_at,omitempty"`
}

// Expired reports whether the access token is expired (with a small leeway).
func (t *Tokens) Expired() bool {
	if t.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() >= t.ExpiresAt-30
}

// tokenResponse models the JSON returned by the token endpoint.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	ExpiresIn        int64  `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (r *tokenResponse) toTokens() *Tokens {
	t := &Tokens{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		IDToken:      r.IDToken,
		TokenType:    r.TokenType,
		Scope:        r.Scope,
	}
	if r.ExpiresIn > 0 {
		t.ExpiresAt = time.Now().Unix() + r.ExpiresIn
	}
	return t
}

// authFilePath returns the path to the persisted credentials file.
func authFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "odyc-cli", "auth.json"), nil
}

// saveTokens writes credentials to disk with owner-only permissions.
func saveTokens(t *Tokens) error {
	path, err := authFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// loadTokens reads credentials from disk. It returns (nil, nil) when no
// credentials are stored yet.
func loadTokens() (*Tokens, error) {
	path, err := authFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var t Tokens
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

// clearTokens removes the credentials file. Missing file is not an error.
func clearTokens() error {
	path, err := authFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// refreshTokens exchanges a refresh token for a fresh access token.
func refreshTokens(cfg *OIDCConfig, t *Tokens) (*Tokens, error) {
	if t.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", t.RefreshToken)
	form.Set("client_id", oauthClientID)

	resp, err := http.PostForm(cfg.TokenEndpoint, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tr.Error != "" {
		return nil, fmt.Errorf("%s: %s", tr.Error, tr.ErrorDescription)
	}

	refreshed := tr.toTokens()
	// Some servers omit the refresh token on refresh; keep the existing one.
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = t.RefreshToken
	}

	return refreshed, nil
}

// openBrowser best-effort opens the given URL in the user's default browser.
func openBrowser(target string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{target}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", target}
	default: // linux, bsd, ...
		cmd = "xdg-open"
		args = []string{target}
	}

	return exec.Command(cmd, args...).Start()
}
