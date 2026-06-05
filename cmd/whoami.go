package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently signed-in account",
	Long:  `Fetch and display information about the currently signed-in account from the OAuth userinfo endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		tokens, err := loadTokens()
		if err != nil {
			log.Error("Failed to read stored credentials: " + err.Error())
			return
		}

		if tokens == nil || tokens.AccessToken == "" {
			log.Warn("You are not logged in. Run 'odyc-cli login' first")
			return
		}

		cfg, err := fetchOIDCConfig()
		if err != nil {
			log.Error("Could not contact authorization server: " + err.Error())
			return
		}

		info, err := fetchUserinfo(cfg, tokens.AccessToken)

		// If the token is expired/rejected, transparently try to refresh once.
		if err == errUnauthorized && tokens.RefreshToken != "" {
			log.Debug("Access token rejected, attempting to refresh")
			refreshed, refreshErr := refreshTokens(cfg, tokens)
			if refreshErr != nil {
				log.Warn("Your session has expired. Run 'odyc-cli login' to sign in again")
				return
			}

			if saveErr := saveTokens(refreshed); saveErr != nil {
				log.Debug("Failed to persist refreshed credentials: " + saveErr.Error())
			}

			info, err = fetchUserinfo(cfg, refreshed.AccessToken)
		}

		if err == errUnauthorized {
			log.Warn("Your session has expired. Run 'odyc-cli login' to sign in again")
			return
		}

		if err != nil {
			log.Error("Failed to fetch account info: " + err.Error())
			return
		}

		log.Logf(3, "You are signed in as:")

		// Highlight the most relevant identity fields first, then the rest.
		for _, key := range []string{"name", "email", "sub"} {
			if val, ok := info[key]; ok {
				log.Infof("%s: %v", key, val)
			}
		}

		keys := make([]string, 0, len(info))
		for key := range info {
			switch key {
			case "name", "email", "sub":
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			log.Infof("%s: %v", key, info[key])
		}
	},
}

// fetchUserinfo calls the OIDC userinfo endpoint with the given access token.
func fetchUserinfo(cfg *OIDCConfig, accessToken string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errFromBody(resp.StatusCode, body)
	}

	var info map[string]any
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, errFromBody(resp.StatusCode, body)
	}

	return info, nil
}
