package query

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func TestNewQueryParams(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	if params == nil {
		t.Fatal("NewQueryParams returned nil")
	}

	if params.Page != 1 {
		t.Errorf("Expected Page 1, got %d", params.Page)
	}

	if params.PageSize != 50 {
		t.Errorf("Expected PageSize 50, got %d", params.PageSize)
	}

	if params.Search != "" {
		t.Errorf("Expected empty Search, got %q", params.Search)
	}

	if params.Sort == nil {
		t.Error("Expected Sort slice to be initialized")
	}

	if len(params.Sort) != 0 {
		t.Errorf("Expected empty Sort slice, got %d items", len(params.Sort))
	}

	if params.Filters == nil {
		t.Error("Expected Filters slice to be initialized")
	}

	if len(params.Filters) != 0 {
		t.Errorf("Expected empty Filters slice, got %d items", len(params.Filters))
	}

	if params.IncludeDeleted {
		t.Error("Expected IncludeDeleted to be false")
	}

	if params.OnlyDeleted {
		t.Error("Expected OnlyDeleted to be false")
	}

	if params.Preloads == nil {
		t.Error("Expected Preloads slice to be initialized")
	}

	if len(params.Preloads) != 0 {
		t.Errorf("Expected empty Preloads slice, got %d items", len(params.Preloads))
	}
}

func TestQueryParams_PrepareDefaults(t *testing.T) {
	tests := []struct {
		name           string
		initialParams  QueryParams[*testutil.TestEntity]
		expectedPage   int
		expectedSize   int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name: "Valid page and size",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     3,
				PageSize: 25,
			},
			expectedPage:   3,
			expectedSize:   25,
			expectedOffset: 50,
			expectedLimit:  25,
		},
		{
			name: "Zero page (should default to 1)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     0,
				PageSize: 10,
			},
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name: "Negative page (should default to 1)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     -5,
				PageSize: 10,
			},
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name: "Zero page size (should default to 50)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     2,
				PageSize: 0,
			},
			expectedPage:   2,
			expectedSize:   50,
			expectedOffset: 50,
			expectedLimit:  50,
		},
		{
			name: "Negative page size (should default to 50)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     1,
				PageSize: -10,
			},
			expectedPage:   1,
			expectedSize:   50,
			expectedOffset: 0,
			expectedLimit:  50,
		},
		{
			name: "Page size too large (should cap at 200)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     1,
				PageSize: 500,
			},
			expectedPage:   1,
			expectedSize:   200,
			expectedOffset: 0,
			expectedLimit:  200,
		},
		{
			name: "Page size at limit (200)",
			initialParams: QueryParams[*testutil.TestEntity]{
				Page:     2,
				PageSize: 200,
			},
			expectedPage:   2,
			expectedSize:   200,
			expectedOffset: 200,
			expectedLimit:  200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			params := tt.initialParams

			result := params.PrepareDefaults()

			if result != &params {
				t.Error("PrepareDefaults should return pointer to same instance")
			}

			if params.Page != tt.expectedPage {
				t.Errorf("Expected Page %d, got %d", tt.expectedPage, params.Page)
			}

			if params.PageSize != tt.expectedSize {
				t.Errorf("Expected PageSize %d, got %d", tt.expectedSize, params.PageSize)
			}

			if params.ComputedOffset != tt.expectedOffset {
				t.Errorf("Expected ComputedOffset %d, got %d", tt.expectedOffset, params.ComputedOffset)
			}

			if params.ComputedLimit != tt.expectedLimit {
				t.Errorf("Expected ComputedLimit %d, got %d", tt.expectedLimit, params.ComputedLimit)
			}

			if params.Sort == nil {
				t.Error("Expected Sort slice to be initialized")
			}

			if params.Filters == nil {
				t.Error("Expected Filters slice to be initialized")
			}

			if params.Preloads == nil {
				t.Error("Expected Preloads slice to be initialized")
			}
		})
	}
}

