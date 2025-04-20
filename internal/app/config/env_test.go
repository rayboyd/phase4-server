package config

import (
	"testing"
)

func TestApplyEnvOverrides(t *testing.T) {
	testCases := []struct {
		name         string
		envValue     *string // A pointer to distinguish unset (nil) from empty ("")
		initialDebug bool
		expectDebug  bool
	}{
		{"Not set", nil, false, false},
		{"Not set, initially true", nil, true, true},
		{"Set true", stringPtr("true"), false, true},
		{"Set false", stringPtr("false"), true, false},
		{"Set 1", stringPtr("1"), false, true},
		{"Set 0", stringPtr("0"), true, false},
		{"Set invalid", stringPtr("invalid"), false, false},
		{"Set empty", stringPtr(""), true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != nil {
				t.Setenv("ENV_DEBUG", *tc.envValue)
			}

			cfg := Config{Debug: tc.initialDebug}
			applyEnvOverides(&cfg)

			// Assertions:
			// - Is Debug set to the expected value?
			if cfg.Debug != tc.expectDebug {
				t.Errorf("Initial: %v, Env: %v -> Expected Debug: %v, Got: %v",
					tc.initialDebug, tc.envValue, tc.expectDebug, cfg.Debug)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
