// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"math"
	"phase4/pkg/simd"
	"sort"
)

func NewBPMDetector(sampleRate float64, framesPerBuffer int) *BPMDetector {
	const (
		onsetBufferSize  = 1024
		onsetTimesSize   = 1024
		recentWindowSize = 20
	)

	return &BPMDetector{
		sampleRate:       sampleRate,
		framesPerBuffer:  framesPerBuffer,
		onsetThreshold:   0.1,
		onsetBuffer:      simd.AlignedFloat64(onsetBufferSize),
		onsetTimes:       simd.AlignedFloat64(onsetTimesSize),
		recentBuffer:     simd.AlignedFloat64(recentWindowSize),
		validOnsets:      simd.AlignedFloat64(onsetTimesSize),
		intervals:        simd.AlignedFloat64(onsetTimesSize),
		histogramBins:    make(map[int]int),
		onsetBufferLen:   0,
		onsetTimesLen:    0,
		binCounts:        make([]binCount, 0, 100),
		bpmCandidates:    make([]float64, 0, 20),
		scoredCandidates: make([]scoredBPM, 0, 20),
	}
}

// ProcessFlux analyzes spectral flux for onset detection and BPM calculation
func (bd *BPMDetector) ProcessFlux(flux []float64, frameCount uint64) {
	// Calculate total flux and peak flux from the first 10 bins, this helps
	// reduce noise and emphasizes the most significant spectral changes.
	// Optimize by limiting loop and bounds check.
	totalFlux, peakFlux := 0.0, 0.0
	for i, v := range flux {
		if i >= 10 {
			break
		}
		totalFlux += v
		if v > peakFlux {
			peakFlux = v
		}
	}

	bd.mu.Lock()
	defer bd.mu.Unlock()

	// Update recent buffer with the latest flux value
	if bd.onsetBufferLen < len(bd.onsetBuffer) {
		bd.onsetBuffer[bd.onsetBufferLen] = totalFlux
		bd.onsetBufferLen++
	} else {
		// If buffer is full, shift left by one and add new value, this keeps
		// the buffer size constant and allows for continuous analysis.
		copy(bd.onsetBuffer, bd.onsetBuffer[1:])
		bd.onsetBuffer[bd.onsetBufferLen-1] = totalFlux
	}

	if bd.onsetBufferLen > 20 {
		// Use a fixed window size for statistics, this can be tuned
		// to adapt to different music styles, e.g. 20 for breakbeats.
		windowSize := 20

		// Calculate mean and standard deviation of the last `windowSize` values

		mean := 0.0
		for i := 0; i < windowSize; i++ {
			mean += bd.onsetBuffer[bd.onsetBufferLen-windowSize+i]
		}
		mean /= float64(windowSize)

		variance := 0.0
		for i := 0; i < windowSize; i++ {
			diff := bd.onsetBuffer[bd.onsetBufferLen-windowSize+i] - mean
			variance += diff * diff
		}
		stdDev := math.Sqrt(variance / float64(windowSize))

		// Dynamic threshold based on statistics.
		threshold := max(mean+1.5*stdDev, bd.onsetThreshold)

		// Check if current flux is a peak.
		current := bd.onsetBuffer[bd.onsetBufferLen-1]
		previous := bd.onsetBuffer[bd.onsetBufferLen-2]

		// Peak detection: current > threshold AND current > previous.
		if current > threshold && current > previous*1.3 {
			timeInSeconds := float64(frameCount) * float64(bd.framesPerBuffer) / bd.sampleRate

			// Prevent double-triggers (minimum 100ms between onsets).
			if bd.onsetTimesLen == 0 || timeInSeconds-bd.onsetTimes[bd.onsetTimesLen-1] > 0.1 {
				if bd.onsetTimesLen < len(bd.onsetTimes) {
					bd.onsetTimes[bd.onsetTimesLen] = timeInSeconds
					bd.onsetTimesLen++
				} else {
					// Edge: onset times buffer is full, shift left by one.
					copy(bd.onsetTimes, bd.onsetTimes[1:])
					bd.onsetTimes[bd.onsetTimesLen-1] = timeInSeconds
				}

				// Keep only recent onsets (last 10 seconds)
				validCount := 0
				cutoffTime := timeInSeconds - 10.0

				for i := 0; i < bd.onsetTimesLen; i++ {
					if bd.onsetTimes[i] > cutoffTime {
						bd.validOnsets[validCount] = bd.onsetTimes[i]
						validCount++
					}
				}

				if validCount < bd.onsetTimesLen {
					copy(bd.onsetTimes, bd.validOnsets[:validCount]) // Update the onsetTimes buffer.
					bd.onsetTimesLen = validCount
				}

				if bd.onsetTimesLen >= 4 {
					bd.calculateBPM()
				}
			}
		}
	}
}

