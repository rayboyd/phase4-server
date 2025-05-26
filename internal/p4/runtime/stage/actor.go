// SPDX-License-Identifier: Apache-2.0
package stage

import (
	"context"
	"log"
)

func NewBaseActor(id string, capacity int, processor func(ctx context.Context, msg Message)) *BaseActor {
	if capacity <= 0 {
		capacity = 100
	}

	return &BaseActor{
		id:        id,
		mailbox:   make(chan Message, capacity),
		processor: processor,
	}
}

func (a *BaseActor) ID() string {
	return a.id
}

func (a *BaseActor) Send(msg Message) error {
	a.mu.RLock()

	if a.stopping || !a.started {
		a.mu.RUnlock()
		return ErrActorClosed
	}
	a.mu.RUnlock()

	select {
	case a.mailbox <- msg:
		return nil
	default:
		// Optional: Re-check state after failing to send, in case it changed.
		// This adds robustness against race conditions between RUnlock and select.
		a.mu.RLock()
		stoppedOrNotStarted := a.stopping || !a.started
		a.mu.RUnlock()
		if stoppedOrNotStarted {
			// Test notes:
			// This is the re-check path designed to catch the race condition where
			// the actor stops just as Send is trying to return ErrMailboxFull.
			// Hitting this reliably in a test is difficult due to timing, so not
			// covering it isn't necessarily a problem as long as the race detection
			// tools don't find any issues.
			return ErrActorClosed
		}
		return ErrMailboxFull
	}
}

func (a *BaseActor) SendNonBlocking(msg Message) error {
	a.mu.RLock()
	if a.stopping || !a.started {
		a.mu.RUnlock()
		return ErrActorClosed
	}
	a.mu.RUnlock()

	select {
	case a.mailbox <- msg:
		return nil
	default:
		return ErrMailboxFull // Don't block, just drop
	}
}

func (a *BaseActor) Start(ctx context.Context) error {
	a.mu.Lock()

	if a.stopping {
		a.mu.Unlock()
		return ErrActorClosed
	}

	if a.started {
		a.mu.Unlock()
		return nil // Already started, treat as no-op.
	}

	a.started = true
	a.mu.Unlock()

	a.wg.Add(1)
	go a.processLoop(ctx)

	return nil
}

func (a *BaseActor) Stop() error {
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return nil
	}

	a.stopping = true
	close(a.mailbox) // Signal processLoop to exit.
	a.mu.Unlock()

	// Waits for all in-flight messages to be processed.
	a.wg.Wait()

	return nil
}

func (a *BaseActor) processLoop(ctx context.Context) {
	defer a.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Actor[%s]: Context done, stopping", a.id)
			return

		case msg, ok := <-a.mailbox:
			if !ok {
				log.Printf("Actor[%s]: Mailbox closed, exiting process loop", a.id)
				return
			}
			if a.processor != nil {
				a.processor(ctx, msg)
			}
		}
	}
}
