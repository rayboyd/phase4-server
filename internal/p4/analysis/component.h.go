// SPDX-License-Identifier: Apache-2.0
package analysis

type Component interface {
	Process(in []int32)
	Close() error
}
