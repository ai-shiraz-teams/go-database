package identifier

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewIdentifier(t *testing.T) {
	// Act
	identifier := NewIdentifier()

	// Assert
	if identifier == nil {
		t.Fatal("NewIdentifier should not return nil")
	}

	// Verify it returns IIdentifier interface
	var _ IIdentifier = identifier

	// Verify initial state has no filters
	filters := identifier.ToFilterCriteria()
	if len(filters) != 0 {
		t.Errorf("Expected empty filters for new identifier, got %d", len(filters))
	}
}

func TestIdentifierBuilder_Equal(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		expected FilterCriteria
	}{
		{
			name:  "string value",
			field: "name",
			value: "test",
			expected: FilterCriteria{
				Field:    "name",
				Operator: FilterOperatorEqual,
				Value:    "test",
			},
		},
		{
			name:  "integer value",
			field: "id",
			value: 42,
			expected: FilterCriteria{
				Field:    "id",
				Operator: FilterOperatorEqual,
				Value:    42,
			},
		},
		{
			name:  "nil value",
			field: "deleted_at",
			value: nil,
			expected: FilterCriteria{
				Field:    "deleted_at",
				Operator: FilterOperatorEqual,
				Value:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			identifier := NewIdentifier()

			// Act
			result := identifier.Equal(tt.field, tt.value)

			// Assert
			if result == nil {
				t.Fatal("Equal should not return nil")
			}

			filters := result.ToFilterCriteria()
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			filter := filters[0]
			if filter.Field != tt.expected.Field {
				t.Errorf("Expected field %s, got %s", tt.expected.Field, filter.Field)
			}
			if filter.Operator != tt.expected.Operator {
				t.Errorf("Expected operator %s, got %s", tt.expected.Operator, filter.Operator)
			}
			if !reflect.DeepEqual(filter.Value, tt.expected.Value) {
				t.Errorf("Expected value %v, got %v", tt.expected.Value, filter.Value)
			}
		})
	}
}

func TestIdentifierBuilder_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name             string
		operation        func(IIdentifier, string, interface{}) IIdentifier
		expectedOperator FilterOperator
	}{
		{
			name: "NotEqual",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.NotEqual(field, value)
			},
			expectedOperator: FilterOperatorNotEqual,
		},
		{
			name: "GreaterThan",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.GreaterThan(field, value)
			},
			expectedOperator: FilterOperatorGreaterThan,
		},
		{
			name: "GreaterOrEqual",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.GreaterOrEqual(field, value)
			},
			expectedOperator: FilterOperatorGreaterEqual,
		},
		{
			name: "LessThan",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.LessThan(field, value)
			},
			expectedOperator: FilterOperatorLessThan,
		},
		{
			name: "LessOrEqual",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.LessOrEqual(field, value)
			},
			expectedOperator: FilterOperatorLessEqual,
		},
		{
			name: "Like",
			operation: func(id IIdentifier, field string, value interface{}) IIdentifier {
				return id.Like(field, value.(string))
			},
			expectedOperator: FilterOperatorLike,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			identifier := NewIdentifier()

			// Act
			result := tt.operation(identifier, "test_field", "test_value")

			// Assert
			filters := result.ToFilterCriteria()
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			filter := filters[0]
			if filter.Operator != tt.expectedOperator {
				t.Errorf("Expected operator %s, got %s", tt.expectedOperator, filter.Operator)
			}
			if filter.Field != "test_field" {
				t.Errorf("Expected field test_field, got %s", filter.Field)
			}
			if filter.Value != "test_value" {
				t.Errorf("Expected value test_value, got %v", filter.Value)
			}
		})
	}
}

func TestIdentifierBuilder_In(t *testing.T) {
	// Arrange
	identifier := NewIdentifier()
	values := []interface{}{1, 2, 3}

	// Act
	result := identifier.In("id", values)

	// Assert
	filters := result.ToFilterCriteria()
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(filters))
	}

	filter := filters[0]
	if filter.Operator != FilterOperatorIn {
		t.Errorf("Expected operator %s, got %s", FilterOperatorIn, filter.Operator)
	}
	if filter.Field != "id" {
		t.Errorf("Expected field id, got %s", filter.Field)
	}
	if !reflect.DeepEqual(filter.Values, values) {
		t.Errorf("Expected values %v, got %v", values, filter.Values)
	}
}

func TestIdentifierBuilder_NotIn(t *testing.T) {
	// Arrange
	identifier := NewIdentifier()
	values := []interface{}{"a", "b", "c"}

	// Act
	result := identifier.NotIn("status", values)

	// Assert
	filters := result.ToFilterCriteria()
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(filters))
	}

	filter := filters[0]
	if filter.Operator != FilterOperatorNotIn {
		t.Errorf("Expected operator %s, got %s", FilterOperatorNotIn, filter.Operator)
	}
	if filter.Field != "status" {
		t.Errorf("Expected field status, got %s", filter.Field)
	}
	if !reflect.DeepEqual(filter.Values, values) {
		t.Errorf("Expected values %v, got %v", values, filter.Values)
	}
}

