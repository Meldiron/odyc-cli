package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpritesCommand(t *testing.T) {
	// Cleanup
	err := os.RemoveAll("../tests/resources/sprites.js")
	assert.Nil(t, err, "Failed to do cleanup before tests start")

	// Preparation
	EnsureFile(t, "../odyc")
	var response CommandResult

	// Basic usage
	response = ExecuteCommand(t, "../odyc sprites")
	assert.Equal(t, response.ExitCode, 1)
	assert.Contains(t, response.Output, "Usage:")
	assert.Contains(t, response.Output, `required flag(s) "assets", "output" not set`)

	// --help param
	response = ExecuteCommand(t, "../odyc sprites --help")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, "Usage:")

	// Some required params missing
	response = ExecuteCommand(t, "../odyc sprites -a ../tests/resources/")
	assert.Equal(t, response.ExitCode, 1)
	assert.Contains(t, response.Output, `required flag(s) "output" not set`)

	response = ExecuteCommand(t, "../odyc sprites -o ../tests/resources/sprites.js")
	assert.Equal(t, response.ExitCode, 1)
	assert.Contains(t, response.Output, `required flag(s) "assets" not set`)

	// Successful run
	response = ExecuteCommand(t, "../odyc sprites -a ../tests/resources/ -o ../tests/resources/sprites.js")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, `9 colors found across all sprites`)
	assert.Contains(t, response.Output, `2 sprites found across all PNG files`)
	assert.Contains(t, response.Output, `Sprites configuration generated successfully`)

	// TODO: Validate sprites.js

	// --force param
	response = ExecuteCommand(t, "../odyc sprites -a ../tests/resources/ -o ../tests/resources/sprites.js")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, `Output file already exists`)

	response = ExecuteCommand(t, "../odyc sprites -a ../tests/resources/ -o ../tests/resources/sprites.js -f")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, `Sprites configuration generated successfully`)

	// TODO: Failure tests (folders missing, pngs missing, and so on)
}
