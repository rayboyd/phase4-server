// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"context"
	"encoding/json"
	"log"
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
	"time"
)

func NewWstComponent(id string, capacity int, sender transport.Component) *WstComponent {
	if sender == nil {
		log.Panicf("NewWstComponent requires a non-nil DataSender")
	}

	a := &WstComponent{
		sender: sender,
	}
	a.BaseActor = *stage.NewBaseActor(id, capacity, a.processMessage)

	return a
}

func (a *WstComponent) processMessage(ctx context.Context, msg stage.Message) {
	switch m := msg.(type) {
	case *stage.FFTData:
		payloadMap := map[string]any{
			"type":          "fft_magnitudes",
			"frameCount":    m.FrameCount,
			"startTime":     m.StartTime.Format(time.RFC3339Nano),
			"magnitudes":    m.Magnitudes,
			"spectralFlux":  m.SpectralFlux,
			"bpm":           m.BPM,           // Add BPM
			"bpmConfidence": m.BPMConfidence, // Add confidence
		}

		jsonData, err := json.Marshal(payloadMap)
		if err != nil {
			return
		}

		// Send the JSON data to the WebSocket sender, ignore the error
		_ = a.sender.SendData(jsonData)

	default:
		// log something about unexpected message type
	}
}
