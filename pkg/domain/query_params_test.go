package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewQueryParams(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test default values
	if params.Page != 1 {
		t.Errorf("Expected default Page 1, got %d", params.Page)
	}

	if params.PageSize != 50 {
		t.Errorf("Expected default PageSize 50, got %d", params.PageSize)
	}

	if params.Search != "" {
		t.Errorf("Expected empty Search, got %s", params.Search)
	}

	if params.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be false by default")
	}

	if params.OnlyDeleted {
		t.Error("Expected OnlyDeleted to be false by default")
	}

	if params.Sort == nil {
		t.Error("Expected Sort to be initialized")
	}

	if params.Filters == nil {
		t.Error("Expected Filters to be initialized")
	}

	if params.Preloads == nil {
		t.Error("Expected Preloads to be initialized")
	}
}

func TestQueryParams_PrepareDefaults(t *testing.T) {
	tests := []struct {
		name           string
		initialPage    int
		initialSize    int
		defaultLimit   int
		maxLimit       int
		expectedPage   int
		expectedSize   int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "Valid values",
			initialPage:    2,
			initialSize:    25,
			defaultLimit:   50,
			maxLimit:       100,
			expectedPage:   2,
			expectedSize:   25,
			expectedOffset: 25,
			expectedLimit:  25,
		},
		{
			name:           "Page below minimum",
			initialPage:    0,
			initialSize:    25,
			defaultLimit:   50,
			maxLimit:       100,
			expectedPage:   1,
			expectedSize:   25,
			expectedOffset: 0,
			expectedLimit:  25,
		},
		{
			name:           "PageSize zero uses default",
			initialPage:    1,
			initialSize:    0,
			defaultLimit:   50,
			maxLimit:       100,
			expectedPage:   1,
			expectedSize:   50,
			expectedOffset: 0,
			expectedLimit:  50,
		},
		{
			name:           "PageSize exceeds maximum",
			initialPage:    1,
			initialSize:    150,
			defaultLimit:   50,
			maxLimit:       100,
			expectedPage:   1,
			expectedSize:   100,
			expectedOffset: 0,
			expectedLimit:  100,
		},
		{
			name:           "Page 3 with size 20",
			initialPage:    3,
			initialSize:    20,
			defaultLimit:   50,
			maxLimit:       100,
			expectedPage:   3,
			expectedSize:   20,
			expectedOffset: 40,
			expectedLimit:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &QueryParams[*User]{
				Page:     tt.initialPage,
				PageSize: tt.initialSize,
			}

			params.PrepareDefaults(tt.defaultLimit, tt.maxLimit)

			if params.Page != tt.expectedPage {
				t.Errorf("Expected Page %d, got %d", tt.expectedPage, params.Page)
			}

			if params.PageSize != tt.expectedSize {
				t.Errorf("Expected PageSize %d, got %d", tt.expectedSize, params.PageSize)
			}

			if params.Offset != tt.expectedOffset {
				t.Errorf("Expected Offset %d, got %d", tt.expectedOffset, params.Offset)
			}

			if params.Limit != tt.expectedLimit {
				t.Errorf("Expected Limit %d, got %d", tt.expectedLimit, params.Limit)
			}
		})
	}
}

func TestQueryParams_WithFilters(t *testing.T) {
	params := NewQueryParams[*User]()
	filter := NewIdentifier().Equal("status", "active").GreaterThan("age", 18)

	params.WithFilters(filter)

	if len(params.Filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(params.Filters))
	}

	// Test with nil filter
	params.WithFilters(nil)
	if len(params.Filters) != 0 {
		t.Error("WithFilters(nil) should clear filters")
	}
}

func TestQueryParams_Sorting(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test AddSort
	params.AddSort("name", SortOrderAsc)
	if len(params.Sort) != 1 {
		t.Errorf("Expected 1 sort field, got %d", len(params.Sort))
	}

	if params.Sort[0].Field != "name" {
		t.Errorf("Expected sort field 'name', got %s", params.Sort[0].Field)
	}

	if params.Sort[0].Order != SortOrderAsc {
		t.Errorf("Expected sort order 'asc', got %s", params.Sort[0].Order)
	}

	// Test AddSortDesc
	params.AddSortDesc("createdAt")
	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}

	if params.Sort[1].Order != SortOrderDesc {
		t.Errorf("Expected second sort order 'desc', got %s", params.Sort[1].Order)
	}

	// Test ClearSort
	params.ClearSort()
	if len(params.Sort) != 0 {
		t.Error("ClearSort should remove all sort fields")
	}

	// Test AddSortAsc
	params.AddSortAsc("email")
	if len(params.Sort) != 1 || params.Sort[0].Order != SortOrderAsc {
		t.Error("AddSortAsc should add ascending sort field")
	}
}

