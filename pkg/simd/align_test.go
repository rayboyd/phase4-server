// SPDX-License-Identifier: Apache-2.0
package simd

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestAlignedFloat64(t *testing.T) {
	testSizes := []int{0, 1, 7, 8, 15, 16, 100, 1024}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			slice := AlignedFloat64(size)

			if size == 0 {
				assert.Nil(t, slice, "AlignedFloat64(0) should return nil")
				return // Skip further checks for size 0
			}

			assert.NotNil(t, slice, "Slice should not be nil for size > 0")
			assert.Equal(t, size, len(slice), "Length mismatch")
			assert.GreaterOrEqual(t, cap(slice), size, "Capacity should be >= size")

			t.Logf("Checking alignment for size %d\n", size)
			if len(slice) > 0 {
				ptr := uintptr(unsafe.Pointer(&slice[0]))
				assert.Zero(t, ptr%uintptr(alignment), "Address %p is not aligned to %d bytes", &slice[0], alignment)
			}
		})
	}
}

func TestAlignedComplex128(t *testing.T) {
	testSizes := []int{0, 1, 7, 8, 15, 16, 100, 1024}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			slice := AlignedComplex128(size)

			if size == 0 {
				assert.Nil(t, slice, "AlignedComplex128(0) should return nil")
				return
			}

			assert.NotNil(t, slice, "Slice should not be nil for size > 0")
			assert.Equal(t, size, len(slice), "Length mismatch")
			assert.GreaterOrEqual(t, cap(slice), size, "Capacity should be >= size")

			// Check alignment (complex128 is naturally 16-byte aligned)
			t.Logf("Checking alignment for size %d\n", size)
			if len(slice) > 0 {
				ptr := uintptr(unsafe.Pointer(&slice[0]))
				expectedAlignment := unsafe.Alignof(complex128(0))

				// Check if the pointer is aligned to the *maximum* of the natural alignment and our constant
				requiredAlignment := uintptr(alignment)
				if expectedAlignment > requiredAlignment {
					requiredAlignment = expectedAlignment
				}
				assert.Zero(t, ptr%requiredAlignment, "Address %p is not aligned to %d bytes (natural %d)", &slice[0], requiredAlignment, expectedAlignment)
			}
		})
	}
}

func TestAlignedFloat32(t *testing.T) {
	testSizes := []int{0, 1, 3, 4, 7, 8, 15, 16, 100, 1024}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			slice := AlignedFloat32(size)

			if size == 0 {
				assert.Nil(t, slice, "AlignedFloat32(0) should return nil")
				return
			}

			assert.NotNil(t, slice, "Slice should not be nil for size > 0")
			assert.Equal(t, size, len(slice), "Length mismatch")
			assert.GreaterOrEqual(t, cap(slice), size, "Capacity should be >= size")

			t.Logf("Checking alignment for size %d\n", size)
			if len(slice) > 0 {
				ptr := uintptr(unsafe.Pointer(&slice[0]))
				assert.Zero(t, ptr%uintptr(alignment), "Address %p is not aligned to %d bytes", &slice[0], alignment)
			}
		})
	}
}

func TestAlignedInt32(t *testing.T) {
	testSizes := []int{0, 1, 3, 4, 7, 8, 15, 16, 100, 1024}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			slice := AlignedInt32(size)

			if size == 0 {
				assert.Nil(t, slice, "AlignedInt32(0) should return nil")
				return
			}

			assert.NotNil(t, slice, "Slice should not be nil for size > 0")
			assert.Equal(t, size, len(slice), "Length mismatch")
			assert.GreaterOrEqual(t, cap(slice), size, "Capacity should be >= size")

			t.Logf("Checking alignment for size %d\n", size)
			if len(slice) > 0 {
				ptr := uintptr(unsafe.Pointer(&slice[0]))
				assert.Zero(t, ptr%uintptr(alignment), "Address %p is not aligned to %d bytes", &slice[0], alignment)
			}
		})
	}
}
