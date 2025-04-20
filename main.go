package main

import (
	"log"
	"phase4/internal/app"
	"phase4/internal/app/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		app.HandleFatalAndExit(err)
	}

	log.Printf("Configuration loaded successfully.")
	log.Printf("Debug enabled: %v\n", cfg.Debug)
}
