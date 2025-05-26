// SPDX-License-Identifier: Apache-2.0
package pipeline

import "phase4/internal/p4/runtime/stage"

type RouterComponent struct {
	system    *stage.System
	targetIDs []string
	stage.BaseActor
}
