package config

import (
	"os"
	"phase4/internal/app"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug     bool            `yaml:"debug"`
	Input     InputConfig     `yaml:"input"`
	DSP       DSPConfig       `yaml:"dsp"`
	Transport TrandportConfig `yaml:"transport"`
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

type TrandportConfig struct {
	UDPEnabled      bool          `yaml:"udp_enabled"`
	UDPSendAddress  string        `yaml:"udp_send_address"`
	UDPSendInterval time.Duration `yaml:"udp_send_interval"`
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
		Transport: TrandportConfig{
			UDPEnabled:      false,
			UDPSendAddress:  "127.0.0.1:8888",
			UDPSendInterval: 33 * time.Millisecond,
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
