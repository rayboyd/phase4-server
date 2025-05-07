// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"log"
	"phase4/internal/app/errors"
	"phase4/internal/p4/runtime/stage"
	"runtime"
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

	// Stop and close the stream when the context is cancelled. If an error occurs
	// while stopping or closing the stream, the error is logged, do not return errors
	// to ensure that the stream is closed properly even when an error occurs.
	defer func() {
		if stopErr := e.audio.stream.Stop(); stopErr != nil {
			log.Printf("Engine ➜ Error failed to stop PortAudio stream: %v", stopErr)
		}
		if closeErr := e.audio.stream.Close(); closeErr != nil {
			log.Printf("Engine ➜ Error failed to close PortAudio stream: %v", closeErr)
		}
		e.audio.stream = nil
	}()

	if err := e.audio.stream.Start(); err != nil {
		e.audio.stream = nil
		return &errors.FatalError{
			Message: "failed to start PortAudio stream",
			Err:     err,
		}
	}
	log.Print("Engine ➜ Stream ➜ Started. (Ctrl+C) or (SigTerm) to stop.")

	// Wait for the context to be cancelled, which will happen when the program
	// receives an interrupt signal (Ctrl+C) or a termination signal (SIGTERM).
	<-ctx.Done()
	log.Print("Engine ➜ run() terminated")

	return nil
}

func (e *Engine) processInputStream(inputBuffer []int32) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	frameCount := e.frameCount.Add(1)

	if e.fftProc == nil || e.system == nil {
		return
	}

	// TODO: Current process() func, it works "good enough" for pre-alpha testing.
	e.fftProc.Process(inputBuffer)

	magnitudes := e.fftProc.GetMagnitudes() // ALLOC invoved here?
	if len(magnitudes) == 0 {
		return
	}

	rawMsg := &stage.MagnitudesMessage{
		Magnitudes: magnitudes,
		FrameCount: frameCount,
	}
	_ = e.system.Send("processor", rawMsg)
}
