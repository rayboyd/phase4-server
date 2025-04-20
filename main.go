package main

import (
	"fmt"
	"phase4/internal/app"
	"phase4/internal/app/config"
)

var appCfg *config.Config
var appErr error

func main() {
	if err := bootstrap(); err != nil {
		app.HandleFatalAndExit(err)
	}

	if err := run(); err != nil {
		app.HandleFatalAndExit(err)
	}
}

func bootstrap() error {
	// --- 1. Load the configuration
	appCfg, appErr = config.LoadConfig()
	if appErr != nil {
		return appErr
	}

	return nil
}

func run() error {
	fmt.Printf("%v\n", appCfg)

	return nil
}
