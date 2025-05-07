// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"log"
	"os"
	"os/signal"
	"phase4/internal/app/config"
	"phase4/internal/app/errors"
	"phase4/internal/p4/analysis"
	"phase4/internal/p4/runtime/endpoint"
	"phase4/internal/p4/runtime/pipeline"
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
	"syscall"
)

// NewEngine creates a new audio engine instance with the provided configuration.
// It initializes internal data structures but does not start audio processing.
func NewEngine(cfg *config.Config) *Engine {
	return engine(cfg)
}

func engine(cfg *config.Config) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &Engine{
		config:    cfg,
		command:   &cmd{},
		closables: make([]interface{ Close() error }, 0),
		ctx:       ctx,
		cancel:    cancel,
		system:    stage.NewSystem(),
	}

	engine.audio = &pa{
		client:      newEnginePaClient(),
		initialized: false,
	}

	return engine
}

// Run initializes PortAudio and runs the main loop. It sets up signal handling
// to allow the engine to shut down gracefully. It will return errors from
// initializing PortAudio or from the main loop.
func (e *Engine) Initialize() error {
	if err := initPA(e); err != nil {
		return &errors.FatalError{
			Message: "failed to initialize PortAudio",
			Err:     err,
		}
	}

	//if e.config.DSP.Enabled {}
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

	// --- Components ---

	routerTargets := []string{}
	capacity := e.config.Input.BufferSize // TODO: use a better value.

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

	// if e.config.Transport.UDPEnabled {
	// 	udpComponent := component.NewUdpComponent("udp", capacity, e.udpTransport)
	// 	if err := e.system.Register(udpComponent); err != nil {
	// 		return &app.FatalError{
	// 			Message: "failed to register UdpComponent",
	// 			Err:     err,
	// 		}
	// 	}
	// 	e.closables = append(e.closables, udpComponent)
	// }

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

	if componenrErr := e.system.StartAll(); componenrErr != nil {
		for componentID, err := range componenrErr {
			return &errors.FatalError{
				Message: "failed to start component: " + componentID,
				Err:     err,
			}
		}
	}

	// Exit MUST be called before exiting the program. Failure to do so may result
	// in serious resource leaks, such as audio devices not being available until
	// the next reboot.
	defer func() {
		if err := exitPA(e); err != nil {
			log.Printf("Engine ➜ Error failed to terminate PortAudio: %v", err)
		}
	}()

	// --- Commands ---
	/*if e.command.ListDevices {
	log.Print("Engine ➜ Listing available audio devices...")
	var builder strings.Builder
	builder.WriteString("Engine ➜ Available Audio Devices:\n")
	if len(e.audio.devices) == 0 {
		builder.WriteString("  No devices found.\n")
	} else {
		for i, dev := range e.audio.devices {
			fmt.Fprintf(&builder, "  [%d] %s (In:%d, Out:%d, Rate:%.0f, API:%s)\n",
				i, dev.Name, dev.MaxInputChannels, dev.MaxOutputChannels,
				dev.DefaultSampleRate, dev.HostApi.Name)
		}
	}
	return &app.CommandCompleted{Message: builder.String()}
	}*/

	if err := selectInputDevice(e); err != nil {
		return &errors.FatalError{
			Message: "failed to select input device",
			Err:     err,
		}
	}
	printInputDevice(e.audio.inputDevice)

	// This context will be cancelled when the program receives an interrupt signal.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling, which will allow the engine to shut down gracefully
	// when receiving an interrupt signal (Ctrl+C) or a termination signal (SIGTERM).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("\nEngine ➜ Received signal: %v, shutting down...", sig)
		cancel()
	}()

	return e.startStream(ctx)
}

func (e *Engine) Close() error {
	var firstErr error

	// Close the system first, it may have dependencies on other components.
	if e.system != nil {
		if err := e.system.Close(); err != nil {
			firstErr = err
			log.Printf("Engine ➜ System ➜ Failed to close ➜ %v", err)
		}
	}

	// Close LIFO to ensure dependencies are closed first.
	for i := len(e.closables) - 1; i >= 0; i-- {
		closable := e.closables[i]
		log.Printf("Engine ➜ Closing ➜ %T ...\n", closable)
		if err := closable.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("Engine ➜ Failed to close ➜ %T ...\n", closable)
		}
	}

	if firstErr == nil {
		log.Print("Engine ➜ All components closed successfully.")
		return nil
	} else {
		return &errors.FatalError{
			Message: "failed to close some components",
			Err:     firstErr,
		}
	}
}
