// SPDX-License-Identifier: Apache-2.0
package config

import "github.com/go-playground/validator/v10"

type aDataValidator struct {
	validator *validator.Validate
}

var av aDataValidator
