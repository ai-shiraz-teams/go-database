package query

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
)

// TestQueryParams_WithPreloads validates preload relations setting
func TestQueryParams_WithPreloads(t *testing.T) {
	tests := []struct {
		name     string
		preloads []string
		expected []string
	}{
		{
			name:     "Single preload",
			preloads: []string{"User"},
			expected: []string{"User"},
		},
		{
			name:     "Multiple preloads",
			preloads: []string{"User", "Category", "Tags"},
			expected: []string{"User", "Category", "Tags"},
		},
		{
			name:     "Empty preloads",
			preloads: []string{},
			expected: []string{},
		},
		{
			name:     "Nil preloads",
			preloads: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			params := NewQueryParams[*MockEntity]()

			// Act
			result := params.WithPreloads(tt.preloads)

			// Assert
			if result != params {
				t.Error("WithPreloads should return pointer to same instance")
			}

			if len(params.Preloads) != len(tt.expected) {
				t.Errorf("Expected %d preloads, got %d", len(tt.expected), len(params.Preloads))
			}

			for i, expected := range tt.expected {
				if params.Preloads[i] != expected {
					t.Errorf("Expected preload[%d] %q, got %q", i, expected, params.Preloads[i])
				}
			}
		})
	}
}

// TestQueryParams_AddPreload validates individual preload addition
func TestQueryParams_AddPreload(t *testing.T) {
	// Arrange
	params := NewQueryParams[*MockEntity]()

	// Act & Assert - Add first preload
	result := params.AddPreload("User")

	if result != params {
		t.Error("AddPreload should return pointer to same instance")
	}

	if len(params.Preloads) != 1 {
		t.Errorf("Expected 1 preload, got %d", len(params.Preloads))
	}

	if params.Preloads[0] != "User" {
		t.Errorf("Expected preload 'User', got %q", params.Preloads[0])
	}

	// Add second preload
	params.AddPreload("Category")

	if len(params.Preloads) != 2 {
		t.Errorf("Expected 2 preloads, got %d", len(params.Preloads))
	}

	if params.Preloads[1] != "Category" {
		t.Errorf("Expected preload 'Category', got %q", params.Preloads[1])
	}
}

// TestQueryParams_WithDeletedVisibility validates soft-delete visibility options
func TestQueryParams_WithDeletedVisibility(t *testing.T) {
	tests := []struct {
		name            string
		includeDeleted  bool
		onlyDeleted     bool
		expectedInclude bool
		expectedOnly    bool
	}{
		{
			name:            "Include and only deleted both false",
			includeDeleted:  false,
			onlyDeleted:     false,
			expectedInclude: false,
			expectedOnly:    false,
		},
		{
			name:            "Include deleted true, only deleted false",
			includeDeleted:  true,
			onlyDeleted:     false,
			expectedInclude: true,
			expectedOnly:    false,
		},
		{
			name:            "Include deleted false, only deleted true",
			includeDeleted:  false,
			onlyDeleted:     true,
			expectedInclude: false,
			expectedOnly:    true,
		},
		{
			name:            "Both include and only deleted true",
			includeDeleted:  true,
			onlyDeleted:     true,
			expectedInclude: true,
			expectedOnly:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			params := NewQueryParams[*MockEntity]()

			// Act
			result := params.WithDeletedVisibility(tt.includeDeleted, tt.onlyDeleted)

			// Assert
			if result != params {
				t.Error("WithDeletedVisibility should return pointer to same instance")
			}

			if params.IncludeDeleted != tt.expectedInclude {
				t.Errorf("Expected IncludeDeleted %v, got %v", tt.expectedInclude, params.IncludeDeleted)
			}

			if params.OnlyDeleted != tt.expectedOnly {
				t.Errorf("Expected OnlyDeleted %v, got %v", tt.expectedOnly, params.OnlyDeleted)
			}
		})
	}
}

// TestQueryParams_IncludeDeletedRecords validates include deleted records setting
func TestQueryParams_IncludeDeletedRecords(t *testing.T) {
	// Arrange
	params := NewQueryParams[*MockEntity]()
	params.OnlyDeleted = true // Set initial state

	// Act
	result := params.IncludeDeletedRecords()

	// Assert
	if result != params {
		t.Error("IncludeDeletedRecords should return pointer to same instance")
	}

	if !params.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be true")
	}

	if params.OnlyDeleted {
		t.Error("Expected OnlyDeleted to be false")
	}
}

// TestQueryParams_OnlyDeletedRecords validates only deleted records setting
func TestQueryParams_OnlyDeletedRecords(t *testing.T) {
	// Arrange
	params := NewQueryParams[*MockEntity]()
	params.IncludeDeleted = true // Set initial state

	// Act
	result := params.OnlyDeletedRecords()

	// Assert
	if result != params {
		t.Error("OnlyDeletedRecords should return pointer to same instance")
	}

	if params.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be false")
	}

	if !params.OnlyDeleted {
		t.Error("Expected OnlyDeleted to be true")
	}
}

// TestQueryParams_ExcludeDeletedRecords validates exclude deleted records setting
func TestQueryParams_ExcludeDeletedRecords(t *testing.T) {
	// Arrange
	params := NewQueryParams[*MockEntity]()
	params.IncludeDeleted = true
	params.OnlyDeleted = true

	// Act
	result := params.ExcludeDeletedRecords()

	// Assert
	if result != params {
		t.Error("ExcludeDeletedRecords should return pointer to same instance")
	}

	if params.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be false")
	}

	if params.OnlyDeleted {
		t.Error("Expected OnlyDeleted to be false")
	}
}

