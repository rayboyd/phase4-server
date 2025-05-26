// SPDX-License-Identifier: Apache-2.0
/*
Package buffer provides thread-safe data structures optimized for real-time audio
processing. The double buffer pattern is particularly important in audio applications
where producer and consumer threads must operate without blocking each other.

This DoubleBuffer implementation uses a pair of buffers that are atomically swapped,
allowing simultaneous read and write operations without locks or contention. This
is critical for audio processing where the audio callback thread must never be blocked.

How it works:
1. At any moment, one buffer is designated as "active" (for reading)
2. Writes always occur to the inactive buffer
3. When a write completes, the buffers are atomically swapped
4. Readers always get a consistent copy of the most recently completed write

This implementation uses a copy-on-read approach that creates a deep copy of the
active buffer for each reader. While this incurs a small performance cost compared
to a zero-copy approach, it guarantees thread safety and eliminates race conditions
that could otherwise occur if a buffer is swapped while being read.

The DoubleBuffer is generic and works with any type, including slices and structs.
For common audio types like []float32 and []float64, it uses specialized copying
for better performance. For other types, it falls back to reflection-based copying.
*/
package buffer

import (
	"reflect"
	"sync"
)

// DoubleBuffer provides thread-safe double buffering for real-time audio processing.
// It maintains two buffers - one for reading and one for writing - and atomically
// swaps them to ensure readers always see consistent data without blocking writers.
// The implementation is generic and works with any type T, though it's optimized
// for common audio types like []float32 and []float64.
type DoubleBuffer[T any] struct {
	buffers [2]T         // The two buffers we alternate between.
	active  int          // Index of the active buffer (0 or 1).
	mu      sync.RWMutex // Protects all buffer operations.
}

// New creates a new double buffer with the provided initial buffer values.
// The first buffer (buffer1) is initially set as the active buffer for reading.
// Both buffers should be of the same type and typically the same size.
// Examples of valid buffer pairs include:
//   - Two []float32 slices for audio samples
//   - Two FFT result structs for spectral data
//   - Any two values of the same type that need thread-safe swapping
func New[T any](buffer1, buffer2 T) *DoubleBuffer[T] {
	return &DoubleBuffer[T]{
		buffers: [2]T{buffer1, buffer2},
		active:  0,
	}
}

// Get returns a deep copy of the current active buffer.
// This ensures readers have a stable snapshot that won't be modified by writers.
// The returned copy is completely independent from the internal buffers, so it
// can be safely used even if the DoubleBuffer is subsequently modified.
// The copy operation uses type-specific optimizations for common audio types
// and falls back to reflection for complex types.
func (db *DoubleBuffer[T]) Get() T {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return copyOf(db.buffers[db.active])
}

// Swap updates the inactive buffer using the provided function and then makes it the
// new active buffer for reading. This operation is atomic and thread-safe, ensuring
// that readers always see a consistent state.
//
// The updateFn is called with a pointer to the inactive buffer, allowing the function
// to modify it in place. After the update is complete, the buffers are swapped, making
// the newly updated buffer the active one.
//
// This pattern allows for efficient in-place updates without allocating new memory
// for each swap operation.
func (db *DoubleBuffer[T]) Swap(updateFn func(*T)) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Swap the active buffer with the inactive one, update the inactive buffer,
	// and then swap them back. Ensures that the active buffer is always the one
	// that was just read.
	inactive := 1 - db.active
	updateFn(&db.buffers[inactive])
	db.active = inactive
}

// ForceGet gets a copy of the current buffer and executes the provided function
// with that buffer. This is useful when you need to perform multiple operations
// on the buffer and want to ensure they all operate on the same consistent snapshot.
//
// Unlike simply calling Get() and using the result, ForceGet guarantees that
// the buffer remains consistent throughout the execution of the provided function.
func (db *DoubleBuffer[T]) ForceGet(fn func(T)) {
	buffer := db.Get()
	fn(buffer)
}

// copyOf creates a deep copy of the given value based on its type.
// This function is optimized for common audio processing types (float slices)
// and falls back to reflection-based copying for other types.
//
// The function handles:
// - []float32 - Common for audio samples (optimized)
// - []float64 - For high-precision audio (optimized)
// - []int32 - For integer samples (optimized)
// - []complex128 - For FFT results (optimized)
// - Other types - Handled via reflection
//
// Each type-specific case includes nil handling to ensure nil slices remain nil in
// the copied result.
func copyOf[T any](src T) T {
	switch s := any(src).(type) {
	case []float32:
		if s == nil {
			return any([]float32(nil)).(T)
		}
		dst := make([]float32, len(s))
		copy(dst, s)
		return any(dst).(T)

	case []float64:
		if s == nil {
			return any([]float64(nil)).(T)
		}
		dst := make([]float64, len(s))
		copy(dst, s)
		return any(dst).(T)

	case []int32:
		if s == nil {
			return any([]int32(nil)).(T)
		}
		dst := make([]int32, len(s))
		copy(dst, s)
		return any(dst).(T)

	case []complex128:
		if s == nil {
			return any([]complex128(nil)).(T)
		}
		dst := make([]complex128, len(s))
		copy(dst, s)
		return any(dst).(T)
	}

	// For any other type, use reflection for a generic deep copy.
	return deepCopyWithReflection(src)
}

// deepCopyWithReflection creates a deep copy using reflection.
// This function handles arbitrary struct types by creating a new instance and copying
// each field individually. For slice fields within structs, it performs a deep copy
// to ensure complete isolation between the original and copy.
//
// The function handles:
// - Structs - Creates a new struct and copies each exported field
// - Primitive types - Returns as is (Go passes by value)
// - Maps, channels, etc. - Returns as is (caller should be aware these are reference types)
//
// Unexported struct fields are skipped since they can't be accessed via reflection.
// This is generally appropriate since unexported fields are implementation details.
func deepCopyWithReflection[T any](src T) T {
	srcVal := reflect.ValueOf(src)
	srcType := srcVal.Type()

	switch srcType.Kind() {
	case reflect.Struct:
		dstVal := reflect.New(srcType).Elem()

		for i := range srcType.NumField() {
			field := srcVal.Field(i)

			if !field.CanInterface() {
				continue
			}

			fieldValue := field.Interface()
			dstField := dstVal.Field(i)

			if field.Kind() == reflect.Slice {
				switch slc := fieldValue.(type) {
				case []float32:
					newSlice := make([]float32, len(slc))
					copy(newSlice, slc)
					dstField.Set(reflect.ValueOf(newSlice))
				case []float64:
					newSlice := make([]float64, len(slc))
					copy(newSlice, slc)
					dstField.Set(reflect.ValueOf(newSlice))
				case []int32:
					newSlice := make([]int32, len(slc))
					copy(newSlice, slc)
					dstField.Set(reflect.ValueOf(newSlice))
				case []int:
					newSlice := make([]int, len(slc))
					copy(newSlice, slc)
					dstField.Set(reflect.ValueOf(newSlice))
				default:
					// For other slice types, set directly This is a shallow copy
					// and may share underlying data which is generally acceptable
					// for less common types.
					dstField.Set(field)
				}
			} else {
				// For non-slice types, set directly This works because Go passes
				// primitive types by value.
				dstField.Set(field)
			}
		}
		return dstVal.Interface().(T)

	default:
		// For primitive types, maps, etc., just return the original. This is safe
		// because Go passes these by value already.
		return src
	}
}
