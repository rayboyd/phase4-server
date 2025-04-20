package app

import "github.com/go-playground/validator/v10"

type AppValidator struct {
	validate *validator.Validate
}

var av AppValidator

func init() {
	av.validate = validator.New()

	// Register custom validation functions here.
	// See: https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Custom_Validation_Functions
}

func GetValidator() *validator.Validate {
	return av.validate
}
