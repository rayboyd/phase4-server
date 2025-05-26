// SPDX-License-Identifier: Apache-2.0
package stage

import (
	"sync"
	"time"
)

const (
	TypeControl     = "control"
	TypeData        = "data"
	TypeStatus      = "status"
	TypeRawAudioFFT = "data.audio.fft.raw"       // From hot path -> ingress
	TypeFFTData     = "data.audio.fft.processed" // From ingress -> router -> endpoints
)

type ControlMessage struct {
	Params  map[string]any
	Command string
}

func (m *ControlMessage) Type() string {
	return TypeControl
}

type DataMessage struct {
	Data   any
	Format string
}

func (m *DataMessage) Type() string {
	return TypeData
}

type StatusMessage struct {
	Details map[string]any
	ActorID string
	Status  string
}

func (m *StatusMessage) Type() string {
	return TypeStatus
}

type RawAudioMessage struct {
	Magnitudes    []float64
	SpectralFlux  []float64
	FrameCount    uint64
	BPM           float64
	BPMConfidence float64
}

func (m *RawAudioMessage) Type() string {
	return TypeRawAudioFFT
}

type FFTData struct {
	StartTime     time.Time
	Magnitudes    []float64
	SpectralFlux  []float64
	FrameCount    uint64
	BPM           float64
	BPMConfidence float64
}

func (m *FFTData) Type() string {
	return TypeFFTData
}

var RawMessagePool = sync.Pool{
	New: func() any {
		return &RawAudioMessage{
			Magnitudes: make([]float64, 0, 129), // Pre-allocate typical FFT size
		}
	},
}

func GetRawMessage() *RawAudioMessage {
	return RawMessagePool.Get().(*RawAudioMessage)
}

func PutRawMessage(msg *RawAudioMessage) {
	msg.Magnitudes = msg.Magnitudes[:0] // Reset slice but keep capacity
	msg.FrameCount = 0
	RawMessagePool.Put(msg)
}
