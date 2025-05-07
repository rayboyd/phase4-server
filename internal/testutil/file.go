// SPDX-License-Identifier: Apache-2.0
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func CreateTempConfigFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	filePath := filepath.Join(dir, filename)

	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir for config '%s': %v", filePath, err)
	}

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file '%s': %v", filePath, err)
	}
}
