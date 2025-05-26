// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SignalHandler struct {
	signals chan os.Signal
	done    chan struct{}
	cancel  context.CancelFunc
	once    sync.Once // Add this to prevent double-close
}

func NewSignalHandler(cancel context.CancelFunc) *SignalHandler {
	sh := &SignalHandler{
		signals: make(chan os.Signal, 1),
		done:    make(chan struct{}),
		cancel:  cancel,
	}

	signal.Notify(sh.signals, os.Interrupt, syscall.SIGTERM)
	go sh.handle()

	return sh
}

func (sh *SignalHandler) handle() {
	select {
	case sig := <-sh.signals:
		log.Printf("Received signal: %v, initiating shutdown...", sig)
		sh.cancel()
	case <-sh.done:
		return
	}
}

func (sh *SignalHandler) Stop() {
	signal.Stop(sh.signals)
	sh.once.Do(func() {
		close(sh.done)
	})
}
