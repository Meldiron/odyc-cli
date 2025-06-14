package main

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func EnsureBinary(t *testing.T) {
	_, err := os.Stat("odyc")
	assert.Nil(t, err, "Binary file missing. Run 'go build -o odyc .'")
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func ExecuteCommand(t *testing.T, command string) CommandResult {
	cmd := exec.Command("sh", "-c", command)

	stdout, err := cmd.StdoutPipe()
	assert.Nil(t, err, "Failed to create CLI stdout pipe")

	stderr, err := cmd.StderrPipe()
	assert.Nil(t, err, "Failed to create CLI stderr pipe")

	err = cmd.Start()
	assert.Nil(t, err, "Failed to start command")

	stdoutBytes, err := io.ReadAll(stdout)
	assert.Nil(t, err, "Failed to read CLI stdout")

	stderrBytes, err := io.ReadAll(stderr)
	assert.Nil(t, err, "Failed to read CLI stderr")

	err = cmd.Wait()
	assert.Nil(t, err, "Failed to wait for command to finish")

	stdErr := string(stderrBytes)
	stdOut := string(stdoutBytes)

	exitCode := 0
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.ExitCode()
	}

	return CommandResult{
		ExitCode: exitCode,
		Stdout:   stdOut,
		Stderr:   stdErr,
	}
}
