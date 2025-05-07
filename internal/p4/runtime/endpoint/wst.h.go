// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
)

type WstComponent struct {
	sender transport.Component
	stage.BaseActor
}
