// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"context"
	"phase4/internal/app/config"
	"phase4/internal/p4/analysis"
	"phase4/internal/p4/runtime/stage"
	"sync"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

type Engine struct {
	ctx         context.Context
	audio       *pa
	command     *cmd
	config      *config.Config
	system      *stage.System
	cancel      context.CancelFunc
	fftProc     *analysis.FFTProcessor
	bpmDetector *analysis.BPMDetector
	closables   []interface{ Close() error }
	frameCount  atomic.Uint64
	mu          sync.Mutex
	closed      bool
}

type cmd struct {
	ListDevices bool
}

type pa struct {
	client      paClient
	stream      paStream
	inputDevice *portaudio.DeviceInfo
	devices     []*portaudio.DeviceInfo
	initialized bool
}

// paClient abstracts the PortAudio library to allow for easier testing and mocking, it
// defines the interface for interacting with PortAudio.
type paClient interface {
	Initialize() error
	Terminate() error
	Devices() ([]*portaudio.DeviceInfo, error)
	DefaultInputDevice() (*portaudio.DeviceInfo, error)
	OpenStream(params portaudio.StreamParameters, callback func([]int32)) (paStream, error)
}

// paStream abstracts the PortAudio stream to allow for easier testing and mocking,
// it defines the interface for interacting with PortAudio streams.
type paStream interface {
	Start() error
	Stop() error
	Close() error
}

// This is an implementation of the paClient interface that uses the PortAudio library.
// It provides methods to initialize and terminate PortAudio. Allows for easier testing
// and mocking of the PortAudio library.
type livePaClient struct{}

func newEnginePaClient() paClient {
	return &livePaClient{}
}

func (c *livePaClient) Initialize() error {
	return portaudio.Initialize()
}

func (c *livePaClient) Terminate() error {
	return portaudio.Terminate()
}

func (c *livePaClient) Devices() ([]*portaudio.DeviceInfo, error) {
	return portaudio.Devices()
}

func (c *livePaClient) DefaultInputDevice() (*portaudio.DeviceInfo, error) {
	return portaudio.DefaultInputDevice()
}

func (c *livePaClient) OpenStream(params portaudio.StreamParameters, callback func([]int32)) (paStream, error) {
	stream, err := portaudio.OpenStream(params, callback)
	if err != nil {
		return nil, err
	}

	return &livePaStream{stream: stream}, nil
}

// mockPaClient is a mock implementation of the paClient interface for testing purposes.
// It allows for tracking whether the Initialize, Terminate, Devices, DefaultInputDevice,
// and OpenStream methods were called, and allows for simulating errors in those methods.
/*
type mockPaClient struct {
	InitializeCalled         bool
	InitializeErr            error
	TerminateCalled          bool
	TerminateErr             error
	DevicesCalled            bool
	DevicesErr               error
	DefaultInputDeviceCalled bool
	DefaultInputDeviceErr    error
	DefaultInputDeviceResult *portaudio.DeviceInfo
	DevicesResult            []*portaudio.DeviceInfo
	OpenStreamCalled         bool
	OpenStreamParams         portaudio.StreamParameters
	OpenStreamResult         paStream
	OpenStreamErr            error
}
*/

// livePaStream is an implementation of the paStream interface that uses the PortAudio
// library. It provides methods to start, stop, and close the stream. Allows for easier
// testing and mocking of the PortAudio library.
type livePaStream struct {
	stream *portaudio.Stream
}

func (s *livePaStream) Start() error {
	return s.stream.Start()
}

func (s *livePaStream) Stop() error {
	return s.stream.Stop()
}

func (s *livePaStream) Close() error {
	return s.stream.Close()
}

// mockPAStream is a mock implementation of the paStream interface for testing purposes.
// It allows for tracking whether the Start, Stop, and Close methods were called, and allows
// for simulating errors in those methods.
/*
type mockPaStream struct {
	StartCalled bool
	StopCalled  bool
	CloseCalled bool
	StartErr    error
	StopErr     error
	CloseErr    error
}
*/
