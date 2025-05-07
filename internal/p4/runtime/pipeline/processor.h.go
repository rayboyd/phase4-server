// SPDX-License-Identifier: Apache-2.0
package pipeline

import (
	"phase4/internal/p4/runtime/stage"
	"sync"
)

var FftDataPool = sync.Pool{
	New: func() any {
		return &stage.FFTData{Magnitudes: make([]float64, 0, 129)}
	},
}

type ProcessorComponent struct {
	system   *stage.System
	routerID string
	stage.BaseActor
}
