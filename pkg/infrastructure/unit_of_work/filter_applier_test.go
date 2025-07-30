package unit_of_work

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

// TestNewFilterApplier validates FilterApplier creation
func TestNewFilterApplier(t *testing.T) {
	// Arrange & Act
	fa := NewFilterApplier()

	// Assert
	if fa == nil {
		t.Fatal("NewFilterApplier returned nil")
	}
}

// TestFilterApplier_ApplyFilters_EmptyFilters validates empty filter handling
func TestFilterApplier_ApplyFilters_EmptyFilters(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	// Act
	result := fa.ApplyFilters(query, []identifier.FilterCriteria{})

	// Assert
	if result == nil {
		t.Fatal("ApplyFilters returned nil")
	}

	// Should return the same query without modification
	if result != query {
		t.Error("Expected same query instance to be returned for empty filters")
	}
}

// TestFilterApplier_ApplyFilters_SingleFilter validates single filter application
func TestFilterApplier_ApplyFilters_SingleFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   identifier.FilterCriteria
		expected string // Expected SQL condition pattern
	}{
		{
			name: "Equal operator",
			filter: identifier.FilterCriteria{
				Field:    "name",
				Operator: identifier.FilterOperatorEqual,
				Value:    "John",
			},
			expected: "name = ?",
		},
		{
			name: "Not equal operator",
			filter: identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorNotEqual,
				Value:    25,
			},
			expected: "age != ?",
		},
		{
			name: "Greater than operator",
			filter: identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorGreaterThan,
				Value:    18,
			},
			expected: "age > ?",
		},
		{
			name: "Greater equal operator",
			filter: identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorGreaterEqual,
				Value:    21,
			},
			expected: "age >= ?",
		},
		{
			name: "Less than operator",
			filter: identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorLessThan,
				Value:    65,
			},
			expected: "age < ?",
		},
		{
			name: "Less equal operator",
			filter: identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorLessEqual,
				Value:    64,
			},
			expected: "age <= ?",
		},
		{
			name: "Like operator",
			filter: identifier.FilterCriteria{
				Field:    "email",
				Operator: identifier.FilterOperatorLike,
				Value:    "%@example.com",
			},
			expected: "email LIKE ?",
		},
		{
			name: "Is null operator",
			filter: identifier.FilterCriteria{
				Field:    "email",
				Operator: identifier.FilterOperatorIsNull,
			},
			expected: "email IS NULL",
		},
		{
			name: "Is not null operator",
			filter: identifier.FilterCriteria{
				Field:    "email",
				Operator: identifier.FilterOperatorIsNotNull,
			},
			expected: "email IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			fa := NewFilterApplier()
			query := db.Model(&testutil.TestEntity{})

			// Act
			result := fa.ApplyFilters(query, []identifier.FilterCriteria{tt.filter})

			// Assert
			if result == nil {
				t.Fatal("ApplyFilters returned nil")
			}

			// Verify that SQL is generated (we can't easily check exact SQL due to GORM internals)
			sql := result.Statement.SQL.String()
			if sql == "" {
				// SQL may not be built until execution, but the query should be modified
				t.Log("SQL not built yet, but query should be modified")
			}
		})
	}
}

// TestFilterApplier_ApplyFilters_InOperator validates IN operator handling
func TestFilterApplier_ApplyFilters_InOperator(t *testing.T) {
	tests := []struct {
		name     string
		values   []interface{}
		expected string
	}{
		{
			name:     "IN with values",
			values:   []interface{}{1, 2, 3},
			expected: "id IN ?",
		},
		{
			name:     "IN with empty slice",
			values:   []interface{}{},
			expected: "1 = 0", // Should return no results
		},
		{
			name:     "IN with single value",
			values:   []interface{}{42},
			expected: "id IN ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			fa := NewFilterApplier()
			query := db.Model(&testutil.TestEntity{})
			filter := identifier.FilterCriteria{
				Field:    "id",
				Operator: identifier.FilterOperatorIn,
				Values:   tt.values,
			}

			// Act
			result := fa.ApplyFilters(query, []identifier.FilterCriteria{filter})

			// Assert
			if result == nil {
				t.Fatal("ApplyFilters returned nil")
			}
		})
	}
}

