// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type LifecycleManager struct {
	engine *Engine
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	state  LifecycleState
}

type LifecycleState int

const (
	StateUninitialized LifecycleState = iota
	StateInitialized
	StateRunning
	StateShuttingDown
	StateClosed
)

func NewLifecycleManager(engine *Engine) *LifecycleManager {
	return &LifecycleManager{
		engine: engine,
		state:  StateInitialized, // Important: start in initialized state
	}
}

func (lm *LifecycleManager) Start() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.state != StateInitialized {
		return fmt.Errorf("invalid state for start: %v", lm.state)
	}

	lm.state = StateRunning
	lm.ctx, lm.cancel = context.WithCancel(context.Background())
	lm.done = make(chan struct{})

	go lm.run()
	return nil
}

func (lm *LifecycleManager) run() {
	defer close(lm.done)

	if err := lm.engine.Run(lm.ctx); err != nil {
		log.Printf("Engine run error: %v", err)
	}
}

func (lm *LifecycleManager) Shutdown() {
	lm.mu.Lock()
	if lm.state == StateShuttingDown || lm.state == StateClosed {
		lm.mu.Unlock()
		return
	}
	lm.state = StateShuttingDown
	lm.mu.Unlock()

	log.Print("Starting graceful shutdown...")

	// Cancel context to stop all operations
	if lm.cancel != nil {
		lm.cancel()
	}

	// Wait for run() to complete
	if lm.done != nil {
		<-lm.done
	}

	// Close engine resources
	if err := lm.engine.Close(); err != nil {
		log.Printf("Error during engine close: %v", err)
	}

	lm.mu.Lock()
	lm.state = StateClosed
	lm.mu.Unlock()
}
