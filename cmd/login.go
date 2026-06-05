package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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
	Long:  `Sign in using the OAuth 2.1 device authorization flow. A code is shown which you confirm in your browser to authenticate this device.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := fetchOIDCConfig()
		if err != nil {
			log.Error("Could not contact authorization server: " + err.Error())
			return
		}

		// Already signed in?
		if tokens, _ := loadTokens(); tokens != nil && tokens.AccessToken != "" && !tokens.Expired() {
			log.Warn("You are already logged in. Run 'odyc-cli logout' first to sign in as someone else")
			return
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

		if err := openBrowser(verificationURL); err != nil {
			log.Debug("Could not open browser automatically: " + err.Error())
		}

		log.Info("Waiting for you to authorize in the browser...")

		tokens, err := pollForToken(cfg, device)
		if err != nil {
			log.Error("Sign-in failed: " + err.Error())
			return
		}

		if err := saveTokens(tokens); err != nil {
			log.Error("Signed in, but failed to save credentials: " + err.Error())
			return
		}

		log.Logf(2, "Signed in successfully! Run 'odyc-cli whoami' to see your account")
	},
}

// requestDeviceCode kicks off the device authorization flow.
func requestDeviceCode(cfg *OIDCConfig) (*deviceAuthResponse, error) {
	form := url.Values{}
	form.Set("client_id", oauthClientID)
	form.Set("scope", oauthScope)

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
