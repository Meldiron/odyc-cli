package cmd

import (
	"bufio"
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// gameTemplate is the starter game.js bundled with the CLI. It mirrors the
// default "Create game" template used by the Odyc Play web editor.
//
//go:embed templates/game.js
var gameTemplate string

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create [folder]",
	Short: "Create a new Odyc game and scaffold a folder for it",
	Long:  `Create a new game on your Odyc account and scaffold a local folder containing a starter game.js (and odyc.json linking it to the game), ready to be edited and deployed.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folder := ""
		if len(args) > 0 {
			folder = strings.TrimSpace(args[0])
		}

		if folder == "" {
			folder = prompt("Folder name for your new game: ")
		}

		if folder == "" {
			log.Warn("A folder name is required")
			return
		}

		if _, err := os.Stat(folder); err == nil {
			log.Warn("A file or folder named '" + folder + "' already exists. Please choose another name")
			return
		}

		// Creating a game requires a signed-in account (games.create scope).
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

		log.Info("Creating a new game on your account...")
		name := filepath.Base(folder)
		_, game, err := callGameAPI(cfg, tokens, http.MethodPost, "/v1/games", map[string]any{"name": name})
		if err != nil {
			reportAPIError("Failed to create game", err)
			return
		}

		// Scaffold the local folder linked to the new game.
		if err := os.MkdirAll(folder, 0755); err != nil {
			log.Error("Created the game, but failed to create folder: " + err.Error())
			return
		}

		if err := os.WriteFile(filepath.Join(folder, "game.js"), []byte(gameTemplate), 0644); err != nil {
			log.Error("Failed to write game.js: " + err.Error())
			return
		}

		conf := &odycConfig{GameID: game.ID, Slug: game.Slug}
		if err := saveOdycConfig(filepath.Join(folder, odycConfigFile), conf); err != nil {
			log.Error("Failed to write " + odycConfigFile + ": " + err.Error())
			return
		}

		log.Logf(2, "Created a new Odyc game '%s' in '%s'", game.Name, folder)
		log.Logf(3, "What's next?")
		log.Infof("  cd %s", folder)
		log.Infof("  odyc-cli login --game-id=\"%s\"   # authorize deploys for this game", game.ID)
		log.Info("  odyc-cli deploy")
	},
}

// prompt reads a single line of input from stdin after printing label.
func prompt(label string) string {
	log.Logf(3, "%s", label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return strings.TrimSpace(line)
	}
	return strings.TrimSpace(line)
}
