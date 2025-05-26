// SPDX-License-Identifier: Apache-2.0
package config

import "time"

type Config struct {
	DSP       DSPConfig       `yaml:"dsp"       validate:"required"`
	Transport TransportConfig `yaml:"transport" validate:"required"`
	Input     InputConfig     `yaml:"input"     validate:"required"`
	Debug     bool            `yaml:"debug"`
}

type InputConfig struct {
	Device           int     `yaml:"device"      validate:"gte=-1"`
	Channels         int     `yaml:"channels"    validate:"gt=0"`
	SampleRate       float64 `yaml:"sample_rate" validate:"gt=0"`
	BufferSize       int     `yaml:"buffer_size" validate:"gt=0"`
	LowLatency       bool    `yaml:"low_latency"`
	UseDefaultDevice bool    `yaml:"use_default"`
}

type TransportConfig struct {
	UDPSendAddress   string        `yaml:"udp_send_address"  validate:"required_if=UDPEnabled true,hostname_port"`
	WebSocketAddress string        `yaml:"websocket_address" validate:"required_if=WebSocketEnabled true,hostname_port"`
	WebSocketPath    string        `yaml:"websocket_path"    validate:"required_if=WebSocketEnabled true"`
	UDPSendInterval  time.Duration `yaml:"udp_send_interval" validate:"required_if=UDPEnabled true,gt=0"`
	UDPEnabled       bool          `yaml:"udp_enabled"`
	WebSocketEnabled bool          `yaml:"websocket_enabled"`
}

type DSPConfig struct {
	FFTWindow string `yaml:"fft_window" validate:"required_if=Enabled true,oneof='BartlettHann' 'Blackman' 'BlackmanNuttall' 'Hann' 'Hanning' 'Hamming' 'Lanczos' 'Nuttall'"`
	Enabled   bool   `yaml:"enabled"`
}
