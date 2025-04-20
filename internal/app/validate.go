package app

import "github.com/go-playground/validator/v10"

type aDataValidator struct {
	validator *validator.Validate
}

var av aDataValidator

func init() {
	av.validator = validator.New()

	// Register custom validation functions here.
	// See: https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Custom_Validation_Functions
}

func GetValidator() *validator.Validate {
	return av.validator
}
