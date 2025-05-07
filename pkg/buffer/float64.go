// SPDX-License-Identifier: Apache-2.0
package buffer

import (
	"sync"
)

// Float64DoubleBuffer provides a thread-safe double buffer specifically for []float64.
// It maintains two []float64 buffers - one for reading and one for writing -
// and atomically swaps them.
type Float64DoubleBuffer struct {
	buffers [2][]float64 // The two buffers we alternate between.
	active  int          // Index of the active buffer (0 or 1).
	mu      sync.RWMutex // Protects all buffer operations.
}

// NewFloat64DoubleBuffer creates a new double buffer for []float64
// with the provided initial buffer values.
// The first buffer (buffer1) is initially set as the active buffer for reading.
func NewFloat64DoubleBuffer(buffer1, buffer2 []float64) *Float64DoubleBuffer {
	return &Float64DoubleBuffer{
		buffers: [2][]float64{buffer1, buffer2},
		active:  0,
	}
}

// Get returns a copy of the current active []float64 buffer.
// This ensures readers have a stable snapshot.
func (db *Float64DoubleBuffer) Get() []float64 {
	db.mu.RLock()
	defer db.mu.RUnlock()

	src := db.buffers[db.active]
	if src == nil {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

// Swap updates the inactive []float64 buffer using the provided function
// and then makes it the new active buffer for reading.
func (db *Float64DoubleBuffer) Swap(updateFn func(buffer *[]float64)) {
	db.mu.Lock()
	defer db.mu.Unlock()

	inactive := 1 - db.active
	updateFn(&db.buffers[inactive])
	db.active = inactive
}

// ForceGet gets a copy of the current []float64 buffer and executes the provided
// function with that buffer.
func (db *Float64DoubleBuffer) ForceGet(fn func(buffer []float64)) {
	buffer := db.Get()
	fn(buffer)
}