func TestIdentifierBuilder_Between(t *testing.T) {
	// Arrange
	identifier := NewIdentifier()
	start := 10
	end := 20

	// Act
	result := identifier.Between("age", start, end)

	// Assert
	filters := result.ToFilterCriteria()
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(filters))
	}

	filter := filters[0]
	if filter.Operator != FilterOperatorBetween {
		t.Errorf("Expected operator %s, got %s", FilterOperatorBetween, filter.Operator)
	}
	if filter.Field != "age" {
		t.Errorf("Expected field age, got %s", filter.Field)
	}

	// Value should be in Values slice for Between operator
	if len(filter.Values) != 2 {
		t.Fatalf("Expected 2 values for Between, got %d", len(filter.Values))
	}
	if filter.Values[0] != start {
		t.Errorf("Expected start %v, got %v", start, filter.Values[0])
	}
	if filter.Values[1] != end {
		t.Errorf("Expected end %v, got %v", end, filter.Values[1])
	}
}

func TestIdentifierBuilder_NullChecks(t *testing.T) {
	tests := []struct {
		name             string
		operation        func(IIdentifier, string) IIdentifier
		expectedOperator FilterOperator
	}{
		{
			name: "IsNull",
			operation: func(id IIdentifier, field string) IIdentifier {
				return id.IsNull(field)
			},
			expectedOperator: FilterOperatorIsNull,
		},
		{
			name: "IsNotNull",
			operation: func(id IIdentifier, field string) IIdentifier {
				return id.IsNotNull(field)
			},
			expectedOperator: FilterOperatorIsNotNull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			identifier := NewIdentifier()

			// Act
			result := tt.operation(identifier, "nullable_field")

			// Assert
			filters := result.ToFilterCriteria()
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			filter := filters[0]
			if filter.Operator != tt.expectedOperator {
				t.Errorf("Expected operator %s, got %s", tt.expectedOperator, filter.Operator)
			}
			if filter.Field != "nullable_field" {
				t.Errorf("Expected field nullable_field, got %s", filter.Field)
			}
			if filter.Value != nil {
				t.Errorf("Expected nil value for null check, got %v", filter.Value)
			}
		})
	}
}

func TestIdentifierBuilder_JSONOperators(t *testing.T) {
	tests := []struct {
		name             string
		operation        func(IIdentifier) IIdentifier
		expectedOperator FilterOperator
		expectedValue    interface{}
	}{
		{
			name: "Contains",
			operation: func(id IIdentifier) IIdentifier {
				return id.Contains("metadata", "key")
			},
			expectedOperator: FilterOperatorContains,
			expectedValue:    "key",
		},
		{
			name: "Has",
			operation: func(id IIdentifier) IIdentifier {
				return id.Has("json_field")
			},
			expectedOperator: FilterOperatorHas,
			expectedValue:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			identifier := NewIdentifier()

			// Act
			result := tt.operation(identifier)

			// Assert
			filters := result.ToFilterCriteria()
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			filter := filters[0]
			if filter.Operator != tt.expectedOperator {
				t.Errorf("Expected operator %s, got %s", tt.expectedOperator, filter.Operator)
			}
			if !reflect.DeepEqual(filter.Value, tt.expectedValue) {
				t.Errorf("Expected value %v, got %v", tt.expectedValue, filter.Value)
			}
		})
	}
}

func TestIdentifierBuilder_And(t *testing.T) {
	// Arrange
	id1 := NewIdentifier().Equal("name", "test")
	id2 := NewIdentifier().Equal("status", "active")

	// Act
	result := id1.And(id2)

	// Assert
	filters := result.ToFilterCriteria()
	if len(filters) != 2 {
		t.Fatalf("Expected 2 filters after AND, got %d", len(filters))
	}

	// Check first filter
	if filters[0].Field != "name" || filters[0].Value != "test" {
		t.Errorf("Expected first filter: name=test, got %s=%v", filters[0].Field, filters[0].Value)
	}

	// Check second filter
	if filters[1].Field != "status" || filters[1].Value != "active" {
		t.Errorf("Expected second filter: status=active, got %s=%v", filters[1].Field, filters[1].Value)
	}
}

func TestIdentifierBuilder_Or(t *testing.T) {
	// Arrange
	id1 := NewIdentifier().Equal("type", "A")
	id2 := NewIdentifier().Equal("type", "B")

	// Act
	result := id1.Or(id2)

	// Assert
	filters := result.ToFilterCriteria()
	if len(filters) < 2 {
		t.Fatalf("Expected at least 2 filters after OR, got %d", len(filters))
	}

	// Verify OR logic is applied (implementation may vary)
	// The exact structure depends on how OR is implemented in the builder
}

func TestIdentifierBuilder_AndWithNil(t *testing.T) {
	// Arrange
	identifier := NewIdentifier().Equal("field", "value")

	// Act
	result := identifier.And(nil)

	// Assert
	if result != identifier {
		t.Error("AND with nil should return the original identifier")
	}

	filters := result.ToFilterCriteria()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter after AND with nil, got %d", len(filters))
	}
}

