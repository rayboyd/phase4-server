// SPDX-License-Identifier: Apache-2.0
package stage

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrMailboxFull = errors.New("actor mailbox full")
	ErrActorClosed = errors.New("actor closed or stopping")
)

type Message interface {
	Type() string
}

type Actor interface {
	ID() string                      // ID returns the unique identifier for this actor.
	Send(msg Message) error          // Send delivers a message to this actor's mailbox.
	Start(ctx context.Context) error // Start begins the actor's processing loop.
	Stop() error                     // Stop gracefully shuts down the actor.
}

type TypedActor[T Message] interface {
	Actor
	SendTyped(msg T) error
}

type BaseActor struct {
	mailbox   chan Message
	processor func(ctx context.Context, msg Message)
	id        string
	wg        sync.WaitGroup
	mu        sync.RWMutex
	stopping  bool
	started   bool
}
