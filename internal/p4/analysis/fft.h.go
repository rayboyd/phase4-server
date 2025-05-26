// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"phase4/pkg/buffer"
	"sync/atomic"

	"gonum.org/v1/gonum/dsp/fourier"
)

type FFTProcessor struct {
	fftFunc        *fourier.FFT
	magnitudes     *buffer.Float64DoubleBuffer
	prevMagnitudes []float64
	inputBuffer    []float64
	fftOutput      []complex128
	window         []float64
	frequencyBins  []float64
	spectralFlux   []float64
	fftInputScale  float64
	sampleRate     float64
	fftSize        int
	normFactor     float64
	frameCounter   atomic.Uint64
	debugInterval  int
}
