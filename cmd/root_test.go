package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	EnsureFile(t, "../odyc-cli")
	var response CommandResult

	response = ExecuteCommand(t, "../odyc-cli")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, "Welcome to Odyc.js CLI!")

	response = ExecuteCommand(t, "../odyc-cli --help")
	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Output, "Usage:")
	assert.Contains(t, response.Output, "Available Commands:")
}