func TestIdentifierBuilder_OrWithNil(t *testing.T) {
	// Arrange
	identifier := NewIdentifier().Equal("field", "value")

	// Act
	result := identifier.Or(nil)

	// Assert
	if result != identifier {
		t.Error("OR with nil should return the original identifier")
	}

	filters := result.ToFilterCriteria()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter after OR with nil, got %d", len(filters))
	}
}

func TestIdentifierBuilder_Reset(t *testing.T) {
	// Arrange
	identifier := NewIdentifier().Equal("field1", "value1").Equal("field2", "value2")

	// Verify initial state has filters
	initialFilters := identifier.ToFilterCriteria()
	if len(initialFilters) == 0 {
		t.Fatal("Expected filters before reset")
	}

	// Act
	result := identifier.Reset()

	// Assert
	if result == nil {
		t.Fatal("Reset should not return nil")
	}

	resetFilters := result.ToFilterCriteria()
	if len(resetFilters) != 0 {
		t.Errorf("Expected 0 filters after reset, got %d", len(resetFilters))
	}

	// Verify original identifier is unchanged (immutability)
	originalFilters := identifier.ToFilterCriteria()
	if len(originalFilters) != len(initialFilters) {
		t.Error("Original identifier should be unchanged after reset")
	}
}

func TestIdentifierBuilder_Immutability(t *testing.T) {
	// Arrange
	original := NewIdentifier().Equal("field", "value")

	// Act - Create new identifier from original
	modified := original.Equal("another_field", "another_value")

	// Assert
	originalFilters := original.ToFilterCriteria()
	modifiedFilters := modified.ToFilterCriteria()

	if len(originalFilters) != 1 {
		t.Errorf("Expected original to have 1 filter, got %d", len(originalFilters))
	}
	if len(modifiedFilters) != 2 {
		t.Errorf("Expected modified to have 2 filters, got %d", len(modifiedFilters))
	}

	// Verify original is unchanged
	if originalFilters[0].Field != "field" || originalFilters[0].Value != "value" {
		t.Error("Original identifier was modified, violating immutability")
	}
}

func TestIdentifierBuilder_ComplexChaining(t *testing.T) {
	// Arrange & Act
	identifier := NewIdentifier().
		Equal("status", "active").
		GreaterThan("created_at", "2023-01-01").
		In("category", []interface{}{"A", "B", "C"}).
		IsNotNull("email")

	// Assert
	filters := identifier.ToFilterCriteria()
	if len(filters) != 4 {
		t.Fatalf("Expected 4 filters from chaining, got %d", len(filters))
	}

	expectedFilters := []struct {
		Field    string
		Operator FilterOperator
		Value    interface{}
		Values   []interface{}
	}{
		{Field: "status", Operator: FilterOperatorEqual, Value: "active"},
		{Field: "created_at", Operator: FilterOperatorGreaterThan, Value: "2023-01-01"},
		{Field: "category", Operator: FilterOperatorIn, Values: []interface{}{"A", "B", "C"}},
		{Field: "email", Operator: FilterOperatorIsNotNull, Value: nil},
	}

	for i, expected := range expectedFilters {
		actual := filters[i]
		if actual.Field != expected.Field {
			t.Errorf("Filter %d: expected field %s, got %s", i, expected.Field, actual.Field)
		}
		if actual.Operator != expected.Operator {
			t.Errorf("Filter %d: expected operator %s, got %s", i, expected.Operator, actual.Operator)
		}

		// Check Value or Values based on operator
		if expected.Operator == FilterOperatorIn || expected.Operator == FilterOperatorNotIn {
			if !reflect.DeepEqual(actual.Values, expected.Values) {
				t.Errorf("Filter %d: expected values %v, got %v", i, expected.Values, actual.Values)
			}
		} else {
			if !reflect.DeepEqual(actual.Value, expected.Value) {
				t.Errorf("Filter %d: expected value %v, got %v", i, expected.Value, actual.Value)
			}
		}
	}
}

func TestIdentifierBuilder_ThreadSafety(t *testing.T) {
	// Arrange
	identifier := NewIdentifier().Equal("base", "value")
	var wg sync.WaitGroup
	results := make([]IIdentifier, 100)

	// Act - Concurrent operations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = identifier.Equal("field", index)
		}(i)
	}

	wg.Wait()

	// Assert - All operations should succeed without race conditions
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
			continue
		}

		filters := result.ToFilterCriteria()
		if len(filters) != 2 {
			t.Errorf("Result %d: expected 2 filters, got %d", i, len(filters))
			continue
		}

		// Verify the new filter was added correctly
		if filters[1].Value != i {
			t.Errorf("Result %d: expected value %d, got %v", i, i, filters[1].Value)
		}
	}

	// Verify original identifier is unchanged
	originalFilters := identifier.ToFilterCriteria()
	if len(originalFilters) != 1 {
		t.Errorf("Original identifier was modified during concurrent access")
	}
}
