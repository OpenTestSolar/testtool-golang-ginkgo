package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	path, err := filepath.Abs(".")
	assert.NoError(t, err)
	_, _, err = RunCommand("ls", path, false, false)
	assert.NoError(t, err)
	_, _, err = RunCommand("ls", path, false, true)
	assert.NoError(t, err)
	_, _, err = RunCommand("ls", path, true, false)
	assert.NoError(t, err)
	_, _, err = RunCommand("ls", path, true, true)
	assert.NoError(t, err)
}

func TestRunCommandWithOutput(t *testing.T) {
	path, err := filepath.Abs(".")
	assert.NoError(t, err)
	_, _, err = RunCommandWithOutput("ls", path)
	assert.NoError(t, err)
}

func TestRunCommandWithEnvs(t *testing.T) {
	path, err := filepath.Abs(".")
	assert.NoError(t, err)
	_, _, _, err = RunCommandWithEnvs("ls", path, map[string]string{}, false, false)
	assert.NoError(t, err)
	_, _, _, err = RunCommandWithEnvs("ls", path, map[string]string{}, true, false)
	assert.NoError(t, err)
	_, _, _, err = RunCommandWithEnvs("ls", path, map[string]string{}, true, true)
	assert.NoError(t, err)
	_, _, _, err = RunCommandWithEnvs("ls", path, map[string]string{}, false, true)
	assert.NoError(t, err)
}
