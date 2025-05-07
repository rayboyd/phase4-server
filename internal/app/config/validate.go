// SPDX-License-Identifier: Apache-2.0
package config

import "github.com/go-playground/validator/v10"

func init() {
	av.validator = validator.New()

	// Register custom validation functions here.
	// See: https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Custom_Validation_Functions
}

func GetValidator() *validator.Validate {
	return av.validator
}
