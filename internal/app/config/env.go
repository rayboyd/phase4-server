// SPDX-License-Identifier: Apache-2.0
package config

import (
	"log"
	"os"
	"strconv"

	// Autoload will read the .env file automatically when the package is imported.
	_ "github.com/joho/godotenv/autoload"
)

func applyEnvOverides(cfg *Config) {
	if val, exists := os.LookupEnv("ENV_DEBUG"); exists {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			cfg.Debug = boolVal
			log.Printf("Config ➜ Override, %s set to %v", "cfg.Debug", cfg.Debug)
		}
	}
}
