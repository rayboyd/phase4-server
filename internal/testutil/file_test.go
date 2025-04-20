package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCreateTempConfigFile_Nominal ensures the helper runs without error
// under normal conditions. The actual file creation is implicitly tested
// by the tests that *use* this helper.
func TestCreateTempConfigFile_Nominal(t *testing.T) {
	tempDir := t.TempDir()
	filename := "test_nominal.txt"
	content := "nominal content"

	CreateTempConfigFile(t, tempDir, filename, content)

	expectedPath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Helper failed to create file '%s'", expectedPath)
	}
}
