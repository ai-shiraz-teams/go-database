package query

import "testing"

// TestSortOrder_Constants validates SortOrder constant values
func TestSortOrder_Constants(t *testing.T) {
	// Arrange & Act & Assert
	tests := []struct {
		name     string
		actual   SortOrder
		expected string
	}{
		{
			name:     "SortOrderAsc should be 'asc'",
			actual:   SortOrderAsc,
			expected: "asc",
		},
		{
			name:     "SortOrderDesc should be 'desc'",
			actual:   SortOrderDesc,
			expected: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actual) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.actual)
			}
		})
	}
}

// TestSortOrder_StringComparison validates SortOrder can be compared as strings
func TestSortOrder_StringComparison(t *testing.T) {
	// Arrange
	ascOrder := SortOrderAsc
	descOrder := SortOrderDesc

	// Act & Assert
	if ascOrder == descOrder {
		t.Error("SortOrderAsc should not equal SortOrderDesc")
	}

	if string(ascOrder) != "asc" {
		t.Errorf("Expected 'asc', got %q", ascOrder)
	}

	if string(descOrder) != "desc" {
		t.Errorf("Expected 'desc', got %q", descOrder)
	}
}

// TestSortOrder_Usage tests SortOrder in practical scenarios
func TestSortOrder_Usage(t *testing.T) {
	tests := []struct {
		name        string
		order       SortOrder
		expectedStr string
		isAsc       bool
		isDesc      bool
	}{
		{
			name:        "Ascending order usage",
			order:       SortOrderAsc,
			expectedStr: "asc",
			isAsc:       true,
			isDesc:      false,
		},
		{
			name:        "Descending order usage",
			order:       SortOrderDesc,
			expectedStr: "desc",
			isAsc:       false,
			isDesc:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			actualStr := string(tt.order)
			actualIsAsc := tt.order == SortOrderAsc
			actualIsDesc := tt.order == SortOrderDesc

			// Assert
			if actualStr != tt.expectedStr {
				t.Errorf("Expected string %q, got %q", tt.expectedStr, actualStr)
			}

			if actualIsAsc != tt.isAsc {
				t.Errorf("Expected isAsc %v, got %v", tt.isAsc, actualIsAsc)
			}

			if actualIsDesc != tt.isDesc {
				t.Errorf("Expected isDesc %v, got %v", tt.isDesc, actualIsDesc)
			}
		})
	}
}
