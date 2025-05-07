# Phase4 Audio Engine

A high-performance, real-time audio processing engine built in Go.

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

## Roadmap

Roadmap to `0.0.1`

- startup/shutdown logic centralise
- tests
- wav recording (stream copy and fifo buffer)
- more tests
- udp
