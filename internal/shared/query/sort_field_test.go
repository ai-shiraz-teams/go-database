package query

import (
	"encoding/json"
	"testing"
)

// TestSortField_Creation validates SortField struct creation
func TestSortField_Creation(t *testing.T) {
	// Arrange
	expectedField := "name"
	expectedOrder := SortOrderAsc

	// Act
	sortField := SortField{
		Field: expectedField,
		Order: expectedOrder,
	}

	// Assert
	if sortField.Field != expectedField {
		t.Errorf("Expected Field %q, got %q", expectedField, sortField.Field)
	}

	if sortField.Order != expectedOrder {
		t.Errorf("Expected Order %q, got %q", expectedOrder, sortField.Order)
	}
}

// TestSortField_ZeroValues validates SortField zero values
func TestSortField_ZeroValues(t *testing.T) {
	// Arrange & Act
	var sortField SortField

	// Assert
	if sortField.Field != "" {
		t.Errorf("Expected empty Field, got %q", sortField.Field)
	}

	if sortField.Order != "" {
		t.Errorf("Expected empty Order, got %q", sortField.Order)
	}
}

// TestSortField_JSONSerialization validates JSON marshaling and unmarshaling
func TestSortField_JSONSerialization(t *testing.T) {
	tests := []struct {
		name         string
		sortField    SortField
		expectedJSON string
	}{
		{
			name: "Sort field with ascending order",
			sortField: SortField{
				Field: "created_at",
				Order: SortOrderAsc,
			},
			expectedJSON: `{"field":"created_at","order":"asc"}`,
		},
		{
			name: "Sort field with descending order",
			sortField: SortField{
				Field: "updated_at",
				Order: SortOrderDesc,
			},
			expectedJSON: `{"field":"updated_at","order":"desc"}`,
		},
		{
			name: "Sort field with empty values",
			sortField: SortField{
				Field: "",
				Order: "",
			},
			expectedJSON: `{"field":"","order":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			t.Run("Marshal", func(t *testing.T) {
				// Arrange & Act
				jsonData, err := json.Marshal(tt.sortField)

				// Assert
				if err != nil {
					t.Fatalf("Failed to marshal SortField: %v", err)
				}

				if string(jsonData) != tt.expectedJSON {
					t.Errorf("Expected JSON %q, got %q", tt.expectedJSON, string(jsonData))
				}
			})

			// Test unmarshaling
			t.Run("Unmarshal", func(t *testing.T) {
				// Arrange
				var sortField SortField

				// Act
				err := json.Unmarshal([]byte(tt.expectedJSON), &sortField)

				// Assert
				if err != nil {
					t.Fatalf("Failed to unmarshal SortField: %v", err)
				}

				if sortField.Field != tt.sortField.Field {
					t.Errorf("Expected Field %q, got %q", tt.sortField.Field, sortField.Field)
				}

				if sortField.Order != tt.sortField.Order {
					t.Errorf("Expected Order %q, got %q", tt.sortField.Order, sortField.Order)
				}
			})
		})
	}
}

// TestSortField_FieldValidation validates different field names
func TestSortField_FieldValidation(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		order     SortOrder
	}{
		{
			name:      "Simple field name",
			fieldName: "name",
			order:     SortOrderAsc,
		},
		{
			name:      "Snake case field name",
			fieldName: "created_at",
			order:     SortOrderDesc,
		},
		{
			name:      "Dot notation field name",
			fieldName: "user.email",
			order:     SortOrderAsc,
		},
		{
			name:      "Complex field name",
			fieldName: "metadata.settings.enabled",
			order:     SortOrderDesc,
		},
		{
			name:      "Numeric field name",
			fieldName: "id",
			order:     SortOrderAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			sortField := SortField{
				Field: tt.fieldName,
				Order: tt.order,
			}

			// Assert
			if sortField.Field != tt.fieldName {
				t.Errorf("Expected Field %q, got %q", tt.fieldName, sortField.Field)
			}

			if sortField.Order != tt.order {
				t.Errorf("Expected Order %q, got %q", tt.order, sortField.Order)
			}
		})
	}
}

// TestSortField_Comparison validates SortField equality
func TestSortField_Comparison(t *testing.T) {
	// Arrange
	sortField1 := SortField{Field: "name", Order: SortOrderAsc}
	sortField2 := SortField{Field: "name", Order: SortOrderAsc}
	sortField3 := SortField{Field: "email", Order: SortOrderAsc}
	sortField4 := SortField{Field: "name", Order: SortOrderDesc}

	// Act & Assert
	if sortField1 != sortField2 {
		t.Error("Expected identical SortFields to be equal")
	}

	if sortField1 == sortField3 {
		t.Error("Expected SortFields with different fields to be unequal")
	}

	if sortField1 == sortField4 {
		t.Error("Expected SortFields with different orders to be unequal")
	}
}
