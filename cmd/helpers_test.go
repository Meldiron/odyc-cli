package cmd

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func EnsureFile(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.Nil(t, err, "File missing: "+path)
}

type CommandResult struct {
	ExitCode int
	Output   string
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

	exitCode := 0

	if err != nil {
		handled := false
		if exitErr, ok := err.(*exec.ExitError); ok {
			handled = true
			exitCode = exitErr.ExitCode()
		}

		if !handled {
			assert.Nil(t, err, "Unexpected CLI error: "+err.Error())
		}
	}

	stdErr := string(stderrBytes)
	stdOut := string(stdoutBytes)

	return CommandResult{
		ExitCode: exitCode,
		Output:   stdOut + stdErr,
	}
}
