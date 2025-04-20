package config

import (
	"errors"
	"os"
	"path/filepath"
	"phase4/internal/app"
	"phase4/internal/testutil"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func setupTest(t *testing.T) func() {
	t.Helper()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Setup failed: could not get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Setup failed: could not change to temp dir '%s': %v", tempDir, err)
	}

	return func() {
		if err := os.Chdir(originalWd); err != nil {
			// Use t.Errorf or t.Fatalf - Fatalf might be better as cleanup failed
			t.Fatalf("Failed to change directory back to original '%s': %v", originalWd, err)
		}
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	_, err := LoadConfig()

	if !errors.Is(err, app.ErrFileNotFound) {
		t.Errorf("Expected error '%v', got '%v'", app.ErrFileNotFound, err)
	}
}

func TestLoadConfig_ReadFileError(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	filePath := filepath.Join(".", "config.yaml")
	testutil.CreateTempConfigFile(t, ".", "config.yaml", `debug: true`)
	if err := os.Chmod(filePath, 0000); err != nil { // Make unreadable
		t.Fatalf("Failed to make config file unreadable: %v", err)
	}
	defer func() {
		if err := os.Chmod(filePath, 0644); err != nil {
			t.Errorf("Failed to restore config file permissions: %v", err) // Errorf might be sufficient here
		}
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected an error when reading unreadable config file, but got nil")
	}

	if !errors.Is(err, os.ErrPermission) && !strings.Contains(err.Error(), "permission denied") { // Be a bit flexible
		t.Errorf("Expected a permission error, but got: %v", err)
	}

	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("Expected a permission error (os.ErrPermission), but got: %v (%T)", err, err)
	}

	if errors.Is(err, app.ErrFileNotFound) {
		t.Errorf("Expected a read error, but got ErrFileNotFound")
	}
	if errors.Is(err, app.ErrConfigInvalid) {
		t.Errorf("Expected a read error, but got ErrConfigInvalid")
	}
}

func TestLoadConfig_ValidationError(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `
input:
  channels: 0 # Invalid: must be > 0
`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	_, err := LoadConfig()

	if !errors.Is(err, app.ErrConfigInvalid) {
		t.Fatalf("Expected error to be '%v', got '%v'", app.ErrConfigInvalid, err)
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Errorf("Expected error to wrap '%T', but it did not (err: %v)", validationErrors, err)
	}
}

func TestLoadConfig_InvalidYamlSyntax(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `debug: true: {invalid syntax`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	_, err := LoadConfig()

	if err == nil {
		t.Fatal("Expected an error for invalid YAML syntax, but got nil")
	}
	// We don't assert a specific type here, just that *an* error occurred.
	// The underlying yaml library handles the specific error reporting.
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	yamlContent := `debug: false`
	testutil.CreateTempConfigFile(t, ".", "config.yaml", yamlContent)

	t.Setenv("ENV_DEBUG", "true")

	expected := getDefaultConfig()
	expected.Debug = true

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if !reflect.DeepEqual(*cfg, *expected) {
		t.Errorf("Config mismatch:\nExpected: %+v\nActual:   %+v", *expected, *cfg)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
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

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if !reflect.DeepEqual(*cfg, *expected) {
		t.Errorf("Config mismatch:\nExpected: %+v\nActual:   %+v", *expected, *cfg)
	}
}

func TestLoadConfig_FilePreference(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	testutil.CreateTempConfigFile(t, ".", "config.yaml", `debug: true`)         // Root file
	testutil.CreateTempConfigFile(t, ".", "config/config.yaml", `debug: false`) // Subdir file

	expected := getDefaultConfig()
	expected.Debug = true // Expect value from root file

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if !reflect.DeepEqual(*cfg, *expected) {
		t.Errorf("Config mismatch:\nExpected: %+v\nActual:   %+v", *expected, *cfg)
	}
}
