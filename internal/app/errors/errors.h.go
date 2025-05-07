// SPDX-License-Identifier: Apache-2.0
package errors

type error interface {
	Error() string
}

type FatalError struct {
	Err     error
	Message string
}

type CommandCompleted struct {
	Err     error
	Message string
}
