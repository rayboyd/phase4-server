// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"fmt"
	"log"
	"math"
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

	magnitudeSize := size/2 + 1

	// Pre-compute frequency bins with aligned memory
	frequencyBins := simd.AlignedFloat64(magnitudeSize)
	frequencyResolution := sampleRate / float64(size)
	for i := 0; i < magnitudeSize; i++ {
		frequencyBins[i] = float64(i) * frequencyResolution
	}

	// Create all buffers with SIMD alignment
	magnitudeBuffer1 := simd.AlignedFloat64(magnitudeSize)
	magnitudeBuffer2 := simd.AlignedFloat64(magnitudeSize)
	prevMagnitudes := simd.AlignedFloat64(magnitudeSize)
	spectralFlux := simd.AlignedFloat64(magnitudeSize)

	p := &FFTProcessor{
		fftSize:        size,
		fftFunc:        fftFunc,
		sampleRate:     sampleRate,
		inputBuffer:    simd.AlignedFloat64(size),
		fftOutput:      simd.AlignedComplex128(magnitudeSize),
		magnitudes:     buffer.NewFloat64DoubleBuffer(magnitudeBuffer1, magnitudeBuffer2),
		normFactor:     1.0 / float64(0x80000000), // Converts int32 to float64 range [-1,1).
		window:         windowCoeffs,
		fftInputScale:  1.0 / float64(size),
		frequencyBins:  frequencyBins,
		prevMagnitudes: prevMagnitudes,
		spectralFlux:   spectralFlux,
		debugInterval:  100, // Log every 100 frames (~0.58 seconds at 44.1kHz/256)
	}

	log.Printf("FFT Processor initialized: size=%d, sampleRate=%.0f, bins=%d, resolution=%.2f Hz/bin",
		size, sampleRate, magnitudeSize, frequencyResolution)

	return p, nil
}

func (p *FFTProcessor) Process(inputBuffer []int32) {
	inputLen := len(inputBuffer)
	magnitudeSize := len(p.frequencyBins)

	// Calculate input RMS for debugging
	var inputRMS float64

	// Use direct array indexing instead of range loop for better bounds check elimination
	for i := 0; i < p.fftSize; i++ {
		if i < inputLen {
			normalized := float64(inputBuffer[i]) * p.normFactor
			inputRMS += normalized * normalized
			p.inputBuffer[i] = normalized * p.window[i]
		} else {
			p.inputBuffer[i] = 0.0
		}
	}
	inputRMS = math.Sqrt(inputRMS / float64(p.fftSize))

	p.fftFunc.Coefficients(p.fftOutput, p.inputBuffer)

	var totalFlux float64
	var maxFlux float64
	var bassEnergy float64

	p.magnitudes.Swap(func(currentMagBuffer *[]float64) {
		// Direct indexing for better performance
		for i := 0; i < magnitudeSize; i++ {
			mag := cmplx.Abs(p.fftOutput[i]) * p.fftInputScale

			// Single-sided spectrum energy compensation
			if i > 0 && i < p.fftSize/2 {
				(*currentMagBuffer)[i] = mag * 2.0
			} else {
				(*currentMagBuffer)[i] = mag
			}

			// Track bass energy (0-200Hz)
			if p.frequencyBins[i] < 200 {
				bassEnergy += (*currentMagBuffer)[i]
			}

			// Calculate spectral flux with emphasis on low frequencies
			weight := 1.0
			if p.frequencyBins[i] < 200 {
				weight = 2.0 // Double weight for bass frequencies
			}

			diff := ((*currentMagBuffer)[i] - p.prevMagnitudes[i]) * weight
			if diff > 0 {
				p.spectralFlux[i] = diff
				totalFlux += diff
				if diff > maxFlux {
					maxFlux = diff
				}
			} else {
				p.spectralFlux[i] = 0.0
			}

			// Update previous magnitudes for next frame
			p.prevMagnitudes[i] = (*currentMagBuffer)[i]
		}
	})

	// Debug logging
	frameCount := p.frameCounter.Add(1)
	if frameCount%uint64(p.debugInterval) == 0 {
		// Uncomment if debug logging is needed
		// bassFlux := p.GetSpectralFluxInRange(20, 200)
		// midFlux := p.GetSpectralFluxInRange(200, 2000)
		// highFlux := p.GetSpectralFluxInRange(2000, 20000)
		// log.Printf("FFT Debug [frame %d]: inputRMS=%.4f, bassEnergy=%.4f, totalFlux=%.4f, maxFlux=%.4f",
		//     frameCount, inputRMS, bassEnergy, totalFlux, maxFlux)
	}
}

// GetSpectralFluxInRange returns spectral flux sum for a frequency range
// Optimized to avoid allocations and use direct array access
func (p *FFTProcessor) GetSpectralFluxInRange(lowFreq, highFreq float64) float64 {
	var sum float64
	magnitudeSize := len(p.frequencyBins)

	for i := 0; i < magnitudeSize; i++ {
		freq := p.frequencyBins[i]
		if freq >= lowFreq && freq <= highFreq {
			sum += p.spectralFlux[i]
		}
		if freq > highFreq {
			break // Early exit if frequency exceeds highFreq
		}
	}
	return sum
}

// FindPeakFrequency returns the frequency bin with the highest magnitude
// Optimized for better performance with direct array access
func (p *FFTProcessor) FindPeakFrequency() (freq float64, magnitude float64) {
	magnitudes := p.GetMagnitudes()
	maxMag := 0.0
	maxIdx := 0
	magnitudeSize := len(magnitudes)

	for i := 0; i < magnitudeSize; i++ {
		if magnitudes[i] > maxMag {
			maxMag = magnitudes[i]
			maxIdx = i
		}
	}

	return p.frequencyBins[maxIdx], maxMag
}

// ValidateFFT tests the FFT with a known sine wave
// Optimized with direct array access
func (p *FFTProcessor) ValidateFFT(testFreq float64) (detectedFreq float64, error float64) {
	// Generate a pure sine wave at testFreq
	for i := 0; i < len(p.inputBuffer); i++ {
		t := float64(i) / p.sampleRate
		p.inputBuffer[i] = math.Sin(2*math.Pi*testFreq*t) * p.window[i] * p.fftInputScale
	}

	p.fftFunc.Coefficients(p.fftOutput, p.inputBuffer)

	// Find peak
	maxMag := 0.0
	maxIdx := 0
	magnitudeSize := len(p.fftOutput)

	for i := 0; i < magnitudeSize; i++ {
		mag := cmplx.Abs(p.fftOutput[i])
		if mag > maxMag {
			maxMag = mag
			maxIdx = i
		}
	}

	detectedFreq = p.frequencyBins[maxIdx]
	error = math.Abs(detectedFreq - testFreq)
	return detectedFreq, error
}

func (p *FFTProcessor) GetMagnitudes() []float64 {
	return p.magnitudes.Get()
}

func (p *FFTProcessor) GetFrequencyBins() []float64 {
	return p.frequencyBins
}

func (p *FFTProcessor) GetSpectralFlux() []float64 {
	return p.spectralFlux
}

func (p *FFTProcessor) GetFrequencyResolution() float64 {
	return p.sampleRate / float64(p.fftSize)
}

func (p *FFTProcessor) Close() error {
	// Clean up resources if needed.
	return nil
}
