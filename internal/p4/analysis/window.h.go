// SPDX-License-Identifier: Apache-2.0
package analysis

import "fmt"

type WindowFunc int

const (
	BartlettHann WindowFunc = iota
	Blackman
	BlackmanNuttall
	Hann
	Hamming
	Lanczos
	Nuttall
)

// String returns the string representation of the WindowFunc.
func (w WindowFunc) String() string {
	switch w {
	case BartlettHann:
		return "BartlettHann"
	case Blackman:
		return "Blackman"
	case BlackmanNuttall:
		return "BlackmanNuttall"
	case Hann:
		return "Hann"
	case Hamming:
		return "Hamming"
	case Lanczos:
		return "Lanczos"
	case Nuttall:
		return "Nuttall"
	default:
		// Return a representation for unknown values, useful for testing defaults.
		return fmt.Sprintf("UnknownWindow(%d)", int(w))
	}
}
