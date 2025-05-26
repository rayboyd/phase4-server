// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"sync"
)

type binCount struct {
	bin   int
	count int
}

type scoredBPM struct {
	bpm   float64
	score float64
}

type BPMDetector struct {
	histogramBins    map[int]int
	validOnsets      []float64
	scoredCandidates []scoredBPM
	bpmCandidates    []float64
	binCounts        []binCount
	intervals        []float64
	onsetBuffer      []float64
	onsetTimes       []float64
	recentBuffer     []float64
	confidence       float64
	onsetBufferLen   int
	onsetTimesLen    int
	sampleRate       float64
	currentBPM       float64
	onsetThreshold   float64
	framesPerBuffer  int
	mu               sync.RWMutex
}
