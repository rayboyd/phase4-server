// SPDX-License-Identifier: Apache-2.0
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyEnvOverrides(t *testing.T) {
	testCases := []struct {
		envValue     *string
		name         string
		initialDebug bool
		expectDebug  bool
	}{
		{nil, "Not set", false, false},
		{nil, "Not set, initially true", true, true},
		{stringPtr("true"), "Set true", false, true},
		{stringPtr("false"), "Set false", true, false},
		{stringPtr("1"), "Set 1", false, true},
		{stringPtr("0"), "Set 0", true, false},
		{stringPtr("invalid"), "Set invalid", false, false},
		{stringPtr(""), "Set empty", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != nil {
				t.Setenv("ENV_DEBUG", *tc.envValue)
			}

			cfg := Config{Debug: tc.initialDebug}
			applyEnvOverides(&cfg)
			assert.Equal(t, tc.expectDebug, cfg.Debug, "Debug value mismatch")
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
