package errors

import (
	"testing"
)

func TestEntityNotFoundError_Error(t *testing.T) {
	// Arrange
	err := NewEntityNotFoundError("User", 123)

	// Act
	message := err.Error()

	// Assert
	expected := "User with ID 123 not found"
	if message != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, message)
	}
}

func TestEntityNotFoundError_Fields(t *testing.T) {
	// Arrange
	entityType := "Product"
	id := "abc-123"

	// Act
	err := NewEntityNotFoundError(entityType, id)

	// Assert
	if err.EntityType != entityType {
		t.Errorf("Expected EntityType '%s', got '%s'", entityType, err.EntityType)
	}
	if err.ID != id {
		t.Errorf("Expected ID '%v', got '%v'", id, err.ID)
	}
}

func TestValidationError_Error(t *testing.T) {
	// Arrange
	err := NewValidationError("email", "must be a valid email address")

	// Act
	message := err.Error()

	// Assert
	expected := "validation error on field 'email': must be a valid email address"
	if message != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, message)
	}
}

func TestValidationError_Fields(t *testing.T) {
	// Arrange
	field := "password"
	message := "must be at least 8 characters"

	// Act
	err := NewValidationError(field, message)

	// Assert
	if err.Field != field {
		t.Errorf("Expected Field '%s', got '%s'", field, err.Field)
	}
	if err.Message != message {
		t.Errorf("Expected Message '%s', got '%s'", message, err.Message)
	}
}

func TestDuplicateEntityError_Error(t *testing.T) {
	// Arrange
	err := NewDuplicateEntityError("User", "email", "john@example.com")

	// Act
	message := err.Error()

	// Assert
	expected := "User with email 'john@example.com' already exists"
	if message != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, message)
	}
}

func TestDuplicateEntityError_Fields(t *testing.T) {
	// Arrange
	entityType := "Product"
	field := "sku"
	value := "PROD-123"

	// Act
	err := NewDuplicateEntityError(entityType, field, value)

	// Assert
	if err.EntityType != entityType {
		t.Errorf("Expected EntityType '%s', got '%s'", entityType, err.EntityType)
	}
	if err.Field != field {
		t.Errorf("Expected Field '%s', got '%s'", field, err.Field)
	}
	if err.Value != value {
		t.Errorf("Expected Value '%v', got '%v'", value, err.Value)
	}
}

func TestConcurrencyError_Error(t *testing.T) {
	// Arrange
	err := NewConcurrencyError("Order", 456)

	// Act
	message := err.Error()

	// Assert
	expected := "concurrent modification detected for Order with ID 456"
	if message != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, message)
	}
}

func TestConcurrencyError_Fields(t *testing.T) {
	// Arrange
	entityType := "Invoice"
	id := int64(789)

	// Act
	err := NewConcurrencyError(entityType, id)

	// Assert
	if err.EntityType != entityType {
		t.Errorf("Expected EntityType '%s', got '%s'", entityType, err.EntityType)
	}
	if err.ID != id {
		t.Errorf("Expected ID '%v', got '%v'", id, err.ID)
	}
}

func TestErrorTypes_ImplementError(t *testing.T) {
	// Test that all custom error types implement the error interface
	tests := []struct {
		name string
		err  error
	}{
		{"EntityNotFoundError", NewEntityNotFoundError("User", 1)},
		{"ValidationError", NewValidationError("field", "message")},
		{"DuplicateEntityError", NewDuplicateEntityError("Type", "field", "value")},
		{"ConcurrencyError", NewConcurrencyError("Type", 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() == "" {
				t.Errorf("%s should return non-empty error message", tt.name)
			}
		})
	}
}

func TestErrorMessages_NonEmptyValues(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"EntityNotFoundError", NewEntityNotFoundError("User", 1)},
		{"ValidationError", NewValidationError("email", "invalid")},
		{"DuplicateEntityError", NewDuplicateEntityError("User", "email", "test@example.com")},
		{"ConcurrencyError", NewConcurrencyError("User", 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := tt.error.Error()
			if message == "" {
				t.Errorf("%s should return non-empty error message", tt.name)
			}
		})
	}
}

func TestErrorMessages_WithDifferentTypes(t *testing.T) {
	// Test with different ID types
	tests := []struct {
		name string
		id   interface{}
	}{
		{"int", 123},
		{"string", "abc-123"},
		{"int64", int64(456)},
		{"uint", uint(789)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewEntityNotFoundError("Entity", tt.id)
			message := err.Error()
			if message == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestErrorsWithSpecialCharacters(t *testing.T) {
	// Test that error messages handle special characters properly
	tests := []struct {
		name   string
		create func() error
	}{
		{
			"EntityNotFound with quotes",
			func() error { return NewEntityNotFoundError("User's Profile", "id-with-'quotes'") },
		},
		{
			"Validation with newlines",
			func() error { return NewValidationError("description", "must not contain\nnewlines") },
		},
		{
			"Duplicate with unicode",
			func() error { return NewDuplicateEntityError("Café", "name", "Café Münü") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create()
			message := err.Error()
			if message == "" {
				t.Error("Error message should not be empty even with special characters")
			}
		})
	}
}
