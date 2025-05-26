// SPDX-License-Identifier: Apache-2.0
package config

import (
	"os"
	"path/filepath"
	"phase4/internal/app/errors"
	"phase4/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) func() {
	t.Helper()

	originalWd, err := os.Getwd()
	require.NoError(t, err, "Setup failed: could not get working directory")

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir), "Setup failed: could not change to temp dir")

	return func() {
		require.NoError(t, os.Chdir(originalWd), "Failed to change directory back to original")
	}
}

func TestLoadConfig_ReadFileError(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	filePath := filepath.Join(".", "config.yaml")
	testutil.CreateTempConfigFile(t, ".", "config.yaml", `debug: true`)
	require.NoError(t, os.Chmod(filePath, 0000), "Failed to make config file unreadable")
	defer func() {
		assert.NoError(t, os.Chmod(filePath, 0644), "Failed to restore config file permissions")
	}()

	_, err := Load()
	assert.Error(t, err, "Expected an error when reading unreadable config file")
	assert.ErrorIs(t, err, os.ErrPermission, "Expected a permission error")
}

func TestLoadConfig_InvalidYamlSyntax(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `debug: true: {invalid syntax`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	_, err := Load()

	assert.Error(t, err, "Expected an error for invalid YAML syntax")
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `debug: false`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	t.Setenv("ENV_DEBUG", "true")

	expected := getDefaultConfig()
	expected.Debug = true

	cfg, err := Load()

	require.NoError(t, err, "Load should succeed")
	require.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, *expected, *cfg, "Config mismatch")
}

func TestLoad_ValidFile(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `
debug: true
input:
  channels: 1
`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	expected := getDefaultConfig()
	expected.Debug = true
	expected.Input.Channels = 1

	cfg, err := Load()

	require.NoError(t, err, "LoadConfig should succeed")
	require.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, *expected, *cfg, "Config mismatch")
}

func TestLoadConfig_FilePreference(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	testutil.CreateTempConfigFile(t, ".", "config.yaml", `debug: true`)
	testutil.CreateTempConfigFile(t, ".", "config/config.yaml", `debug: false`)

	expected := getDefaultConfig()
	expected.Debug = true

	cfg, err := Load()

	require.NoError(t, err, "LoadConfig should succeed")
	require.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, *expected, *cfg, "Config mismatch")
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		assert.NoError(t, os.Chdir(originalWd), "Failed to change back to original directory")
	}()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir), "Failed to change to temporary directory")

	cfg, err := Load()

	assert.Nil(t, cfg, "Config should be nil when file is not found")
	assert.Error(t, err, "Error should not be nil when file is not found")

	var fatalErr *errors.FatalError
	assert.ErrorAs(t, err, &fatalErr, "Error should be of type *app.FatalError")

	if fatalErr != nil {
		assert.Equal(t, "file not found", fatalErr.Message, "FatalError message mismatch")
		expectedErrMsg := "config.yaml was not found in the current directory or any candidate subdirectory"
		assert.EqualError(t, fatalErr.Err, expectedErrMsg, "Underlying error message mismatch")
	}
}

func TestLoadConfig_ValidationError(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get CWD")
	defer func() { assert.NoError(t, os.Chdir(originalWd)) }()
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir), "Failed to change to temp dir")

	invalidConfigContent := `
transport:
  udp_enabled: true
  udp_send_address: "invalid-address"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(invalidConfigContent), 0644), "Failed to write invalid config file")

	cfg, loadErr := Load()
	assert.Nil(t, cfg, "Config should be nil when validation fails")
	assert.Error(t, loadErr, "Expected an error for validation failure")

	if assert.Error(t, loadErr) {
		assert.Contains(t, loadErr.Error(), "UDPSendAddress", "Error message should mention the invalid field 'UDPSendAddress'")
		assert.Contains(t, loadErr.Error(), "hostname_port", "Error message should mention the failed tag 'hostname_port'")
		var fatalErr *errors.FatalError
		assert.ErrorAs(t, loadErr, &fatalErr, "Error should be wrapped in app.FatalError")
		if fatalErr != nil {
			assert.Equal(t, "config YAML invalid", fatalErr.Message)
		}
	}
}
