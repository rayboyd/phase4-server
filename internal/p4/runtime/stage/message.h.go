// SPDX-License-Identifier: Apache-2.0
package stage

import "time"

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

type MagnitudesMessage struct {
	Magnitudes []float64
	FrameCount uint64
}

func (m *MagnitudesMessage) Type() string {
	return TypeRawAudioFFT
}

type FFTData struct {
	StartTime  time.Time
	Magnitudes []float64
	FrameCount uint64
}

func (m *FFTData) Type() string {
	return TypeFFTData
}
