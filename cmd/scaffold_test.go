package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaffoldTemplateWritesProject(t *testing.T) {
	dest := t.TempDir()

	err := scaffoldTemplate(dest)
	assert.Nil(t, err, "scaffoldTemplate should succeed")

	// The starter project must be playable and modular out of the box.
	for _, rel := range []string{
		"index.html",
		"index.js",
		filepath.Join("scenes", "title.js"),
		filepath.Join("scenes", "world.js"),
		filepath.Join("utils", "sprites.js"),
		"README.md",
		".gitignore",
	} {
		EnsureFile(t, filepath.Join(dest, rel))
	}

	html, err := os.ReadFile(filepath.Join(dest, "index.html"))
	assert.Nil(t, err)
	// index.html must pull odyc from a CDN so the game is instantly playable.
	assert.Contains(t, string(html), "odyc@")
	assert.Contains(t, string(html), "index.global.js")
}

func TestBundleGameConcatenatesInOrder(t *testing.T) {
	dest := t.TempDir()
	assert.Nil(t, scaffoldTemplate(dest))

	bundle, err := bundleGame(dest)
	assert.Nil(t, err, "bundleGame should succeed")

	// utils precede scenes precede index.js so definitions are available before
	// the entry point runs.
	utilsAt := strings.Index(bundle, "utils/sprites.js")
	scenesAt := strings.Index(bundle, "scenes/title.js")
	indexAt := strings.Index(bundle, "// === index.js ===")
	assert.NotEqual(t, -1, utilsAt)
	assert.NotEqual(t, -1, scenesAt)
	assert.NotEqual(t, -1, indexAt)
	assert.Less(t, utilsAt, scenesAt, "utils should come before scenes")
	assert.Less(t, scenesAt, indexAt, "scenes should come before index.js")

	// The bundle should actually contain the game code.
	assert.Contains(t, bundle, "openTitleScene")
	assert.Contains(t, bundle, "openWorldScene")
	assert.Contains(t, bundle, "PLAYER_SPRITE")
}

func TestBundleGameLegacyGameJS(t *testing.T) {
	dest := t.TempDir()
	assert.Nil(t, os.WriteFile(filepath.Join(dest, "game.js"), []byte("createGame({})"), 0644))

	bundle, err := bundleGame(dest)
	assert.Nil(t, err, "bundleGame should fall back to game.js")
	assert.Equal(t, "createGame({})", bundle)
}

func TestBundleGameMissing(t *testing.T) {
	dest := t.TempDir()

	_, err := bundleGame(dest)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