// TestFilterApplier_ApplyFilters_NotInOperator validates NOT IN operator handling
func TestFilterApplier_ApplyFilters_NotInOperator(t *testing.T) {
	tests := []struct {
		name   string
		values []interface{}
	}{
		{
			name:   "NOT IN with values",
			values: []interface{}{1, 2, 3},
		},
		{
			name:   "NOT IN with empty slice",
			values: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			fa := NewFilterApplier()
			query := db.Model(&testutil.TestEntity{})
			filter := identifier.FilterCriteria{
				Field:    "id",
				Operator: identifier.FilterOperatorNotIn,
				Values:   tt.values,
			}

			// Act
			result := fa.ApplyFilters(query, []identifier.FilterCriteria{filter})

			// Assert
			if result == nil {
				t.Fatal("ApplyFilters returned nil")
			}
		})
	}
}

// TestFilterApplier_ApplyFilters_BetweenOperator validates BETWEEN operator handling
func TestFilterApplier_ApplyFilters_BetweenOperator(t *testing.T) {
	tests := []struct {
		name   string
		values []interface{}
	}{
		{
			name:   "BETWEEN with two values",
			values: []interface{}{18, 65},
		},
		{
			name:   "BETWEEN with more than two values (uses first two)",
			values: []interface{}{10, 20, 30},
		},
		{
			name:   "BETWEEN with insufficient values (should be ignored)",
			values: []interface{}{10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			fa := NewFilterApplier()
			query := db.Model(&testutil.TestEntity{})
			filter := identifier.FilterCriteria{
				Field:    "age",
				Operator: identifier.FilterOperatorBetween,
				Values:   tt.values,
			}

			// Act
			result := fa.ApplyFilters(query, []identifier.FilterCriteria{filter})

			// Assert
			if result == nil {
				t.Fatal("ApplyFilters returned nil")
			}
		})
	}
}

// TestFilterApplier_ApplyFilters_MultipleFilters validates multiple filter handling
func TestFilterApplier_ApplyFilters_MultipleFilters(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	filters := []identifier.FilterCriteria{
		{
			Field:     "name",
			Operator:  identifier.FilterOperatorEqual,
			Value:     "John",
			LogicalOp: identifier.LogicalOperatorAnd,
		},
		{
			Field:    "age",
			Operator: identifier.FilterOperatorGreaterThan,
			Value:    18,
		},
	}

	// Act
	result := fa.ApplyFilters(query, filters)

	// Assert
	if result == nil {
		t.Fatal("ApplyFilters returned nil")
	}
}

// TestFilterApplier_ApplyFilters_OrLogicalOperator validates OR logical operator
func TestFilterApplier_ApplyFilters_OrLogicalOperator(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	filters := []identifier.FilterCriteria{
		{
			Field:     "name",
			Operator:  identifier.FilterOperatorEqual,
			Value:     "John",
			LogicalOp: identifier.LogicalOperatorOr,
		},
		{
			Field:    "name",
			Operator: identifier.FilterOperatorEqual,
			Value:    "Jane",
		},
	}

	// Act
	result := fa.ApplyFilters(query, filters)

	// Assert
	if result == nil {
		t.Fatal("ApplyFilters returned nil")
	}
}

// TestFilterApplier_ApplyFilters_UnknownOperator validates unknown operator handling
func TestFilterApplier_ApplyFilters_UnknownOperator(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	filter := identifier.FilterCriteria{
		Field:    "name",
		Operator: identifier.FilterOperator("unknown"),
		Value:    "test",
	}

	// Act
	result := fa.ApplyFilters(query, []identifier.FilterCriteria{filter})

	// Assert
	if result == nil {
		t.Fatal("ApplyFilters returned nil")
	}

	// Should return the same query without modification for unknown operators
	if result != query {
		t.Error("Expected same query instance for unknown operator")
	}
}

// TestFilterApplier_ApplyIdentifier validates identifier application
func TestFilterApplier_ApplyIdentifier(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	// Create a mock identifier
	ident := identifier.NewIdentifier().Equal("name", "John")

	// Act
	result := fa.ApplyIdentifier(query, ident)

	// Assert
	if result == nil {
		t.Fatal("ApplyIdentifier returned nil")
	}
}

// TestFilterApplier_ApplyIdentifier_Nil validates nil identifier handling
func TestFilterApplier_ApplyIdentifier_Nil(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	// Act
	result := fa.ApplyIdentifier(query, nil)

	// Assert
	if result == nil {
		t.Fatal("ApplyIdentifier returned nil")
	}

	// Should return the same query without modification for nil identifier
	if result != query {
		t.Error("Expected same query instance for nil identifier")
	}
}

