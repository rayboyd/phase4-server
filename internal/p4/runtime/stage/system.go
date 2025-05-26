// SPDX-License-Identifier: Apache-2.0
package stage

import (
	"context"
	"fmt"
	"log"
	"maps"
)

func NewSystem() *System {
	ctx, cancel := context.WithCancel(context.Background())

	return &System{
		actors: make(map[string]Actor),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *System) Register(actor Actor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := actor.ID()
	if _, exists := s.actors[id]; exists {
		return fmt.Errorf("actor with ID %s already registered", id)
	}

	s.actors[id] = actor
	log.Printf("Engine ➜ Stage ➜ Actor registered: %s", id)

	return nil
}

func (s *System) Get(id string) (Actor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actor, exists := s.actors[id]

	return actor, exists
}

func (s *System) Send(actorID string, msg Message) error {
	s.mu.RLock()
	actor, exists := s.actors[actorID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("actor with ID %s not found", actorID)
	}

	return actor.Send(msg)
}

func (s *System) SendNonBlocking(actorID string, msg Message) error {
	s.mu.RLock()
	actor, exists := s.actors[actorID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("actor with ID %s not found", actorID)
	}

	if baseActor, ok := actor.(*BaseActor); ok {
		return baseActor.SendNonBlocking(msg)
	}

	// Fallback to blocking send for other actor types
	return actor.Send(msg)
}

func (s *System) StartAll() map[string]error {
	s.mu.RLock()
	actors := make(map[string]Actor, len(s.actors))
	maps.Copy(actors, s.actors)
	s.mu.RUnlock()

	errors := make(map[string]error)
	for id, actor := range actors {
		if err := actor.Start(s.ctx); err != nil {
			errors[id] = err
			log.Printf("Stage ➜ Failed to start actor %s: %v", id, err)
		} else {
			log.Printf("Stage ➜ Started actor: %s", id)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func (s *System) StopAll() map[string]error {
	// Cancel context first to signal all actors to begin shutdown.
	s.cancel()

	s.mu.RLock()
	actors := make(map[string]Actor, len(s.actors))
	maps.Copy(actors, s.actors)
	s.mu.RUnlock()

	errors := make(map[string]error)
	for id, actor := range actors {
		if err := actor.Stop(); err != nil {
			errors[id] = err
			log.Printf("Stage ➜ Failed to stop actor %s: %v", id, err)
		} else {
			log.Printf("Stage ➜ Stopped actor: %s", id)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func (s *System) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear actors map to prevent further operations
	s.actors = make(map[string]Actor)
	return nil
}
