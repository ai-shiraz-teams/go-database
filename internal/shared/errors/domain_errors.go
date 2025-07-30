package errors

import "fmt"

// EntityNotFoundError represents an error when an entity is not found
type EntityNotFoundError struct {
	EntityType string
	ID         interface{}
}

func (e *EntityNotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %v not found", e.EntityType, e.ID)
}

// NewEntityNotFoundError creates a new EntityNotFoundError
func NewEntityNotFoundError(entityType string, id interface{}) *EntityNotFoundError {
	return &EntityNotFoundError{
		EntityType: entityType,
		ID:         id,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// DuplicateEntityError represents an error when trying to create a duplicate entity
type DuplicateEntityError struct {
	EntityType string
	Field      string
	Value      interface{}
}

func (e *DuplicateEntityError) Error() string {
	return fmt.Sprintf("%s with %s '%v' already exists", e.EntityType, e.Field, e.Value)
}

// NewDuplicateEntityError creates a new DuplicateEntityError
func NewDuplicateEntityError(entityType, field string, value interface{}) *DuplicateEntityError {
	return &DuplicateEntityError{
		EntityType: entityType,
		Field:      field,
		Value:      value,
	}
}

// ConcurrencyError represents an optimistic locking error
type ConcurrencyError struct {
	EntityType string
	ID         interface{}
}

func (e *ConcurrencyError) Error() string {
	return fmt.Sprintf("concurrent modification detected for %s with ID %v", e.EntityType, e.ID)
}

// NewConcurrencyError creates a new ConcurrencyError
func NewConcurrencyError(entityType string, id interface{}) *ConcurrencyError {
	return &ConcurrencyError{
		EntityType: entityType,
		ID:         id,
	}
}
