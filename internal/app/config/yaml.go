package config

import (
	"fmt"
	"os"
	"phase4/internal/app"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug     bool            `yaml:"debug"`
	Input     InputConfig     `yaml:"input"     validate:"required"`
	Transport TransportConfig `yaml:"transport" validate:"required"`
	DSP       DSPConfig       `yaml:"dsp"       validate:"required"`
}

type InputConfig struct {
	Device     int     `yaml:"device"      validate:"required,gte=-1"`
	Channels   int     `yaml:"channels"    validate:"required,gt=0"`
	SampleRate float64 `yaml:"sample_rate" validate:"required,gt=0"`
	BufferSize int     `yaml:"buffer_size" validate:"required,gt=0"`
	LowLatency bool    `yaml:"low_latency"`
}

type TransportConfig struct {
	UDPEnabled      bool          `yaml:"udp_enabled"`
	UDPSendAddress  string        `yaml:"udp_send_address"  validate:"required_if=UDPEnabled true,hostname_port"`
	UDPSendInterval time.Duration `yaml:"udp_send_interval" validate:"required_if=UDPEnabled true,gt=0"`
}

type DSPConfig struct {
	Enabled   bool   `yaml:"enabled"`
	FFTWindow string `yaml:"fft_window" validate:"required_if=Enabled true,oneof='BartlettHann' 'Blackman' 'BlackmanNuttall' 'Hann' 'Hamming' 'Lanczos' 'Nuttall'"`
}

// LoadConfig will attempt to load the configuration from a list of candidate
// files. If no file is found, it will return an error. The configuration will
// be parsed and validated. If any errors occur during loading or validation,
// they will be returned. The function will apply any environment variables to
// the configuration, taking precedence over the file values.
func LoadConfig() (*Config, error) {
	cfg := Config{
		Debug: false,
		// Input: InputConfig{
		// 	Device:     -1,
		// 	Channels:   2,
		// 	SampleRate: 44100,
		// 	BufferSize: 512,
		// 	LowLatency: false,
		// },
		Transport: TransportConfig{
			UDPEnabled:      false,
			UDPSendAddress:  "127.0.0.1:8888",
			UDPSendInterval: 33 * time.Millisecond,
		},
		DSP: DSPConfig{
			Enabled:   false,
			FFTWindow: "Hann",
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
		return nil, fmt.Errorf("%d: %w", app.ErrConfigInvalid, err)
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(cfg)
}