// TestQueryParams_HasSearch validates search term detection
func TestQueryParams_HasSearch(t *testing.T) {
	tests := []struct {
		name       string
		searchTerm string
		expected   bool
	}{
		{
			name:       "Non-empty search term",
			searchTerm: "john",
			expected:   true,
		},
		{
			name:       "Empty search term",
			searchTerm: "",
			expected:   false,
		},
		{
			name:       "Whitespace only search term",
			searchTerm: "   ",
			expected:   true, // Non-empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			params := NewQueryParams[*MockEntity]()
			params.Search = tt.searchTerm

			// Act
			result := params.HasSearch()

			// Assert
			if result != tt.expected {
				t.Errorf("Expected HasSearch() %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestQueryParams_HasFilters validates filter detection
func TestQueryParams_HasFilters(t *testing.T) {
	// Arrange & Act - No filters
	params := NewQueryParams[*MockEntity]()
	result := params.HasFilters()

	// Assert
	if result {
		t.Error("Expected HasFilters() to be false when no filters are present")
	}

	// Add filters and test again
	params.Filters = make([]identifier.FilterCriteria, 1)
	result = params.HasFilters()

	if !result {
		t.Error("Expected HasFilters() to be true when filters are present")
	}
}

// TestQueryParams_HasSort validates sort detection
func TestQueryParams_HasSort(t *testing.T) {
	// Arrange & Act - No sort
	params := NewQueryParams[*MockEntity]()
	result := params.HasSort()

	// Assert
	if result {
		t.Error("Expected HasSort() to be false when no sort fields are present")
	}

	// Add sort and test again
	params.AddSort("name", SortOrderAsc)
	result = params.HasSort()

	if !result {
		t.Error("Expected HasSort() to be true when sort fields are present")
	}
}

// TestQueryParams_HasPreloads validates preload detection
func TestQueryParams_HasPreloads(t *testing.T) {
	// Arrange & Act - No preloads
	params := NewQueryParams[*MockEntity]()
	result := params.HasPreloads()

	// Assert
	if result {
		t.Error("Expected HasPreloads() to be false when no preloads are present")
	}

	// Add preload and test again
	params.AddPreload("User")
	result = params.HasPreloads()

	if !result {
		t.Error("Expected HasPreloads() to be true when preloads are present")
	}
}

// TestQueryParams_Clone validates deep copying
func TestQueryParams_Clone(t *testing.T) {
	// Arrange
	original := NewQueryParams[*MockEntity]()
	original.Page = 2
	original.PageSize = 25
	original.Offset = 25
	original.Limit = 25
	original.Search = "test search"
	original.IncludeDeleted = true
	original.OnlyDeleted = false
	original.AddSort("name", SortOrderAsc)
	original.AddSort("created_at", SortOrderDesc)
	original.AddPreload("User")
	original.AddPreload("Category")

	// Create some filter criteria
	original.Filters = make([]identifier.FilterCriteria, 2)
	original.Filters[0] = identifier.FilterCriteria{Field: "active", Value: true}
	original.Filters[1] = identifier.FilterCriteria{Field: "type", Value: "premium"}

	// Act
	clone := original.Clone()

	// Assert - Basic fields
	if clone == original {
		t.Error("Clone should return a different instance")
	}

	if clone.Page != original.Page {
		t.Errorf("Expected cloned Page %d, got %d", original.Page, clone.Page)
	}

	if clone.PageSize != original.PageSize {
		t.Errorf("Expected cloned PageSize %d, got %d", original.PageSize, clone.PageSize)
	}

	if clone.Search != original.Search {
		t.Errorf("Expected cloned Search %q, got %q", original.Search, clone.Search)
	}

	if clone.IncludeDeleted != original.IncludeDeleted {
		t.Errorf("Expected cloned IncludeDeleted %v, got %v", original.IncludeDeleted, clone.IncludeDeleted)
	}

	// Assert - Slice independence (deep copy)
	if &clone.Sort == &original.Sort {
		t.Error("Clone should have independent Sort slice")
	}

	if len(clone.Sort) != len(original.Sort) {
		t.Errorf("Expected cloned Sort length %d, got %d", len(original.Sort), len(clone.Sort))
	}

	// Modify original and verify clone is unaffected
	original.Sort[0].Field = "modified"
	if clone.Sort[0].Field == "modified" {
		t.Error("Modifying original Sort should not affect clone")
	}

	// Test preloads independence
	if &clone.Preloads == &original.Preloads {
		t.Error("Clone should have independent Preloads slice")
	}

	original.Preloads[0] = "Modified"
	if clone.Preloads[0] == "Modified" {
		t.Error("Modifying original Preloads should not affect clone")
	}

	// Test filters independence
	if &clone.Filters == &original.Filters {
		t.Error("Clone should have independent Filters slice")
	}

	original.Filters[0].Field = "modified"
	if clone.Filters[0].Field == "modified" {
		t.Error("Modifying original Filters should not affect clone")
	}
}

// TestQueryParams_Clone_NilSlices validates cloning with nil slices
func TestQueryParams_Clone_NilSlices(t *testing.T) {
	// Arrange
	original := &QueryParams[*MockEntity]{
		Page:     1,
		PageSize: 50,
		Sort:     nil,
		Filters:  nil,
		Preloads: nil,
	}

	// Act
	clone := original.Clone()

	// Assert
	if clone.Sort != nil {
		t.Error("Expected cloned Sort to be nil when original is nil")
	}

	if clone.Filters != nil {
		t.Error("Expected cloned Filters to be nil when original is nil")
	}

	if clone.Preloads != nil {
		t.Error("Expected cloned Preloads to be nil when original is nil")
	}
}