func TestQueryParams_PrepareDefaults_NilSlices(t *testing.T) {

	params := QueryParams[*testutil.TestEntity]{
		Page:     1,
		PageSize: 10,
		Sort:     nil,
		Filters:  nil,
		Preloads: nil,
	}

	params.PrepareDefaults()

	if params.Sort == nil {
		t.Error("Expected Sort slice to be initialized")
	}

	if len(params.Sort) != 0 {
		t.Errorf("Expected empty Sort slice, got %d items", len(params.Sort))
	}

	if params.Filters == nil {
		t.Error("Expected Filters slice to be initialized")
	}

	if len(params.Filters) != 0 {
		t.Errorf("Expected empty Filters slice, got %d items", len(params.Filters))
	}

	if params.Preloads == nil {
		t.Error("Expected Preloads slice to be initialized")
	}

	if len(params.Preloads) != 0 {
		t.Errorf("Expected empty Preloads slice, got %d items", len(params.Preloads))
	}
}

func TestQueryParams_AddSort(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	result := params.AddSort("name", SortOrderAsc)

	if result != params {
		t.Error("AddSort should return pointer to same instance")
	}

	if len(params.Sort) != 1 {
		t.Errorf("Expected 1 sort field, got %d", len(params.Sort))
	}

	sortField := params.Sort[0]
	if sortField.Field != "name" {
		t.Errorf("Expected Field 'name', got %q", sortField.Field)
	}

	if sortField.Order != SortOrderAsc {
		t.Errorf("Expected Order %q, got %q", SortOrderAsc, sortField.Order)
	}

	params.AddSort("created_at", SortOrderDesc)

	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}

	secondSort := params.Sort[1]
	if secondSort.Field != "created_at" {
		t.Errorf("Expected Field 'created_at', got %q", secondSort.Field)
	}

	if secondSort.Order != SortOrderDesc {
		t.Errorf("Expected Order %q, got %q", SortOrderDesc, secondSort.Order)
	}
}

func TestQueryParams_AddSortAsc(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	result := params.AddSortAsc("email")

	if result != params {
		t.Error("AddSortAsc should return pointer to same instance")
	}

	if len(params.Sort) != 1 {
		t.Errorf("Expected 1 sort field, got %d", len(params.Sort))
	}

	sortField := params.Sort[0]
	if sortField.Field != "email" {
		t.Errorf("Expected Field 'email', got %q", sortField.Field)
	}

	if sortField.Order != SortOrderAsc {
		t.Errorf("Expected Order %q, got %q", SortOrderAsc, sortField.Order)
	}
}

func TestQueryParams_AddSortDesc(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	result := params.AddSortDesc("updated_at")

	if result != params {
		t.Error("AddSortDesc should return pointer to same instance")
	}

	if len(params.Sort) != 1 {
		t.Errorf("Expected 1 sort field, got %d", len(params.Sort))
	}

	sortField := params.Sort[0]
	if sortField.Field != "updated_at" {
		t.Errorf("Expected Field 'updated_at', got %q", sortField.Field)
	}

	if sortField.Order != SortOrderDesc {
		t.Errorf("Expected Order %q, got %q", SortOrderDesc, sortField.Order)
	}
}

func TestQueryParams_ClearSort(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()
	params.AddSort("name", SortOrderAsc)
	params.AddSort("email", SortOrderDesc)

	if len(params.Sort) != 2 {
		t.Fatalf("Expected 2 sort fields, got %d", len(params.Sort))
	}

	result := params.ClearSort()

	if result != params {
		t.Error("ClearSort should return pointer to same instance")
	}

	if len(params.Sort) != 0 {
		t.Errorf("Expected 0 sort fields after clear, got %d", len(params.Sort))
	}

	if params.Sort == nil {
		t.Error("Expected Sort slice to remain initialized after clear")
	}
}

func TestQueryParams_WithSearch(t *testing.T) {
	tests := []struct {
		name       string
		searchTerm string
	}{
		{
			name:       "Simple search term",
			searchTerm: "john",
		},
		{
			name:       "Multi-word search term",
			searchTerm: "john doe",
		},
		{
			name:       "Empty search term",
			searchTerm: "",
		},
		{
			name:       "Special characters search",
			searchTerm: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			params := NewQueryParams[*testutil.TestEntity]()

			result := params.WithSearch(tt.searchTerm)

			if result != params {
				t.Error("WithSearch should return pointer to same instance")
			}

			if params.Search != tt.searchTerm {
				t.Errorf("Expected Search %q, got %q", tt.searchTerm, params.Search)
			}
		})
	}
}