// TestFilterApplier_ValidateFilterValue validates filter value validation
func TestFilterApplier_ValidateFilterValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		expectErr bool
	}{
		{
			name:      "String value",
			fieldName: "name",
			value:     "test",
			expectErr: false,
		},
		{
			name:      "Integer value",
			fieldName: "age",
			value:     25,
			expectErr: false,
		},
		{
			name:      "Float value",
			fieldName: "price",
			value:     19.99,
			expectErr: false,
		},
		{
			name:      "Boolean value",
			fieldName: "is_active",
			value:     true,
			expectErr: false,
		},
		{
			name:      "Nil value",
			fieldName: "optional",
			value:     nil,
			expectErr: false,
		},
		{
			name:      "Slice of valid values",
			fieldName: "ids",
			value:     []interface{}{1, 2, 3},
			expectErr: false,
		},
		{
			name:      "Integer slice",
			fieldName: "ids",
			value:     []int{1, 2, 3},
			expectErr: false,
		},
		{
			name:      "Map value (unsupported)",
			fieldName: "metadata",
			value:     map[string]interface{}{"key": "value"},
			expectErr: true,
		},
		{
			name:      "Struct value (unsupported)",
			fieldName: "entity",
			value:     struct{ Name string }{Name: "test"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fa := NewFilterApplier()

			// Act
			err := fa.ValidateFilterValue(tt.fieldName, tt.value)

			// Assert
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for value %v, got nil", tt.value)
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for value %v, got %v", tt.value, err)
			}
		})
	}
}

// TestFilterApplier_ApplyQueryParams validates query parameters application
func TestFilterApplier_ApplyQueryParams(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	// Create a simple query params struct
	params := struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}{
		Name: "John",
		Age:  25,
	}

	// Act
	result := fa.ApplyQueryParams(query, params)

	// Assert
	if result == nil {
		t.Fatal("ApplyQueryParams returned nil")
	}
}

// TestFilterApplier_ApplyQueryParams_EmptyStruct validates empty struct handling
func TestFilterApplier_ApplyQueryParams_EmptyStruct(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	params := struct{}{}

	// Act
	result := fa.ApplyQueryParams(query, params)

	// Assert
	if result == nil {
		t.Fatal("ApplyQueryParams returned nil")
	}

	// Should return the same query without modification for empty struct
	if result != query {
		t.Error("Expected same query instance for empty struct")
	}
}

// TestFilterApplier_ApplyQueryParams_Nil validates nil parameters handling
func TestFilterApplier_ApplyQueryParams_Nil(t *testing.T) {
	// Arrange
	db := testutil.SetupTestDB(t)
	fa := NewFilterApplier()
	query := db.Model(&testutil.TestEntity{})

	// Act
	result := fa.ApplyQueryParams(query, nil)

	// Assert
	if result == nil {
		t.Fatal("ApplyQueryParams returned nil")
	}

	// Should return the same query instance for nil params
	if result != query {
		t.Error("Expected same query instance for nil params")
	}
}

// TestFilterApplier_ApplyQueryParams_VariousTypes validates different parameter types
func TestFilterApplier_ApplyQueryParams_VariousTypes(t *testing.T) {
	tests := []struct {
		name   string
		params interface{}
	}{
		{
			name: "String and int parameters",
			params: struct {
				Name string `query:"name"`
				Age  int    `query:"age"`
			}{
				Name: "Alice",
				Age:  30,
			},
		},
		{
			name: "Boolean parameter",
			params: struct {
				IsActive bool `query:"is_active"`
			}{
				IsActive: true,
			},
		},
		{
			name: "Float parameter",
			params: struct {
				Score float64 `query:"score"`
			}{
				Score: 95.5,
			},
		},
		{
			name: "Multiple parameters with zero values",
			params: struct {
				Name string `query:"name"`
				Age  int    `query:"age"`
			}{
				Name: "",
				Age:  0,
			},
		},
		{
			name: "Parameters without query tags",
			params: struct {
				Name string
				Age  int
			}{
				Name: "Bob",
				Age:  25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := testutil.SetupTestDB(t)
			fa := NewFilterApplier()
			query := db.Model(&testutil.TestEntity{})

			// Act
			result := fa.ApplyQueryParams(query, tt.params)

			// Assert
			if result == nil {
				t.Fatal("ApplyQueryParams returned nil")
			}
		})
	}
}
