// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"log"
	"os"
	"phase4/internal/app/errors"
	"time"

	"gopkg.in/yaml.v2"
)

// Load will attempt to load the configuration from a list of candidate
// files. If no file is found, it will return an error. The configuration will
// be parsed and validated. If any errors occur during loading or validation,
// they will be returned. The function will apply any environment variables to
// the configuration, taking precedence over the file values.
func Load() (*Config, error) {
	cfg := getDefaultConfig()

	var filePath = ""
	candidates := []string{
		"config.yaml",
		"config/config.yaml",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			filePath = candidate
			log.Printf("Config ➜ Candidate file %s found", filePath)
			break
		}
	}
	if filePath == "" {
		return nil, &errors.FatalError{
			Message: "file not found",
			Err:     fmt.Errorf("config.yaml was not found in the current directory or any candidate subdirectory"),
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnvOverides(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, &errors.FatalError{
			Message: "config YAML invalid",
			Err:     err,
		}
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	validate := GetValidator()
	return validate.Struct(cfg)
}

func getDefaultConfig() *Config {
	return &Config{
		Debug: false,
		Input: InputConfig{
			Device:     -1,
			Channels:   2,
			SampleRate: 44100,
			BufferSize: 512,
			LowLatency: false,
		},
		Transport: TransportConfig{
			UDPEnabled:       false,
			UDPSendAddress:   "127.0.0.1:8888",
			UDPSendInterval:  33 * time.Millisecond,
			WebSocketEnabled: false,
			WebSocketAddress: "127.0.0.1:8889",
			WebSocketPath:    "/ws",
		},
		DSP: DSPConfig{
			Enabled:   false,
			FFTWindow: "Hann",
		},
	}
}
