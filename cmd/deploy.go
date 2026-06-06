package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// odycConfigFile is the per-project file that links a local game folder to the
// game it deploys to.
const odycConfigFile = "odyc.json"

func init() {
	rootCmd.AddCommand(deployCmd)
}

// odycConfig is the contents of odyc.json.
type odycConfig struct {
	GameID string `json:"gameId"`
	Slug   string `json:"slug,omitempty"`
}

// gameDoc is the subset of an Odyc game document the CLI cares about.
type gameDoc struct {
	ID   string `json:"$id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// gameResponse models the JSON returned by the game API endpoints.
type gameResponse struct {
	Game    gameDoc `json:"game"`
	Message string  `json:"message"`
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current folder's game code to Odyc",
	Long:  `Bundle the current folder's game code (utils/, scenes/ and index.js) into a single file and update the linked game with it. The game must already exist (run 'odyc-cli create' first), and you must be signed in (run 'odyc-cli login', which authorizes deploys for all your games).`,
	Run: func(cmd *cobra.Command, args []string) {
		code, err := bundleGame(".")
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Warn("No game found in the current folder (expected index.js or game.js). Run 'odyc-cli create' first, or cd into your game folder")
				return
			}
			log.Error("Failed to read game code: " + err.Error())
			return
		}

		conf, err := loadOdycConfig(odycConfigFile)
		if err != nil {
			log.Error("Failed to read " + odycConfigFile + ": " + err.Error())
			return
		}
		if conf == nil || conf.GameID == "" {
			log.Warn("No game linked in this folder. Run 'odyc-cli create' to create a game first")
			return
		}

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

		// Update the game's code.
		_, game, err := callGameAPI(cfg, tokens, http.MethodPut, "/v1/games/"+conf.GameID+"/code", map[string]any{"code": code})
		if err != nil {
			var apiErr *apiError
			if errors.As(err, &apiErr) && apiErr.status == http.StatusForbidden {
				log.Warn(apiErr.message)
				log.Info("Authorize deploys by signing in again: odyc-cli login")
				return
			}
			reportAPIError("Failed to deploy game code", err)
			return
		}

		// Keep the recorded slug in sync (it may have been generated server-side).
		if game.Slug != "" && game.Slug != conf.Slug {
			conf.Slug = game.Slug
			if err := saveOdycConfig(odycConfigFile, conf); err != nil {
				log.Debug("Failed to update " + odycConfigFile + ": " + err.Error())
			}
		}

		slug := game.Slug
		if slug == "" {
			slug = conf.GameID
		}

		log.Logf(2, "Deployed successfully!")
		log.Logf(3, "Play your game at:")
		log.Info(odycAPIBase + "/g/" + slug)
	},
}

// bundleGame assembles the deployable, single-file game code from the project
// rooted at dir. Files are concatenated in the same order index.html loads them
// — every *.js in utils/, then every *.js in scenes/, then index.js last — so
// the uploaded bundle behaves exactly like the game does when played locally.
// Because all files share one scope, definitions in earlier files are available
// to later ones, and the entry point in index.js runs after everything is
// defined.
//
// For backwards compatibility with older single-file projects, a project with a
// game.js and no index.js is deployed as-is.
func bundleGame(dir string) (string, error) {
	indexPath := filepath.Join(dir, "index.js")
	if _, err := os.Stat(indexPath); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		// Legacy layout: a lone game.js.
		data, err := os.ReadFile(filepath.Join(dir, "game.js"))
		if err != nil {
			if os.IsNotExist(err) {
				return "", os.ErrNotExist
			}
			return "", err
		}
		return string(data), nil
	}

	var parts []string

	appendDir := func(sub string) error {
		entries, err := os.ReadDir(filepath.Join(dir, sub))
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".js") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)

		for _, name := range names {
			rel := filepath.ToSlash(filepath.Join(sub, name))
			data, err := os.ReadFile(filepath.Join(dir, sub, name))
			if err != nil {
				return err
			}
			parts = append(parts, fmt.Sprintf("// === %s ===\n%s", rel, string(data)))
		}
		return nil
	}

	if err := appendDir("utils"); err != nil {
		return "", err
	}
	if err := appendDir("scenes"); err != nil {
		return "", err
	}

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return "", err
	}
	parts = append(parts, "// === index.js ===\n"+string(indexData))

	return strings.Join(parts, "\n\n"), nil
}

// reportAPIError logs an API failure, mapping an expired/rejected session to a
// friendly hint.
func reportAPIError(prefix string, err error) {
	if err == errUnauthorized {
		log.Warn("Your session has expired. Run 'odyc-cli login' to sign in again")
		return
	}
	log.Error(prefix + ": " + err.Error())
}

// loadOdycConfig reads odyc.json from path. It returns (nil, nil) when the file
// does not exist.
func loadOdycConfig(path string) (*odycConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var conf odycConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// saveOdycConfig writes odyc.json to path.
func saveOdycConfig(path string, conf *odycConfig) error {
	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// callGameAPI performs an authenticated JSON request against the Odyc game API.
// It transparently refreshes the access token once on a 401 and retries.
// It returns the (possibly refreshed) tokens and the game from the response.
func callGameAPI(cfg *OIDCConfig, tokens *Tokens, method, endpoint string, payload any) (*Tokens, *gameDoc, error) {
	do := func(accessToken string) (int, []byte, error) {
		var body io.Reader
		if payload != nil {
			data, err := json.Marshal(payload)
			if err != nil {
				return 0, nil, err
			}
			body = bytes.NewReader(data)
		}

		req, err := http.NewRequest(method, odycAPIBase+endpoint, body)
		if err != nil {
			return 0, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, nil, err
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, respBody, nil
	}

	status, body, err := do(tokens.AccessToken)
	if err != nil {
		return tokens, nil, err
	}

	// Token rejected: refresh once and retry.
	if status == http.StatusUnauthorized && tokens.RefreshToken != "" {
		log.Debug("Access token rejected, attempting to refresh")
		refreshed, refreshErr := refreshTokens(cfg, tokens)
		if refreshErr == nil {
			if saveErr := saveTokens(refreshed); saveErr != nil {
				log.Debug("Failed to persist refreshed credentials: " + saveErr.Error())
			}
			tokens = refreshed
			status, body, err = do(tokens.AccessToken)
			if err != nil {
				return tokens, nil, err
			}
		}
	}

	if status == http.StatusUnauthorized {
		return tokens, nil, errUnauthorized
	}

	var gr gameResponse
	if jsonErr := json.Unmarshal(body, &gr); jsonErr != nil {
		return tokens, nil, errFromBody(status, body)
	}

	if status >= 200 && status < 300 {
		return tokens, &gr.Game, nil
	}

	if gr.Message != "" {
		return tokens, nil, &apiError{status: status, message: gr.Message}
	}
	return tokens, nil, errFromBody(status, body)
}

// apiError carries the HTTP status alongside the server's error message so
// callers can react to specific conditions (e.g. a 403 missing grant).
type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }
