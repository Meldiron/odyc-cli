package cmd

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

// deviceAuthResponse models the device authorization endpoint response (RFC 8628).
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int64  `json:"interval"`
	Error                   string `json:"error"`
	ErrorDescription        string `json:"error_description"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to Odyc using your browser",
	Long:  `Sign in using the OAuth 2.1 device authorization flow. A code is shown which you confirm in your browser to authenticate this device. Signing in also authorizes code deploys for all of your games.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := fetchOIDCConfig()
		if err != nil {
			log.Error("Could not contact authorization server: " + err.Error())
			return
		}

		// Already signed in? Sign out the previous session automatically so the
		// user can log in (or re-authorize) without a manual logout step.
		if tokens, _ := loadTokens(); tokens != nil && tokens.AccessToken != "" {
			log.Info("Signing out the previous session...")
			if err := signOut(cfg, tokens); err != nil {
				log.Debug("Failed to clear previous credentials: " + err.Error())
			}
		}

		device, err := requestDeviceCode(cfg)
		if err != nil {
			log.Error("Failed to start sign-in: " + err.Error())
			return
		}

		verificationURL := device.VerificationURIComplete
		if verificationURL == "" {
			verificationURL = device.VerificationURI
		}

		log.Logf(3, "To sign in, visit the following URL in your browser:")
		log.Info(verificationURL)
		log.Logf(3, "And confirm this code: %s", device.UserCode)

		// Start polling immediately, in the background: the developer may open
		// the URL manually (or copy it elsewhere) before pressing ENTER, so we
		// must be ready to receive the token regardless.
		type pollResult struct {
			tokens *Tokens
			err    error
		}
		tokenCh := make(chan pollResult, 1)
		go func() {
			tokens, err := pollForToken(cfg, device)
			tokenCh <- pollResult{tokens, err}
		}()

		// Wait for ENTER to open the browser, but don't block sign-in on it.
		enterCh := make(chan struct{}, 1)
		go func() {
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')
			enterCh <- struct{}{}
		}()

		log.Logf(3, "Press ENTER to open the URL in your browser, or open it yourself")
		log.Info("Waiting for you to authorize...")

		for {
			select {
			case <-enterCh:
				if err := openBrowser(verificationURL); err != nil {
					log.Debug("Could not open browser automatically: " + err.Error())
				}
				// Keep waiting for authorization.
			case res := <-tokenCh:
				if res.err != nil {
					log.Error("Sign-in failed: " + res.err.Error())
					return
				}
				if err := saveTokens(res.tokens); err != nil {
					log.Error("Signed in, but failed to save credentials: " + err.Error())
					return
				}
				log.Logf(2, "Signed in successfully! Run 'odyc-cli create' to create your first game")
				return
			}
		}
	},
}

// requestDeviceCode kicks off the device authorization flow.
func requestDeviceCode(cfg *OIDCConfig) (*deviceAuthResponse, error) {
	form := url.Values{}
	form.Set("client_id", oauthClientID)
	form.Set("scope", oauthScope)

	// Always request the Rich Authorization Request so a single sign-in
	// authorizes code deploys for all of the user's games.
	form.Set("authorization_details", authorizationDetails())

	resp, err := http.PostForm(cfg.DeviceAuthorizationEndpoint, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var device deviceAuthResponse
	if err := json.Unmarshal(body, &device); err != nil {
		return nil, errFromBody(resp.StatusCode, body)
	}

	if device.Error != "" {
		return nil, errOAuth(device.Error, device.ErrorDescription)
	}

	if device.Interval <= 0 {
		device.Interval = 5
	}

	return &device, nil
}

// pollForToken polls the token endpoint until the user authorizes, the request
// is denied, or the device code expires.
func pollForToken(cfg *OIDCConfig, device *deviceAuthResponse) (*Tokens, error) {
	interval := time.Duration(device.Interval) * time.Second

	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)
	if device.ExpiresIn <= 0 {
		deadline = time.Now().Add(5 * time.Minute)
	}

	for {
		if time.Now().After(deadline) {
			return nil, errExpired
		}

		time.Sleep(interval)

		form := url.Values{}
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		form.Set("device_code", device.DeviceCode)
		form.Set("client_id", oauthClientID)

		resp, err := http.PostForm(cfg.TokenEndpoint, form)
		if err != nil {
			return nil, err
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			return nil, errFromBody(resp.StatusCode, body)
		}

		// Success: access token issued.
		if tr.Error == "" && tr.AccessToken != "" {
			return tr.toTokens(), nil
		}

		switch tr.Error {
		case "authorization_pending":
			// Keep waiting.
			continue
		case "slow_down":
			// Server asks us to back off; increase the interval.
			interval += 5 * time.Second
			continue
		case "access_denied":
			return nil, errDenied
		case "expired_token":
			return nil, errExpired
		default:
			return nil, errOAuth(tr.Error, tr.ErrorDescription)
		}
	}
}
