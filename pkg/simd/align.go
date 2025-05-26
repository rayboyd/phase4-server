// SPDX-License-Identifier: Apache-2.0
/*
Package simd provides memory allocation utilities optimized for SIMD (Single
Instruction Multiple Data) operations. It ensures that the starting address of the
slice's underlying data is aligned to a specific byte boundary (e.g., 16 bytes),
which can significantly improve performance for certain CPU vector instructions.
*/
package simd

import (
	"unsafe"
)

// alignment specifies the desired memory alignment boundary in bytes.
// Common values are 16 (SSE), 32 (AVX), or 64 (AVX512) depending on the target SIMD
// instruction set requirements.
const alignment = 16

// AlignedFloat64 returns a float64 slice of the requested size. The underlying data
// buffer's starting address is guaranteed to be aligned to the package's 'alignment'
// constant (16 bytes). If size is 0, it returns nil.
func AlignedFloat64(size int) []float64 {
	if size == 0 {
		return nil
	}

	// Calculate the total buffer size needed. We need enough space for 'size' elements
	// plus potential padding bytes required to reach the next alignment boundary.
	// Since float64 is 8 bytes, we might need up to (alignment/8 - 1) extra elements.
	totalSize := size + (alignment/int(unsafe.Sizeof(float64(0))) - 1)
	rawBuffer := make([]float64, totalSize)

	// Get the memory address of the first element in the allocated buffer.
	startPtr := uintptr(unsafe.Pointer(&rawBuffer[0]))

	// Calculate the aligned starting address within the raw buffer.
	// (startPtr + alignment - 1) ensures we reach into the next alignment block.
	// &^ (alignment - 1) masks off the lower bits to round down to the nearest  alignment boundary.
	alignedPtr := (startPtr + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1)

	// Calculate the offset (in number of elements) from the start of the raw buffer
	// to the aligned starting position.
	offset := (alignedPtr - startPtr) / uintptr(unsafe.Sizeof(float64(0)))

	// Create the final slice starting at the aligned offset with the requested size.
	// This slice shares the underlying memory with rawBuffer but starts at the aligned address.
	alignedSlice := rawBuffer[offset : offset+uintptr(size)]

	// Sanity check: Ensure the final slice has the exact requested size.
	// This should always pass if the calculations above are correct.
	if len(alignedSlice) != size {
		// Panic indicates an internal logic error in the alignment calculation. This is not
		// an expected runtime condition and should not occur under normal circumstances.
		panic("AlignedFloat64: internal error - calculated slice length mismatch")
	}

	return alignedSlice
}

// AlignedComplex128 returns a complex128 slice of the requested size. Complex128 values
// are naturally 16 bytes, which matches the common SIMD alignment requirement. Therefore,
// standard allocation via make() is typically sufficient. This function primarily handles
// the size=0 case and returns a standard slice, relying on the Go runtime allocator's
// default behavior for complex128 alignment. If size is 0, it returns nil.
func AlignedComplex128(size int) []complex128 {
	if size == 0 {
		return nil
	}

	// Standard make() is used as complex128 naturally aligns to 16 bytes.
	return make([]complex128, size)
}

// AlignedFloat32 returns a float32 slice of the requested size. The underlying data
// buffer's starting address is guaranteed to be aligned to the package's 'alignment
// constant (16 bytes). If size is 0, it returns nil.
func AlignedFloat32(size int) []float32 {
	if size == 0 {
		return nil
	}

	// Calculate total buffer size: 'size' elements + padding for alignment.
	// float32 is 4 bytes. Need up to (alignment/4 - 1) extra elements.
	totalSize := size + (alignment/int(unsafe.Sizeof(float32(0))) - 1)
	rawBuffer := make([]float32, totalSize)

	// Get the start address of the raw buffer.
	startPtr := uintptr(unsafe.Pointer(&rawBuffer[0]))

	// Calculate the next aligned address within the buffer.
	alignedPtr := (startPtr + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1)

	// Calculate the element offset to the aligned address.
	offset := (alignedPtr - startPtr) / uintptr(unsafe.Sizeof(float32(0)))

	// Create the aligned slice with the requested size.
	alignedSlice := rawBuffer[offset : offset+uintptr(size)]

	// Sanity check: Ensure the final slice has the exact requested size.
	// This should always pass if the calculations above are correct.
	if len(alignedSlice) != size {
		// Panic indicates an internal logic error in the alignment calculation. This is not
		// an expected runtime condition and should not occur under normal circumstances.
		panic("AlignedFloat32: internal error - calculated slice length mismatch")
	}

	// Note: This function guarantees the *starting address* alignment.
	// Some SIMD operations might also require the *length* of the slice to be
	// a multiple of the vector size (e.g., multiple of 4 for 128-bit vectors of
	// float32). If length padding is needed, it must be handled separately by the
	// caller or by modifying this function to allocate even more space initially.

	return alignedSlice
}

// AlignedInt32 returns an int32 slice of the requested size. The underlying
// data buffer's starting address is guaranteed to be aligned to the package's
// 'alignment' constant (16 bytes). If size is 0, it returns nil.
func AlignedInt32(size int) []int32 {
	if size == 0 {
		return nil
	}

	// Calculate total buffer size: 'size' elements + padding for alignment.
	// int32 is 4 bytes. Need up to (alignment/4 - 1) extra elements.
	totalSize := size + (alignment/int(unsafe.Sizeof(int32(0))) - 1)
	rawBuffer := make([]int32, totalSize)

	// Get the start address of the raw buffer.
	startPtr := uintptr(unsafe.Pointer(&rawBuffer[0]))

	// Calculate the next aligned address within the buffer.
	alignedPtr := (startPtr + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1)

	// Calculate the element offset to the aligned address.
	offset := (alignedPtr - startPtr) / uintptr(unsafe.Sizeof(int32(0)))

	// Create the aligned slice with the requested size.
	alignedSlice := rawBuffer[offset : offset+uintptr(size)]

	// Sanity check: Ensure the final slice has the exact requested size.
	// This should always pass if the calculations above are correct.
	if len(alignedSlice) != size {
		// Panic indicates an internal logic error in the alignment calculation. This is not
		// an expected runtime condition and should not occur under normal circumstances.
		panic("AlignedInt32: internal error - calculated slice length mismatch")
	}

	// Note: Similar to AlignedFloat32, this guarantees start address alignment.
	// Length padding (e.g., to a multiple of 4) might be needed separately
	// depending on the specific SIMD usage.

	return alignedSlice
}