func TestQueryParams_Search(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test initial state
	if params.HasSearch() {
		t.Error("HasSearch should return false initially")
	}

	// Test WithSearch
	params.WithSearch("john doe")
	if params.Search != "john doe" {
		t.Errorf("Expected search 'john doe', got %s", params.Search)
	}

	if !params.HasSearch() {
		t.Error("HasSearch should return true after setting search")
	}
}

func TestQueryParams_Preloads(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test initial state
	if params.HasPreloads() {
		t.Error("HasPreloads should return false initially")
	}

	// Test WithPreloads
	params.WithPreloads("Profile", "Orders")
	if len(params.Preloads) != 2 {
		t.Errorf("Expected 2 preloads, got %d", len(params.Preloads))
	}

	if !params.HasPreloads() {
		t.Error("HasPreloads should return true after setting preloads")
	}

	// Test AddPreload
	params.AddPreload("Settings")
	if len(params.Preloads) != 3 {
		t.Errorf("Expected 3 preloads after AddPreload, got %d", len(params.Preloads))
	}

	if params.Preloads[2] != "Settings" {
		t.Errorf("Expected third preload 'Settings', got %s", params.Preloads[2])
	}
}

func TestQueryParams_DeletedVisibility(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test initial state (exclude deleted by default)
	if params.IncludeDeleted || params.OnlyDeleted {
		t.Error("Both IncludeDeleted and OnlyDeleted should be false initially")
	}

	// Test IncludeDeletedRecords
	params.IncludeDeletedRecords()
	if !params.IncludeDeleted || params.OnlyDeleted {
		t.Error("IncludeDeletedRecords should set IncludeDeleted=true, OnlyDeleted=false")
	}

	// Test OnlyDeletedRecords
	params.OnlyDeletedRecords()
	if params.IncludeDeleted || !params.OnlyDeleted {
		t.Error("OnlyDeletedRecords should set IncludeDeleted=false, OnlyDeleted=true")
	}

	// Test ExcludeDeletedRecords
	params.ExcludeDeletedRecords()
	if params.IncludeDeleted || params.OnlyDeleted {
		t.Error("ExcludeDeletedRecords should set both flags to false")
	}

	// Test WithDeletedVisibility
	params.WithDeletedVisibility(true, false)
	if !params.IncludeDeleted || params.OnlyDeleted {
		t.Error("WithDeletedVisibility(true, false) failed")
	}
}

func TestQueryParams_HasMethods(t *testing.T) {
	params := NewQueryParams[*User]()

	// Test initial state - all should be false/empty
	if params.HasFilters() {
		t.Error("HasFilters should return false initially")
	}

	if params.HasSort() {
		t.Error("HasSort should return false initially")
	}

	if params.HasSearch() {
		t.Error("HasSearch should return false initially")
	}

	if params.HasPreloads() {
		t.Error("HasPreloads should return false initially")
	}

	// Add content and test again
	params.WithFilters(NewIdentifier().Equal("status", "active"))
	params.AddSortAsc("name")
	params.WithSearch("test")
	params.AddPreload("Profile")

	if !params.HasFilters() {
		t.Error("HasFilters should return true after adding filters")
	}

	if !params.HasSort() {
		t.Error("HasSort should return true after adding sort")
	}

	if !params.HasSearch() {
		t.Error("HasSearch should return true after adding search")
	}

	if !params.HasPreloads() {
		t.Error("HasPreloads should return true after adding preloads")
	}
}

func TestQueryParams_ToListOptions(t *testing.T) {
	params := NewQueryParams[*User]()
	params.Page = 2
	params.PageSize = 25
	params.PrepareDefaults(50, 100)
	params.WithFilters(NewIdentifier().Equal("status", "active"))
	params.AddSortDesc("createdAt")
	params.IncludeDeletedRecords()

	opts := params.ToListOptions()

	if opts.Limit != 25 {
		t.Errorf("Expected Limit 25, got %d", opts.Limit)
	}

	if opts.Offset != 25 {
		t.Errorf("Expected Offset 25, got %d", opts.Offset)
	}

	if len(opts.Filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(opts.Filters))
	}

	if opts.SortBy != "createdAt" {
		t.Errorf("Expected SortBy 'createdAt', got %s", opts.SortBy)
	}

	if opts.SortOrder != "desc" {
		t.Errorf("Expected SortOrder 'desc', got %s", opts.SortOrder)
	}

	if !opts.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be true")
	}

	// Test with no sort - should default to id asc
	params.ClearSort()
	opts = params.ToListOptions()

	if opts.SortBy != "id" {
		t.Errorf("Expected default SortBy 'id', got %s", opts.SortBy)
	}

	if opts.SortOrder != "asc" {
		t.Errorf("Expected default SortOrder 'asc', got %s", opts.SortOrder)
	}
}

