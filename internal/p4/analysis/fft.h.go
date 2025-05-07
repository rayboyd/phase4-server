// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"phase4/pkg/buffer"

	"gonum.org/v1/gonum/dsp/fourier"
)

// FFTProcessor implements the analysis.Component interface.
type FFTProcessor struct {
	fftFunc       *fourier.FFT
	inputBuffer   []float64
	magnitudes    *buffer.Float64DoubleBuffer
	window        []float64
	fftOutput     []complex128
	fftSize       int
	sampleRate    float64
	normFactor    float64
	fftInputScale float64
}
