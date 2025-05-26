# Phase4 Audio Engine

A high-performance, real-time audio processing engine built in Go with actor-based concurrency and lock-free hot path design.

## Architecture

### Core Components

**Engine** - Central orchestrator managing initialization, execution, and shutdown of all components with proper lifecycle management.

**Actor System** - Message-passing concurrency model with non-blocking sends from the audio hot path. Actors handle FFT processing, routing, and transport endpoints.

**Hot Path** - Lock-free, allocation-free audio callback (`processInputStream`) that processes real-time audio data without blocking operations.

**Transport Layer** - WebSocket and UDP endpoints for streaming processed audio data to external clients.

### Key Design Patterns

**Lock-Free Hot Path**

```go
func (e *Engine) processInputStream(inputBuffer []int32) {
    // Non-blocking FFT processing
    // Message pool allocation (no GC pressure)
    // Non-blocking actor system send
}
```

**Actor Message Flow**

```
Audio Callback → Processor Actor → Router Actor → Transport Endpoints
```

**Lifecycle Management**

```
Initialize() → Start() → Run() → Shutdown() → Close()
```

### Real-Time Safety

- **No allocations** in audio callback (message pooling)
- **No locks** in hot path (lock-free reads)
- **Non-blocking sends** to actor system
- **Frame dropping** under load (never block audio thread)
- **Graceful degradation** when actors are busy

### Performance Characteristics

- **Buffer size**: 256 samples (configurable)
- **Sample rate**: 44.1kHz (configurable)
- **Latency**: ~10ms (low latency mode)
- **Concurrency**: Actor-based, scales with available cores
- **Memory**: Pre-allocated buffers, minimal GC pressure

## Development

This project uses Go modules and a `Makefile` for common development tasks.

### Prerequisites

- **Go:** Version 1.24 or later.
- **Make:** Standard build utility.
- **PortAudio:** Required for audio I/O. Install via Homebrew:
  ```bash
  brew install portaudio
  ```
- **(Optional) golangci-lint:** For code linting. Install via Homebrew or see [official instructions](https://golangci-lint.run/usage/install/).
  ```bash
  brew install golangci-lint
  ```

### Building

To build the application binary:

```bash
make build
```

The binary will be placed in the `bin` directory.

### Running

To build and run the audio engine:

```bash
make run
```

The engine will start processing audio from the default input device and serve FFT visualization data on `ws://127.0.0.1:8889/ws`.

### Testing

To run all tests:

```bash
make test
```

To run tests and generate a coverage report:

```bash
make cover
# Then view the report:
go tool cover -html=coverage.out
```

### Race Detection

To run tests with race detection:

```bash
make race
```

This will run critical components with Go's race detector to identify potential race conditions.

### Benchmarking

To run performance benchmarks:

```bash
make bench
```

### Linting

To run the code linter:

```bash
make lint
```

### Cleaning

To remove build artifacts (binary and coverage report):

```bash
make clean
```

### Available Commands

Run `make help` to see all available commands defined in the `Makefile`.

## Configuration

The engine uses YAML configuration with environment variable overrides:

```yaml
input:
  device: -1 # -1 for default device
  channels: 1 # Mono input
  buffer_size: 256 # Samples per buffer
  sample_rate: 44100 # Hz
  low_latency: true # Use low-latency audio buffers

transport:
  websocket_enabled: true
  websocket_address: "127.0.0.1:8889"
  websocket_path: "/ws"

dsp:
  fft_window: "hann" # Window function for FFT
```

## Client Integration

Connect to the WebSocket endpoint to receive real-time FFT data:

```javascript
const ws = new WebSocket("ws://127.0.0.1:8889/ws");
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // data.magnitudes contains FFT magnitude array
  // data.frameCount contains audio frame counter
};
```

A complete visualization client is available at `public/index.html`.

## Roadmap

Roadmap to `0.0.1`

- ✅ startup/shutdown logic centralise
- tests
- wav recording (stream copy and fifo buffer)
- more tests
- udp
