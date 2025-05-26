// SPDX-License-Identifier: Apache-2.0
package pipeline

import (
	"context"
	"fmt"
	"log"
	"phase4/internal/p4/runtime/stage"
)

func NewRouter(id string, capacity int, targetIDs []string, system *stage.System) (*RouterComponent, error) {
	if system == nil {
		return nil, fmt.Errorf("RouterComponent[%s] requires a non-nil system", id)
	}

	a := &RouterComponent{
		targetIDs: targetIDs,
		system:    system,
	}
	a.BaseActor = *stage.NewBaseActor(id, capacity, a.processMessage)

	return a, nil
}

func (a *RouterComponent) processMessage(ctx context.Context, msg stage.Message) {
	fftMsg, ok := msg.(*stage.FFTData)
	if !ok {
		log.Printf("Router[%s] ➜ Warning ➜ Received unexpected message type: %T", a.ID(), msg)
		// If this unexpected message happens to be a pooled type, it might leak.
		// The LogComponent is the designated pool handler, so we don't Put here.
		// Consider if more robust handling is needed for unexpected pooled types.
		return
	}

	// Sends the FFTData message to all target clients.
	for _, targetID := range a.targetIDs {
		if err := a.system.Send(targetID, fftMsg); err != nil {
			log.Printf("Engine ➜ Stage ➜ Router[%s] ➜ Error ➜ Failed to forward message to target '%s': %v", a.ID(), targetID, err)
			// Note: If sending fails to one target, it continues trying others.
		}
	}

	// Note: The RouterComponent does not need to handle the message pool.
	// The message pool is managed by the LogComponent, which is responsible for
	// returning messages to the pool after processing.
}
