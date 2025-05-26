// SPDX-License-Identifier: Apache-2.0
package transport

type Component interface {
	SendData(data []byte) error
	Close() error
}
