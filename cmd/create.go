package cmd

import (
	"bufio"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// templateRoot is the path, inside templateFS, to the starter project scaffold.
const templateRoot = "templates/game"

// templateFS holds the starter game project bundled with the CLI: index.html so
// the game is instantly playable in a browser, an index.js entry point, and
// scenes/ and utils/ folders showcasing a small, modular game. The "all:"
// prefix ensures dotfiles such as .gitignore are embedded too.
//
//go:embed all:templates/game
var templateFS embed.FS

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create [folder]",
	Short: "Create a new Odyc game and scaffold a folder for it",
	Long:  `Create a new game on your Odyc account and scaffold a local folder containing a starter project (index.html, index.js, scenes/ and utils/, plus an odyc.json linking it to the game), ready to be played, edited and deployed.`,
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

		if err := scaffoldTemplate(folder); err != nil {
			log.Error("Failed to scaffold the game folder: " + err.Error())
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
		log.Info("  open index.html   # play your game locally")
		log.Info("  odyc-cli deploy   # publish your changes")
	},
}

// scaffoldTemplate writes the embedded starter project into dest, recreating the
// template's folder structure (index.html, index.js, scenes/, utils/, …).
func scaffoldTemplate(dest string) error {
	return fs.WalkDir(templateFS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
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
