package app

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestGetValidator_ReturnsInstance(t *testing.T) {
	instance := GetValidator()

	if instance == nil {
		t.Fatal("GetValidator() returned nil, expected a *validator.Validate instance")
	}

	type testStruct struct {
		Value int `validate:"required"`
	}

	err := instance.Struct(testStruct{})
	if err == nil {
		t.Error("Expected validation error for missing required field, but got nil")
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Errorf("Expected error to be validator.ValidationErrors, got %T", err)
	}
	if len(validationErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(validationErrs))
	}
}

func TestGetValidator_IsSingleton(t *testing.T) {
	instance1 := GetValidator()
	instance2 := GetValidator()

	if instance1 == nil || instance2 == nil {
		t.Fatal("GetValidator() returned nil unexpectedly")
	}

	// Checks if the memory addresses are the same.
	if instance1 != instance2 {
		t.Errorf("Expected GetValidator() to return the same instance, but got different instances: %p vs %p", instance1, instance2)
	}
}
