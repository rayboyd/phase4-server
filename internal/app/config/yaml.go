package config

import (
	"os"
	"phase4/internal/app"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug bool        `yaml:"debug"`
	Input InputConfig `yaml:"input"`
	DSP   DSPConfig   `yaml:"dsp"`
}

type DSPConfig struct {
	Enabled   bool   `yaml:"enabled"`
	FFTWindow string `yaml:"fft_window"`
}

type InputConfig struct {
	Device     int     `yaml:"device"`
	Channels   int     `yaml:"channels"`
	SampleRate float64 `yaml:"sample_rate"`
	BufferSize int     `yaml:"buffer_size"`
	LowLatency bool    `yaml:"low_latency"`
}

// LoadConfig will attempt to load the configuration from a list of candidate
// files. If no file is found, it will return an error. The configuration will
// be parsed and validated. If any errors occur during loading or validation,
// they will be returned. The function will apply any environment variables to
// the configuration, taking precedence over the file values.
func LoadConfig() (*Config, error) {
	cfg := Config{
		Debug: false,
		Input: InputConfig{
			Device:     -1,
			Channels:   2,
			SampleRate: 44100,
			BufferSize: 512,
			LowLatency: false,
		},
	}

	var fp string = ""
	candidates := []string{
		"config.yaml",
		"config/config.yaml",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			fp = candidate
			break
		}
	}
	if fp == "" {
		return nil, app.ErrFileNotFound
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyEnvOverides(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	return nil
}
