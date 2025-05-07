// SPDX-License-Identifier: Apache-2.0
package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestGetValidator_ReturnsInstance(t *testing.T) {
	instance := GetValidator()
	assert.NotNil(t, instance, "GetValidator() should return a non-nil instance")

	type testStruct struct {
		Value int `validate:"required"`
	}

	err := instance.Struct(testStruct{})
	assert.Error(t, err, "Expected validation error for missing required field")

	var validationErrs validator.ValidationErrors
	assert.ErrorAs(t, err, &validationErrs, "Error should be of type validator.ValidationErrors")
	assert.Len(t, validationErrs, 1, "Expected 1 validation error")
}

func TestGetValidator_IsSingleton(t *testing.T) {
	instance1 := GetValidator()
	instance2 := GetValidator()

	assert.NotNil(t, instance1, "GetValidator() returned nil unexpectedly (instance1)")
	assert.NotNil(t, instance2, "GetValidator() returned nil unexpectedly (instance2)")
	assert.Same(t, instance1, instance2, "Expected GetValidator() to return the same instance")
}