func TestQueryParams_Clone(t *testing.T) {
	original := NewQueryParams[*User]()
	original.Page = 2
	original.PageSize = 25
	original.Search = "test search"
	original.AddSortAsc("name")
	original.AddSortDesc("createdAt")
	original.WithFilters(NewIdentifier().Equal("status", "active"))
	original.AddPreload("Profile")
	original.AddPreload("Orders")
	original.IncludeDeletedRecords()

	clone := original.Clone()

	// Test that values are copied
	if clone.Page != original.Page {
		t.Error("Page not cloned correctly")
	}

	if clone.Search != original.Search {
		t.Error("Search not cloned correctly")
	}

	if clone.IncludeDeleted != original.IncludeDeleted {
		t.Error("IncludeDeleted not cloned correctly")
	}

	// Test that slices are deep copied
	if len(clone.Sort) != len(original.Sort) {
		t.Error("Sort slice not cloned correctly")
	}

	if len(clone.Filters) != len(original.Filters) {
		t.Error("Filters slice not cloned correctly")
	}

	if len(clone.Preloads) != len(original.Preloads) {
		t.Error("Preloads slice not cloned correctly")
	}

	// Test that modifying clone doesn't affect original
	clone.AddSortAsc("email")
	if len(original.Sort) == len(clone.Sort) {
		t.Error("Modifying clone affected original Sort slice")
	}

	clone.AddPreload("Settings")
	if len(original.Preloads) == len(clone.Preloads) {
		t.Error("Modifying clone affected original Preloads slice")
	}
}

func TestQueryParams_JSONSerialization(t *testing.T) {
	params := NewQueryParams[*User]()
	params.Page = 2
	params.PageSize = 25
	params.Search = "test"
	params.AddSortAsc("name")
	params.WithFilters(NewIdentifier().Equal("status", "active"))
	params.AddPreload("Profile")
	params.IncludeDeletedRecords()

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal QueryParams: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled QueryParams[*User]
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal QueryParams: %v", err)
	}

	// Test that values are preserved
	if unmarshaled.Page != params.Page {
		t.Error("Page not preserved in JSON serialization")
	}

	if unmarshaled.Search != params.Search {
		t.Error("Search not preserved in JSON serialization")
	}

	if len(unmarshaled.Sort) != len(params.Sort) {
		t.Error("Sort not preserved in JSON serialization")
	}

	if len(unmarshaled.Filters) != len(params.Filters) {
		t.Error("Filters not preserved in JSON serialization")
	}

	// Note: Offset and Limit should not be serialized (json:"-" tag)
	if unmarshaled.Offset != 0 {
		t.Error("Offset should not be serialized")
	}

	if unmarshaled.Limit != 0 {
		t.Error("Limit should not be serialized")
	}
}

func TestSortField_JSONSerialization(t *testing.T) {
	field := SortField{
		Field: "name",
		Order: SortOrderDesc,
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal SortField: %v", err)
	}

	var unmarshaled SortField
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SortField: %v", err)
	}

	if !reflect.DeepEqual(field, unmarshaled) {
		t.Errorf("SortField serialization failed: expected %+v, got %+v", field, unmarshaled)
	}
}

func TestQueryParams_ChainableMethods(t *testing.T) {
	// Test that methods can be chained fluently
	params := NewQueryParams[*User]().
		WithSearch("john").
		AddSortAsc("name").
		AddSortDesc("createdAt").
		WithPreloads("Profile", "Orders").
		AddPreload("Settings").
		IncludeDeletedRecords().
		WithFilters(NewIdentifier().Equal("status", "active"))

	// Verify all operations were applied
	if params.Search != "john" {
		t.Error("Chained WithSearch failed")
	}

	if len(params.Sort) != 2 {
		t.Error("Chained sort methods failed")
	}

	if len(params.Preloads) != 3 {
		t.Error("Chained preload methods failed")
	}

	if !params.IncludeDeleted {
		t.Error("Chained IncludeDeletedRecords failed")
	}

	if len(params.Filters) != 1 {
		t.Error("Chained WithFilters failed")
	}
}

func TestQueryParams_WithSpecificEntityTypes(t *testing.T) {
	// Test that QueryParams works with different entity types
	userParams := NewQueryParams[*User]()
	productParams := NewQueryParams[*Product]()
	categoryParams := NewQueryParams[*Category]()

	// These should compile and work without issues
	userParams.WithSearch("user search")
	productParams.WithSearch("product search")
	categoryParams.WithSearch("category search")

	// Verify they are independent instances
	if userParams.Search == productParams.Search {
		t.Error("Different entity type params should be independent")
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkQueryParams_NewQueryParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewQueryParams[*User]()
	}
}

func BenchmarkQueryParams_PrepareDefaults(b *testing.B) {
	params := NewQueryParams[*User]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.PrepareDefaults(50, 100)
	}
}

func BenchmarkQueryParams_AddSort(b *testing.B) {
	params := NewQueryParams[*User]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.AddSort("field", SortOrderAsc)
	}
}

func BenchmarkQueryParams_Clone(b *testing.B) {
	params := NewQueryParams[*User]()
	params.AddSortAsc("name")
	params.WithFilters(NewIdentifier().Equal("status", "active"))
	params.AddPreload("Profile")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = params.Clone()
	}
}
