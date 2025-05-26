// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"log"
	"os"
	"phase4/internal/app/config"
	"phase4/internal/app/errors"
	"phase4/internal/p4"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		errors.HandleFatalAndExit(err)
	}

	engine := p4.NewEngine(cfg)
	lifecycle := p4.NewLifecycleManager(engine)

	// Initialize but don't start yet
	if err := engine.Initialize(); err != nil {
		errors.HandleFatalAndExit(err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	signalHandler := p4.NewSignalHandler(cancel)
	defer signalHandler.Stop()

	// Start the engine
	if err := lifecycle.Start(); err != nil {
		errors.HandleFatalAndExit(err)
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		lifecycle.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		log.Print("Shutdown completed successfully")
	case <-shutdownCtx.Done():
		log.Print("Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}
}
