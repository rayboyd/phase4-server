// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"fmt"
	"math/cmplx"
	"phase4/pkg/bitint"
	"phase4/pkg/buffer"
	"phase4/pkg/simd"

	"gonum.org/v1/gonum/dsp/fourier"
)

func NewFFTProcessor(size int, sampleRate float64, windowType WindowFunc) (*FFTProcessor, error) {
	if !bitint.IsPowerOfTwo(size) {
		return nil, fmt.Errorf("fft size must be a power of 2, got %d", size)
	}

	fftFunc := fourier.NewFFT(size)
	windowCoeffs := simd.AlignedFloat64(size)
	applyWindowFunc(windowCoeffs, windowType)

	magnitudeSize := size/2 + 1 // FFTs size for real input is N/2 + 1 complex values.

	// Create initial magnitude buffers - we need two identical buffers.
	magnitudeBuffer1 := simd.AlignedFloat64(magnitudeSize)
	magnitudeBuffer2 := simd.AlignedFloat64(magnitudeSize)

	p := &FFTProcessor{
		fftSize:       size,
		fftFunc:       fftFunc,
		sampleRate:    sampleRate,
		inputBuffer:   simd.AlignedFloat64(size),
		fftOutput:     simd.AlignedComplex128(magnitudeSize),
		magnitudes:    buffer.NewFloat64DoubleBuffer(magnitudeBuffer1, magnitudeBuffer2),
		normFactor:    1.0 / float64(0x80000000), // Converts int32 to float64 range [-1,1).
		window:        windowCoeffs,
		fftInputScale: 1.0 / float64(size),
	}

	return p, nil
}

func (p *FFTProcessor) Process(inputBuffer []int32) {
	inputLen := len(inputBuffer)
	for i := range p.fftSize {
		if i < inputLen {
			p.inputBuffer[i] = float64(inputBuffer[i]) * p.normFactor * p.window[i] * p.fftInputScale
		} else {
			p.inputBuffer[i] = 0.0
		}
	}

	p.fftFunc.Coefficients(p.fftOutput, p.inputBuffer)

	p.magnitudes.Swap(func(currentMagBuffer *[]float64) {
		for i, c := range p.fftOutput {
			mag := cmplx.Abs(c)

			// For a single-sided spectrum from a real input:
			// - DC component (index 0) and Nyquist component (index N/2) are unique.
			// - Other frequency components (0 < index < N/2) have their energy
			//   doubled because their negative frequency counterparts are folded in.
			if i > 0 && i < p.fftSize/2 { // Check if it's not DC and not Nyquist
				(*currentMagBuffer)[i] = mag * 2.0
			} else {
				(*currentMagBuffer)[i] = mag
			}
		}
	})
}

func (p *FFTProcessor) GetMagnitudes() []float64 {
	return p.magnitudes.Get()
}

func (p *FFTProcessor) Close() error {
	// Clean up resources if needed.
	return nil
}
