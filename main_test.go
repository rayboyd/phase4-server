package main

import (
	"bytes"
	"os"
	"os/exec"
	"phase4/internal/testutil"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_MAIN_FOR_TEST") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestBootstrap_ConfigNotFound(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestMain$")
	cmd.Env = append(os.Environ(), "RUN_MAIN_FOR_TEST=1")

	tempDir := t.TempDir()
	cmd.Dir = tempDir

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected process to exit with error, but got: %v\nStderr: %s", err, stderrBuf.String())
	}
	if e.Success() {
		t.Fatalf("Expected process to exit with non-zero status, but it exited successfully.\nStderr: %s", stderrBuf.String())
	}

	output := stderrBuf.String()
	expectedError := "file not found"
	if !strings.Contains(output, expectedError) {
		t.Errorf("Expected stderr to contain %q, but got:\n%s", expectedError, output)
	}
}

func TestBootstrap_InvalidConfig(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestMain$")
	cmd.Env = append(os.Environ(), "RUN_MAIN_FOR_TEST=1")

	tempDir := t.TempDir()
	cmd.Dir = tempDir
	testutil.CreateTempConfigFile(t, tempDir, "config.yaml", "input:\n  channels: 0")

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected process to exit with error, but got: %v\nStderr: %s", err, stderrBuf.String())
	}
	if e.Success() {
		t.Fatalf("Expected process to exit with non-zero status, but it exited successfully.\nStderr: %s", stderrBuf.String())
	}

	output := stderrBuf.String()
	expectedError := "config invalid"
	if !strings.Contains(output, expectedError) {
		t.Errorf("Expected stderr to contain %q, but got:\n%s", expectedError, output)
	}
}

func TestBootstrap_Success(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestMain$")
	cmd.Env = append(os.Environ(), "RUN_MAIN_FOR_TEST=1")

	tempDir := t.TempDir()
	cmd.Dir = tempDir
	testutil.CreateTempConfigFile(t, tempDir, "config.yaml", "debug: true")

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Run the command and expect it to succeed (exit 0) because bootstrap passes
	// and the current run() function also returns nil.
	err := cmd.Run()

	if err != nil {
		// Any error here is unexpected, including ExitError.
		t.Fatalf("Expected process to exit successfully (0), but got error: %v\nStderr: %s", err, stderrBuf.String())
	}

	// Assert that stderr does NOT contain known error messages.
	output := stderrBuf.String()
	if strings.Contains(output, "file not found") || strings.Contains(output, "config invalid") {
		t.Errorf("Expected clean stderr on success, but got:\n%s", output)
	}
}
