// SPDX-License-Identifier: Apache-2.0
/*
Package bitint provides bit manipulation functions optimized for real-time audio
processing. The package focuses on power-of-2 operations commonly needed in FFT
and buffer sizing. NextPowerOfTwo returns the next power of 2 greater than or
equal to size. For powers of 2, it returns the same value. For other values, it
returns the next higher power of 2.

The subtraction (size-1) is critical, without the subtraction, powers of 2 would be
incorrectly doubled.

WITH subtraction (correct):
- For input 8 (already a power of 2):
	size-1 = 7 (binary 0111)
	bits.Len32(7) = 3 (highest bit position is 2^2)
	1 << 3 = 8 (correctly preserves original power of 2)

WITHOUT subtraction (incorrect):
- For input 8 (already a power of 2):
	bits.Len32(8) = 4 (binary 1000 has its highest bit position at 2^3)
	1 << 4 = 16 (incorrectly doubles the input)

This ensures we get exactly the right shift amount to return the same value for
powers of 2, and the next power of 2 for all other values.
*/
package bitint

import "math/bits"

// NextPowerOfTwo returns the next power of 2 >= size.
func NextPowerOfTwo(size int) int {
	if size <= 0 {
		return 1
	}
	// 64-bit platforms (where int is 64-bit).
	if ^uint(0)>>63 == 0 {
		return int(1 << (bits.Len64(uint64(size - 1))))
	}
	// 32-bit platforms.
	return int(1 << (bits.Len32(uint32(size - 1))))
}

// For 32-bit integers.
func NextPowerOfTwo32(size int32) int32 {
	if size <= 0 {
		return 1
	}
	return int32(1 << (bits.Len32(uint32(size - 1))))
}

// For 64-bit integers.
func NextPowerOfTwo64(size int64) int64 {
	if size <= 0 {
		return 1
	}
	return int64(1 << (bits.Len64(uint64(size - 1))))
}

// IsPowerOfTwo checks if n is a power of 2.
func IsPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// For 32-bit integers.
func IsPowerOfTwo32(n int32) bool {
	return n > 0 && (n&(n-1)) == 0
}

// For 64-bit integers.
func IsPowerOfTwo64(n int64) bool {
	return n > 0 && (n&(n-1)) == 0
}
