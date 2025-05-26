// SPDX-License-Identifier: Apache-2.0
package buffer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoubleBuffer_Basic(t *testing.T) {
	buffer1 := make([]float32, 10)
	buffer2 := make([]float32, 10)

	db := New(buffer1, buffer2)

	initialBuffer := db.Get()
	assert.Equal(t, buffer1, initialBuffer)

	testValues := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	db.Swap(func(buffer *[]float32) {
		copy(*buffer, testValues)
	})

	updatedBuffer := db.Get()
	assert.Equal(t, testValues, updatedBuffer)

	newValues := []float32{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	db.Swap(func(buffer *[]float32) {
		copy(*buffer, newValues)
	})

	finalBuffer := db.Get()
	assert.Equal(t, newValues, finalBuffer)
}

func TestDoubleBuffer_ForceGet(t *testing.T) {
	buffer1 := make([]float32, 5)
	buffer2 := make([]float32, 5)

	db := New(buffer1, buffer2)

	values := []float32{1, 2, 3, 4, 5}
	db.Swap(func(buffer *[]float32) {
		copy(*buffer, values)
	})

	db.ForceGet(func(buffer []float32) {
		assert.Equal(t, float32(1), buffer[0])
		assert.Equal(t, float32(2), buffer[1])
		assert.Equal(t, float32(3), buffer[2])
		assert.Equal(t, float32(4), buffer[3])
		assert.Equal(t, float32(5), buffer[4])
	})
}

func TestDoubleBuffer_ConcurrentAccess(t *testing.T) {
	buffer1 := make([]float32, 100)
	buffer2 := make([]float32, 100)

	db := New(buffer1, buffer2)

	const iterations = 1000
	const readers = 5

	// Synchronize test completion.
	var wg sync.WaitGroup
	wg.Add(readers + 1) // +1 for the writer.
	errorChan := make(chan string, readers)

	// Writer goroutines.
	go func() {
		defer wg.Done()

		for i := range iterations {
			value := float32(i + 1) // Non-zero values.

			db.Swap(func(buffer *[]float32) {
				// Fills entire buffer with same value to check consistency.
				for j := range *buffer {
					(*buffer)[j] = value
				}
			})

			// Sleep to simulate work and allow readers to run (catchup).
			if i%100 == 0 {
				time.Sleep(time.Microsecond)
			}
		}
	}()

	// Reader goroutines.
	for r := range readers {
		go func(readerID int) {
			defer wg.Done()

			for i := range iterations * 5 {
				buffer := db.Get()

				// Check that all values in the buffer are the same,
				// (we filled the entire buffer with the same value).
				if len(buffer) > 0 {
					first := buffer[0]
					for j := 1; j < len(buffer); j++ {
						if buffer[j] != first {
							errorChan <- fmt.Sprintf(
								"Reader %d: inconsistent buffer values at iteration %d: %f != %f at index %d",
								readerID, i, buffer[j], first, j)
							return
						}
					}
				}

				if i%50 == 0 {
					db.ForceGet(func(buf []float32) {
						_ = buf[0]
					})
				}
			}
		}(r)
	}

	wg.Wait()

	close(errorChan)
	for err := range errorChan {
		t.Fatal(err)
	}
}

func TestDoubleBuffer_NonStructTypes(t *testing.T) {
	// Test with primitive types
	intBuffer := New(42, 0)
	intResult := intBuffer.Get()
	assert.Equal(t, 42, intResult)

	// Test with string
	strBuffer := New("hello", "")
	strResult := strBuffer.Get()
	assert.Equal(t, "hello", strResult)

	// Test with bool
	boolBuffer := New(true, false)
	boolResult := boolBuffer.Get()
	assert.Equal(t, true, boolResult)

	// Test with map
	mapBuffer := New(map[string]int{"key": 100}, nil)
	mapResult := mapBuffer.Get()
	assert.Equal(t, map[string]int{"key": 100}, mapResult)
}

func TestDoubleBuffer_ComplexStructs(t *testing.T) {
	type ComplexData struct {
		Name        string
		FloatData   []float64
		IntData     []int32
		ComplexData []complex128
	}

	data1 := ComplexData{
		FloatData:   []float64{1.1, 2.2},
		IntData:     []int32{1, 2},
		ComplexData: []complex128{complex(1, 2)},
		Name:        "test",
	}

	data2 := ComplexData{}
	db := New(data1, data2)

	result := db.Get()
	assert.Equal(t, data1.Name, result.Name)
	assert.Equal(t, data1.FloatData, result.FloatData)
	assert.Equal(t, data1.IntData, result.IntData)
	assert.Equal(t, data1.ComplexData, result.ComplexData)
}

func TestDoubleBuffer_StructWithUnexportedFields(t *testing.T) {
	type MixedStruct struct {
		ExportedField   string
		unexportedField int // This is unexported.
	}

	data1 := MixedStruct{
		ExportedField:   "test",
		unexportedField: 42, // Will be ignored in copy.
	}
	data2 := MixedStruct{}

	db := New(data1, data2)
	result := db.Get()

	assert.Equal(t, "test", result.ExportedField)
}

func TestDoubleBuffer_StructWithIntSlice(t *testing.T) {
	type IntSliceStruct struct {
		Name     string
		IntSlice []int
	}

	data1 := IntSliceStruct{
		IntSlice: []int{1, 2, 3, 4, 5},
		Name:     "int-slice-test",
	}
	data2 := IntSliceStruct{}

	db := New(data1, data2)
	result := db.Get()

	assert.Equal(t, []int{1, 2, 3, 4, 5}, result.IntSlice)
	assert.Equal(t, "int-slice-test", result.Name)
}

func TestDoubleBuffer_NonSliceTypes(t *testing.T) {
	type AudioData struct {
		Samples   []float32
		Timestamp int64
		Channel   int
	}

	data1 := AudioData{
		Samples:   make([]float32, 10),
		Timestamp: 100,
		Channel:   1,
	}
	data2 := AudioData{
		Samples:   make([]float32, 10),
		Timestamp: 100,
		Channel:   2,
	}

	db := New(data1, data2)

	initial := db.Get()
	assert.Equal(t, data1, initial)

	db.Swap(func(data *AudioData) {
		data.Timestamp = 200
		data.Channel = 3
		for i := range data.Samples {
			data.Samples[i] = float32(i + 1)
		}
	})

	updated := db.Get()
	assert.Equal(t, int64(200), updated.Timestamp)
	assert.Equal(t, 3, updated.Channel)

	db.Swap(func(data *AudioData) {
		data.Timestamp = 300
	})

	final := db.Get()
	assert.Equal(t, int64(300), final.Timestamp)
}

func TestDoubleBuffer_AllNilSlices(t *testing.T) {
	// Test nil float32 slices
	var nilFloat32 []float32
	float32Buffer := New(nilFloat32, nilFloat32)
	float32Result := float32Buffer.Get()
	assert.Nil(t, float32Result)

	// Test nil float64 slices (already has one test, but make sure we have coverage)
	var nilFloat64 []float64
	float64Buffer := New(nilFloat64, nilFloat64)
	float64Result := float64Buffer.Get()
	assert.Nil(t, float64Result)

	// Test nil int32 slices
	var nilInt32 []int32
	int32Buffer := New(nilInt32, nilInt32)
	int32Result := int32Buffer.Get()
	assert.Nil(t, int32Result)

	// Test nil complex128 slices
	var nilComplex []complex128
	complexBuffer := New(nilComplex, nilComplex)
	complexResult := complexBuffer.Get()
	assert.Nil(t, complexResult)
}

func TestDoubleBuffer_AllSliceTypes(t *testing.T) {
	// Test float32 (already covered in other tests)
	float32Buffer := New(make([]float32, 5), make([]float32, 5))
	float32Buffer.Swap(func(buffer *[]float32) {
		(*buffer)[0] = 1.0
	})
	float32Result := float32Buffer.Get()
	assert.Equal(t, float32(1.0), float32Result[0])

	// Test float64
	float64Buffer := New(make([]float64, 5), make([]float64, 5))
	float64Buffer.Swap(func(buffer *[]float64) {
		(*buffer)[0] = 2.0
	})
	float64Result := float64Buffer.Get()
	assert.Equal(t, float64(2.0), float64Result[0])

	// Test int32
	int32Buffer := New(make([]int32, 5), make([]int32, 5))
	int32Buffer.Swap(func(buffer *[]int32) {
		(*buffer)[0] = 3
	})
	int32Result := int32Buffer.Get()
	assert.Equal(t, int32(3), int32Result[0])

	// Test complex128
	complexBuffer := New(make([]complex128, 5), make([]complex128, 5))
	complexBuffer.Swap(func(buffer *[]complex128) {
		(*buffer)[0] = complex(4, 5)
	})
	complexResult := complexBuffer.Get()
	assert.Equal(t, complex(4, 5), complexResult[0])
}
