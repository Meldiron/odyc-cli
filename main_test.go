package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainCommand(t *testing.T) {
	EnsureBinary(t)

	response := ExecuteCommand(t, "./odyc")

	assert.Equal(t, response.ExitCode, 0)
	assert.Contains(t, response.Stdout, "Welcome to Odyc.js CLI!")
}
