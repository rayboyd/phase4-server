package config

import (
	"os"
	"strconv"
)

func applyEnvOverides(c *Config) {
	if val, exists := os.LookupEnv("ENV_DEBUG"); exists {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			c.Debug = boolVal
		}
	}
}
