// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"fmt"
	"phase4/internal/app/config"
	"phase4/internal/app/errors"
	"phase4/internal/p4/analysis"
	"phase4/internal/p4/runtime/endpoint"
	"phase4/internal/p4/runtime/pipeline"
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
)

// NewEngine creates a new audio engine instance with the provided configuration.
// It initializes internal data structures but does not start audio processing.
func NewEngine(cfg *config.Config) *Engine {
	return engine(cfg)
}

func engine(cfg *config.Config) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		config:    cfg,
		command:   &cmd{},
		closables: make([]interface{ Close() error }, 0),
		ctx:       ctx,
		cancel:    cancel,
		system:    stage.NewSystem(),
		audio: &pa{
			client:      newEnginePaClient(),
			initialized: false,
		},
	}
}

func (e *Engine) Initialize() error {
	if err := e.initializePortAudio(); err != nil {
		return err
	}
	if err := e.initializeAnalysis(); err != nil {
		return err
	}
	if err := e.initializeSystem(); err != nil {
		return err
	}
	if err := e.selectAndConfigureDevice(); err != nil {
		return err
	}
	return nil
}

func (e *Engine) initializePortAudio() error {
	if err := initPA(e); err != nil {
		return &errors.FatalError{
			Message: "failed to initialize PortAudio",
			Err:     err,
		}
	}
	return nil
}

func (e *Engine) initializeAnalysis() error {
	fftWindowFunc, _ := analysis.ParseWindowFunc(e.config.DSP.FFTWindow)
	fftProcessor, err := analysis.NewFFTProcessor(
		e.config.Input.BufferSize,
		e.config.Input.SampleRate,
		fftWindowFunc,
	)
	if err != nil {
		return &errors.FatalError{
			Message: "failed to create FFT processor",
			Err:     err,
		}
	}
	e.fftProc = fftProcessor
	e.closables = append(e.closables, fftProcessor)

	e.bpmDetector = analysis.NewBPMDetector(
		e.config.Input.SampleRate,
		e.config.Input.BufferSize,
	)

	return nil
}

func (e *Engine) initializeSystem() error {
	routerTargets := []string{}
	capacity := 2024

	// Processor -> Router -> Transport

	processorComponent, err := pipeline.NewProcessor("processor", capacity, "router", e.system)
	if err != nil {
		return &errors.FatalError{
			Message: "failed to create ProcessorComponent",
			Err:     err,
		}
	}
	if err := e.system.Register(processorComponent); err != nil {
		return &errors.FatalError{
			Message: "failed to register ProcessorComponent",
			Err:     err,
		}
	}

	if e.config.Transport.WebSocketEnabled {
		wsTransport, err := transport.NewWebSocketTransport(
			e.config.Transport.WebSocketAddress,
			e.config.Transport.WebSocketPath,
		)
		if err != nil {
			return &errors.FatalError{
				Message: "failed to create WebSocketTransport",
				Err:     err,
			}
		}
		e.closables = append(e.closables, wsTransport)

		wstComponent := endpoint.NewWstComponent("ws", capacity, wsTransport)
		if err := e.system.Register(wstComponent); err != nil {
			return &errors.FatalError{
				Message: "failed to register WstComponent",
				Err:     err,
			}
		}
		routerTargets = append(routerTargets, "ws")
	}

	routerComponent, err := pipeline.NewRouter("router", capacity, routerTargets, e.system)
	if err != nil {
		return &errors.FatalError{
			Message: "failed to create RouterComponent",
			Err:     err,
		}
	}
	if err := e.system.Register(routerComponent); err != nil {
		return &errors.FatalError{
			Message: "failed to register RouterComponent",
			Err:     err,
		}
	}

	return nil
}

func (e *Engine) selectAndConfigureDevice() error {
	if err := selectInputDevice(e); err != nil {
		return &errors.FatalError{
			Message: "failed to select input device",
			Err:     err,
		}
	}
	printInputDevice(e.audio.inputDevice)
	return nil
}

func (e *Engine) Run(ctx context.Context) error {
	if err := e.system.StartAll(); err != nil {
		return fmt.Errorf("failed to start actor system: %v", err)
	}
	return e.startStream(ctx)
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}
	e.closed = true

	var errs []error

	// 1. Stop audio stream first (most critical)
	if e.audio.stream != nil {
		if err := e.stopAudioStream(); err != nil {
			errs = append(errs, fmt.Errorf("audio stream: %w", err))
		}
	}

	// 2. Stop actor system (may depend on other components)
	if e.system != nil {
		if err := e.system.StopAll(); err != nil {
			errs = append(errs, fmt.Errorf("actor system stop: %v", err))
		}
		if err := e.system.Close(); err != nil {
			errs = append(errs, fmt.Errorf("actor system close: %w", err))
		}
	}

	// 3. Close components in reverse order
	for i := len(e.closables) - 1; i >= 0; i-- {
		if err := e.closables[i].Close(); err != nil {
			errs = append(errs, fmt.Errorf("component %T: %w", e.closables[i], err))
		}
	}

	// 4. Terminate PortAudio last
	if err := exitPA(e); err != nil {
		errs = append(errs, fmt.Errorf("portaudio: %w", err))
	}

	if len(errs) > 0 {
		return &errors.FatalError{
			Message: "shutdown errors occurred",
			Err:     fmt.Errorf("multiple errors: %v", errs),
		}
	}

	return nil
}
