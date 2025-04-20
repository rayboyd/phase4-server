package config

import (
	"os"
	"strconv"
)

func applyEnvOverides(cfg *Config) {
	if val, exists := os.LookupEnv("ENV_DEBUG"); exists {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			cfg.Debug = boolVal
		}
	}
}
