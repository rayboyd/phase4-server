// SPDX-License-Identifier: Apache-2.0
package pipeline

import (
	"context"
	"fmt"
	"log"
	"phase4/internal/p4/runtime/stage"
	"time"
)

func NewProcessor(id string, capacity int, routerID string, system *stage.System) (*ProcessorComponent, error) {
	if system == nil {
		return nil, fmt.Errorf("ProcessorComponent[%s] requires a non-nil system", id)
	}
	if routerID == "" {
		return nil, fmt.Errorf("ProcessorComponent[%s] requires a non-empty routerID", id)
	}

	a := &ProcessorComponent{
		routerID: routerID,
		system:   system,
	}
	a.BaseActor = *stage.NewBaseActor(id, capacity, a.processMessage)

	return a, nil
}

func (a *ProcessorComponent) processMessage(ctx context.Context, msg stage.Message) {
	rawMsg, ok := msg.(*stage.MagnitudesMessage)
	if !ok {
		log.Printf("Processor[%s] ➜ Warning ➜ Received unexpected message type: %T", a.ID(), msg)
		return
	}

	fftMsg := FftDataPool.Get().(*stage.FFTData)
	fftMsg.FrameCount = rawMsg.FrameCount
	fftMsg.StartTime = time.Now()

	if cap(fftMsg.Magnitudes) < len(rawMsg.Magnitudes) {
		fftMsg.Magnitudes = make([]float64, len(rawMsg.Magnitudes))
	} else {
		fftMsg.Magnitudes = fftMsg.Magnitudes[:len(rawMsg.Magnitudes)]
	}
	copy(fftMsg.Magnitudes, rawMsg.Magnitudes)

	if err := a.system.Send(a.routerID, fftMsg); err != nil {
		log.Printf("Processor[%s] ➜ Error ➜ Failed to send message to router '%s': %v", a.ID(), a.routerID, err)
		FftDataPool.Put(fftMsg)
	}
}
