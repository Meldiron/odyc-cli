package cmd

import (
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out and remove stored credentials",
	Long:  `Revoke the current tokens at the authorization server (best effort) and delete the locally stored credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		tokens, err := loadTokens()
		if err != nil {
			log.Error("Failed to read stored credentials: " + err.Error())
			return
		}

		if tokens == nil || tokens.AccessToken == "" {
			log.Warn("You are not logged in")
			return
		}

		cfg, _ := fetchOIDCConfig()
		if err := signOut(cfg, tokens); err != nil {
			log.Error("Failed to remove stored credentials: " + err.Error())
			return
		}

		log.Logf(2, "Signed out successfully")
	},
}

// signOut revokes the given tokens at the authorization server (best effort)
// and removes the locally stored credentials. cfg may be nil, in which case
// revocation is skipped and only local cleanup happens.
func signOut(cfg *OIDCConfig, tokens *Tokens) error {
	if cfg != nil && cfg.RevocationEndpoint != "" {
		if tokens.RefreshToken != "" {
			revokeToken(cfg, tokens.RefreshToken, "refresh_token")
		}
		revokeToken(cfg, tokens.AccessToken, "access_token")
	}

	return clearTokens()
}

// revokeToken makes a best-effort call to the revocation endpoint (RFC 7009).
func revokeToken(cfg *OIDCConfig, token, hint string) {
	form := url.Values{}
	form.Set("token", token)
	form.Set("token_type_hint", hint)
	form.Set("client_id", oauthClientID)

	resp, err := http.PostForm(cfg.RevocationEndpoint, form)
	if err != nil {
		log.Debug("Token revocation request failed: " + err.Error())
		return
	}
	resp.Body.Close()
}
