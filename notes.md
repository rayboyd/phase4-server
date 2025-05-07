# Understanding Actor Pattern, Double Buffer, and Hot Path in Audio Processing

## Hot Path in Audio Processing

The "hot path" in audio processing refers to the time-critical execution path that handles real-time audio data. This is the audio callback function:

```go
func (e *Engine) processInputStream(inputBuffer []int32) {
    // Process audio samples with real-time requirements

    if len(inputBuffer) > 0 && e.system != nil && time.Now().UnixNano()%100000 == 0 {
        // Extract FFT data
        magnitudes := e.fftProc.GetMagnitudes() // Thread-safe read via double buffer

        // Send non-blocking message to actor system
        go func() {
            _ = e.system.Send("log-actor", &actor.LogFFTMagnitudesMessage{
                Magnitudes: magnitudes,
                Timestamp:  time.Now(),
            })
        }()
    }
}
```

**Characteristics of the Hot Path:**

- **Real-time requirements**: Must complete within the audio buffer duration (typically 5-10ms)
- **No blocking operations**: Cannot wait for locks, I/O, or memory allocation
- **Deterministic performance**: Must have consistent execution time
- **Zero garbage collection**: Cannot allocate memory that triggers GC
- **Thread affinity**: Usually locked to a specific CPU core

## Actor Model

The actor model is a concurrency design pattern where "actors" are the universal primitives of computation:

```go
type LogComponent struct {
    id                string
    mailbox           chan stage.Message
    ctx               context.Context
    started           bool
    stopped           bool
    messagesProcessed uint64
}

func (a *LogComponent) processMessage(ctx context.Context, msg stage.Message) error {
    switch m := msg.(type) {
    case *LogFrameMessage:
        log.Printf("LogComponent[%s]: Processed %d messages, latest buffer size: %d",
            a.id, a.messagesProcessed, m.FrameCount)
        return nil
    case *LogFFTMagnitudesMessage:
        // Find peak magnitude
        maxVal, maxIdx := 0.0, 0
        for i, v := range m.Magnitudes {
            if v > maxVal {
                maxVal, maxIdx = v, i
            }
        }

        log.Printf("LogComponent[%s]: FFT max: %f at index %d (of %d)",
            a.id, maxVal, maxIdx, len(m.Magnitudes))
        return nil
    }
    return nil
}
```

**Key Properties of Actors:**

- **Encapsulation**: Internal state is private and protected from direct access
- **Message passing**: Communication only through messages, not shared memory
- **Asynchronous processing**: Messages are processed independently
- **Isolation**: Failures in one actor don't directly affect others
- **Concurrency**: Multiple actors run concurrently without explicit locks

## Double Buffer Pattern

The double buffer pattern provides thread-safe data exchange between the real-time audio thread and non-real-time processing:

```go
type DoubleBuffer[T any] struct {
    buffers [2]T
    active  atomic.Uint32
    mu      sync.Mutex
}

// Get returns a copy of the current read buffer (thread-safe, no locks)
func (b *DoubleBuffer[T]) Get() T {
    activeIdx := b.active.Load()
    return deepCopy(b.buffers[activeIdx])
}

// Swap updates the inactive buffer and makes it the active one
func (b *DoubleBuffer[T]) Swap(updateFn func(*T)) {
    b.mu.Lock()
    defer b.mu.Unlock()

    // Get inactive buffer index
    activeIdx := b.active.Load()
    inactiveIdx := 1 - activeIdx

    // Update the inactive buffer
    updateFn(&b.buffers[inactiveIdx])

    // Swap the active buffer
    b.active.Store(inactiveIdx)
}
```

**Key Benefits of Double Buffering:**

- **Lock-free reads**: Hot path can read data without acquiring locks
- **Thread safety**: Write operations are properly synchronized
- **No allocation**: Reuses pre-allocated buffers
- **Deep copying**: Ensures data integrity across thread boundaries

## How These Patterns Work Together

1. **Hot Path -> Double Buffer -> Actor**

   ```
   Audio Thread (real-time) → Double Buffer → Actor (non-real-time)
   ```

   The hot path writes audio data to a double buffer and signals an actor to process this data. The hot path is never blocked by actor processing.

2. **Actor -> Double Buffer -> Hot Path**

   ```
   Actor (non-real-time) → Double Buffer → Audio Thread (real-time)
   ```

   Actors process data and update double buffers. The hot path reads these buffers without synchronization overhead.

## Example: FFT Processing Flow

```
Audio Callback → FFT Component → Double Buffer → FFT Actor → WebSocket Transport
```

1. Audio callback receives audio data in real-time
2. FFT component processes data and writes to double buffer
3. FFT actor asynchronously reads the double buffer and processes data
4. Actor sends processed data to WebSocket transport for visualization

## Benefits of This Architecture

1. **Real-time safety**: The hot path never blocks on I/O or locks
2. **Separation of concerns**:
   - Real-time code focuses solely on audio processing
   - Non-real-time code (actors) handles communication, persistence, UI updates
3. **Scalability**: Additional processing can be added without affecting real-time performance
4. **Resilience**: Failures in actors don't crash the audio pipeline

## Implementation Considerations

1. **Message design**: Keep messages small and immutable
2. **Buffer sizing**: Choose appropriate sizes for actor mailboxes
3. **Thread priorities**: Set audio thread to high priority
4. **Memory management**: Pre-allocate buffers and avoid GC in the hot path
5. **Error handling**: Use non-blocking error reporting from the hot path