func (bd *BPMDetector) calculateBPM() {
	if bd.onsetTimesLen < 4 {
		return
	}

	// Calculate inter-onset intervals.
	intervalCount := 0
	for i := 1; i < bd.onsetTimesLen; i++ {
		interval := bd.onsetTimes[i] - bd.onsetTimes[i-1]
		if interval > 0.2 && interval < 2.0 { // 30-300 BPM range
			bd.intervals[intervalCount] = interval
			intervalCount++
		}
	}
	if intervalCount < 3 {
		return
	}

	// Create histogram of intervals to find clusters. This handles breakbeats better
	// by identifying recurring patterns.
	for k := range bd.histogramBins {
		delete(bd.histogramBins, k)
	}

	// Initialize histogram bins for 0.5 BPM resolution (200 bins for 0-100 BPM).
	for i := 0; i < intervalCount; i++ {
		bin := int(bd.intervals[i] * 200)
		bd.histogramBins[bin]++
	}

	// Reset binCounts slice to reuse memory.
	bd.binCounts = bd.binCounts[:0]
	for bin, count := range bd.histogramBins {
		bd.binCounts = append(bd.binCounts, binCount{bin: bin, count: count})
	}

	// Sort binCounts by count in descending order to find the most common intervals.
	sort.Slice(bd.binCounts, func(i, j int) bool {
		return bd.binCounts[i].count > bd.binCounts[j].count
	})

	// Try different interpretations based on the most common intervals.
	bd.bpmCandidates = bd.bpmCandidates[:0]
	maxBins := min(len(bd.binCounts), 3) // Top 3 most common intervals.

	for i := 0; i < maxBins; i++ {
		interval := float64(bd.binCounts[i].bin) / 200.0
		if interval > 0 {
			// Add the base tempo and related tempos.
			baseBPM := 60.0 / interval
			bd.bpmCandidates = append(bd.bpmCandidates, baseBPM)

			// For dance music, half-tempo often works better - but not for drum & bass range.
			if baseBPM > 130 && (baseBPM < 160 || baseBPM > 180) {
				bd.bpmCandidates = append(bd.bpmCandidates, baseBPM/2)
			}

			// Add double tempo for slower rhythms.
			if baseBPM < 80 {
				bd.bpmCandidates = append(bd.bpmCandidates, baseBPM*2)
			}

			// Special case for tempos near 85 BPM - check if it's actually half of "breaks" tempo.
			if baseBPM >= 80 && baseBPM <= 90 {
				doubleTempo := baseBPM * 2
				if doubleTempo >= 160 && doubleTempo <= 180 {
					bd.bpmCandidates = append(bd.bpmCandidates, doubleTempo)
				}
			}
		}
	}

	// Add whole beat analysis for breakbeats. Calculate average of all intervals.
	avgInterval := 0.0
	for i := 0; i < intervalCount; i++ {
		avgInterval += bd.intervals[i]
	}
	avgInterval /= float64(intervalCount)
	rawBPM := 60.0 / avgInterval
	bd.bpmCandidates = append(bd.bpmCandidates, rawBPM)

	// Simple rounding to musically useful tempos. Round to nearest 0.5 BPM.
	for i := range bd.bpmCandidates {
		bd.bpmCandidates[i] = math.Round(bd.bpmCandidates[i]*2) / 2
	}

	// Score each candidate BPM based on alignment with intervals.
	bd.scoredCandidates = bd.scoredCandidates[:0]

	// For breakbeats, emphasize stability and typical tempo ranges.
	for _, candidateBPM := range bd.bpmCandidates {
		if candidateBPM < 60 || candidateBPM > 200 {
			continue
		}

		// Calculate expected interval for this BPM. This is the ideal beat interval
		// that we expect to see in the intervals.
		expectedInterval := 60.0 / candidateBPM

		// Calculate grid alignment score
		alignmentScore := 0.0
		totalWeight := 0.0

		// FML. Breakbeats often have irregular intervals ...
		for i := 0; i < intervalCount; i++ {
			// Find closest grid position (including half/double).
			// For breakbeats, we check if interval fits 1/4, 1/3, 1/2, 1, 2 times the beat.
			possibleGrids := []float64{0.25, 0.33, 0.5, 1.0, 2.0}
			bestError := math.MaxFloat64

			// Check each grid position for alignment.
			// This allows us to find the best fit for breakbeats, which often have
			// irregular intervals that can be 1/4, 1/3, or 1/2 of the expected beat.
			for _, grid := range possibleGrids {
				gridPos := expectedInterval * grid
				error := math.Abs(bd.intervals[i]-gridPos) / gridPos
				if error < bestError {
					bestError = error
				}
			}

			// Weight inversely by error (closer = higher weight).
			weight := 1.0 / (1.0 + bestError*10)
			alignmentScore += weight
			totalWeight += 1.0
		}

		// Normalize alignment score by total weight.
		if totalWeight > 0 {
			alignmentScore /= totalWeight
		}

		// Prefer certain BPM ranges for breakbeats (90-110).
		rangeBonus := 1.0
		if candidateBPM >= 90 && candidateBPM <= 110 {
			rangeBonus = 1.3 // 30% bonus for breakbeat range.
		} else if candidateBPM >= 160 && candidateBPM <= 180 {
			rangeBonus = 1.4 // 40% bonus for drum & bass range.
		} else if candidateBPM >= 120 && candidateBPM <= 140 {
			rangeBonus = 1.2 // 20% bonus for house/techno range.
		}

		// Apply hysteresis bonus for stability.
		stabilityBonus := 1.0
		if bd.currentBPM > 0 {
			relativeDiff := math.Abs(candidateBPM-bd.currentBPM) / bd.currentBPM
			if relativeDiff < 0.05 { // Within 5% of current BPM.
				stabilityBonus = 1.2 // 20% bonus for stability.
			}
		}

		// Final score is a combination of alignment, range, and stability.
		// This allows us to prefer candidates that are close to the current BPM
		// and have a good alignment with the detected intervals.
		finalScore := alignmentScore * rangeBonus * stabilityBonus

		bd.scoredCandidates = append(bd.scoredCandidates, scoredBPM{
			bpm:   candidateBPM,
			score: finalScore,
		})
	}

	// Remove duplicates by rounding to nearest 0.5 BPM.
	uniqueCandidates := make(map[float64]scoredBPM)
	for _, candidate := range bd.scoredCandidates {
		roundedBPM := math.Round(candidate.bpm*2) / 2
		if existing, ok := uniqueCandidates[roundedBPM]; !ok || candidate.score > existing.score {
			uniqueCandidates[roundedBPM] = candidate
		}
	}

	// Convert back to slice and sort.
	bd.scoredCandidates = bd.scoredCandidates[:0]
	for _, candidate := range uniqueCandidates {
		bd.scoredCandidates = append(bd.scoredCandidates, candidate)
	}

	// Sort candidates by score (descending).
	sort.Slice(bd.scoredCandidates, func(i, j int) bool {
		return bd.scoredCandidates[i].score > bd.scoredCandidates[j].score
	})

	// Simply use the highest-scoring candidate.
	if len(bd.scoredCandidates) > 0 {
		bestCandidate := bd.scoredCandidates[0]

		// Calculate standard deviation for confidence.
		stdDev := 0.0
		for i := 0; i < intervalCount; i++ {
			diff := bd.intervals[i] - avgInterval
			stdDev += diff * diff
		}
		stdDev = math.Sqrt(stdDev / float64(intervalCount))

		// Use coefficient of variation as confidence measure.
		// Lower relative variation = higher confidence.
		confidenceScore := math.Max(0.1, math.Min(1.0, 1.0/(1.0+stdDev/avgInterval*5)))

		// If we have a strong confidence, update the BPM.
		bd.currentBPM = bestCandidate.bpm
		bd.confidence = confidenceScore * bestCandidate.score
	}

	// Log with more precision
	// log.Printf("BPM Detected: %.1f BPM (confidence: %.2f, onsets: %d, intervals: %v)",
	// 	bd.currentBPM, bd.confidence, bd.onsetTimesLen, bd.intervals[:intervalCount])
}

func (bd *BPMDetector) GetBPM() (bpm float64, confidence float64) {
	bd.mu.RLock()
	defer bd.mu.RUnlock()
	return bd.currentBPM, bd.confidence
}

func (bd *BPMDetector) GetOnsetCount() int {
	bd.mu.RLock()
	defer bd.mu.RUnlock()
	return bd.onsetTimesLen
}
