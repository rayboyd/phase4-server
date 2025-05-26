// SPDX-License-Identifier: Apache-2.0
package analysis

import (
	"fmt"
	"strings"

	"gonum.org/v1/gonum/dsp/window"
)

// ParseWindowFunc converts a string name (case-insensitive) to a WindowFunc
// enum, returns a known default (Hann) and an error if the name is unknown.
func ParseWindowFunc(name string) (WindowFunc, error) {
	switch strings.ToLower(name) {
	case "bartletthann":
		return BartlettHann, nil
	case "blackman":
		return Blackman, nil
	case "blackmannuttall":
		return BlackmanNuttall, nil
	case "hann", "hanning":
		return Hann, nil
	case "hamming":
		return Hamming, nil
	case "lanczos":
		return Lanczos, nil
	case "nuttall":
		return Nuttall, nil
	default:
		return Hann, fmt.Errorf("unknown window function name: '%s'", name)
	}
}

func applyWindowFunc(coeffs []float64, windowType WindowFunc) {
	// Ensure coeffs is not nil or empty before proceeding.
	if len(coeffs) == 0 {
		return
	}

	// Initialize coeffs to 1.0 before applying the window in place, necessary because
	// gonum window functions multiply in place.
	for i := range coeffs {
		coeffs[i] = 1.0
	}

	switch windowType {
	case BartlettHann:
		window.BartlettHann(coeffs)
	case Blackman:
		window.Blackman(coeffs)
	case BlackmanNuttall:
		window.BlackmanNuttall(coeffs)
	case Hann:
		window.Hann(coeffs)
	case Hamming:
		window.Hamming(coeffs)
	case Lanczos:
		window.Lanczos(coeffs)
	case Nuttall:
		window.Nuttall(coeffs)
	default:
		window.Hann(coeffs)
	}
}
