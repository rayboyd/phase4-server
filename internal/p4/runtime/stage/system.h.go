// SPDX-License-Identifier: Apache-2.0
package stage

import (
	"context"
	"sync"
)

type System struct {
	ctx    context.Context
	actors map[string]Actor
	cancel context.CancelFunc
	mu     sync.RWMutex
}
