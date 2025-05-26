// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"fmt"
	"log"
	"phase4/internal/app/errors"
	"phase4/internal/p4/runtime/stage"
	"time"

	"github.com/gordonklaus/portaudio"
)

func (e *Engine) startStream(ctx context.Context) error {
	if e.audio.stream != nil {
		log.Print("Engine ➜ Stream already active")
		return nil
	}

	var latency time.Duration
	if e.config.Input.LowLatency {
		latency = e.audio.inputDevice.DefaultLowInputLatency
	} else {
		latency = e.audio.inputDevice.DefaultHighInputLatency
	}
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   e.audio.inputDevice,
			Channels: e.config.Input.Channels,
			Latency:  latency,
		},
		SampleRate:      e.config.Input.SampleRate,
		FramesPerBuffer: e.config.Input.BufferSize,
	}
	log.Printf("Engine ➜ Stream ➜ SampleRate: %.2f, BufferSize: %d, Channels: %d",
		streamParams.SampleRate,
		streamParams.FramesPerBuffer,
		streamParams.Input.Channels,
	)

	stream, err := e.audio.client.OpenStream(streamParams, e.processInputStream)
	if err != nil {
		return &errors.FatalError{
			Message: "failed to open PortAudio stream",
			Err:     err,
		}
	}
	e.audio.stream = stream

	if err := e.audio.stream.Start(); err != nil {
		e.audio.stream = nil
		return &errors.FatalError{
			Message: "failed to start PortAudio stream",
			Err:     err,
		}
	}
	log.Print("Engine ➜ Stream ➜ Started. (Ctrl+C) or (SigTerm) to stop.")

	// Wait for the context to be cancelled
	<-ctx.Done()
	log.Print("Engine ➜ run() terminated")

	return nil
}

func (e *Engine) processInputStream(inputBuffer []int32) {
	frameCount := e.frameCount.Add(1)

	if e.fftProc == nil || e.system == nil {
		return
	}

	e.fftProc.Process(inputBuffer)
	magnitudes := e.fftProc.GetMagnitudes()
	spectralFlux := e.fftProc.GetSpectralFlux()

	if len(magnitudes) == 0 {
		return
	}

	// Process flux for BPM detection
	var bpm, confidence float64
	if e.bpmDetector != nil {
		e.bpmDetector.ProcessFlux(spectralFlux, frameCount)
		bpm, confidence = e.bpmDetector.GetBPM()
	}

	// Pre-allocate this message to avoid hot path allocation
	rawMsg := stage.GetRawMessage()
	rawMsg.Magnitudes = magnitudes
	rawMsg.SpectralFlux = spectralFlux
	rawMsg.FrameCount = frameCount
	rawMsg.BPM = bpm
	rawMsg.BPMConfidence = confidence

	// Non-blocking send - if system is busy, drop the frame
	select {
	case <-e.ctx.Done():
		stage.PutRawMessage(rawMsg)
		return
	default:
		if err := e.system.SendNonBlocking("processor", rawMsg); err != nil {
			stage.PutRawMessage(rawMsg) // Return to pool on error
		}
	}
}

func (e *Engine) stopAudioStream() error {
	if e.audio.stream == nil {
		return nil
	}

	var errs []error

	// Stop first, then close
	if err := e.audio.stream.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("stop: %w", err))
	}

	if err := e.audio.stream.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close: %w", err))
	}

	e.audio.stream = nil

	if len(errs) > 0 {
		return fmt.Errorf("stream shutdown errors: %v", errs)
	}

	return nil
}
