// SPDX-License-Identifier: Apache-2.0
package main

import (
	"phase4/internal/app/config"
	"phase4/internal/app/errors"
	"phase4/internal/p4"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		errors.HandleFatalAndExit(err)
	}

	engine := p4.NewEngine(cfg)
	defer func() {
		if err := engine.Close(); err != nil {
			errors.HandleFatalAndExit(err)
		}
	}()

	if err := engine.Initialize(); err != nil {
		errors.HandleFatalAndExit(err)
	}
}
