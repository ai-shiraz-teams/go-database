package http

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

func TestQueryParams_PrepareDefaults(t *testing.T) {
	tests := []struct {
		name           string
		initialParams  QueryParams
		expectedOffset int
		expectedLimit  int
	}{
		{
			name: "Valid offset and limit",
			initialParams: QueryParams{
				Offset: 3,
				Limit:  25,
			},
			expectedOffset: 3,
			expectedLimit:  25,
		},
		{
			name: "Zero offset (should default to 1)",
			initialParams: QueryParams{
				Offset: 0,
				Limit:  10,
			},
			expectedOffset: 1,
			expectedLimit:  10,
		},
		{
			name: "Negative offset (should default to 1)",
			initialParams: QueryParams{
				Offset: -5,
				Limit:  10,
			},
			expectedOffset: 1,
			expectedLimit:  10,
		},
		{
			name: "Zero limit (should default to 50)",
			initialParams: QueryParams{
				Offset: 2,
				Limit:  0,
			},
			expectedOffset: 2,
			expectedLimit:  50,
		},
		{
			name: "Negative limit (should default to 50)",
			initialParams: QueryParams{
				Offset: 1,
				Limit:  -10,
			},
			expectedOffset: 1,
			expectedLimit:  50,
		},
		{
			name: "Excessive limit (should be capped at 200)",
			initialParams: QueryParams{
				Offset: 1,
				Limit:  500,
			},
			expectedOffset: 1,
			expectedLimit:  200,
		},
		{
			name: "Limit at boundary (should be preserved)",
			initialParams: QueryParams{
				Offset: 2,
				Limit:  200,
			},
			expectedOffset: 2,
			expectedLimit:  200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.initialParams.PrepareDefaults()

			// Assert
			if err != nil {
				t.Errorf("PrepareDefaults() returned error: %v", err)
			}

			if tt.initialParams.Offset != tt.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tt.expectedOffset, tt.initialParams.Offset)
			}

			if tt.initialParams.Limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, tt.initialParams.Limit)
			}

			// Verify IsValidated returns true after PrepareDefaults
			if !tt.initialParams.IsValidated() {
				t.Error("Expected IsValidated() to return true after PrepareDefaults()")
			}
		})
	}
}

func TestQueryParams_parseSortFields(t *testing.T) {
	tests := []struct {
		name     string
		sortStr  string
		expected []domain.SortField
	}{
		{
			name:     "Empty sort string",
			sortStr:  "",
			expected: nil,
		},
		{
			name:    "Single field ascending (default)",
			sortStr: "name",
			expected: []domain.SortField{
				{Field: "name", Order: domain.SortOrderAsc},
			},
		},
		{
			name:    "Single field descending with minus prefix",
			sortStr: "-created_at",
			expected: []domain.SortField{
				{Field: "created_at", Order: domain.SortOrderDesc},
			},
		},
		{
			name:    "Single field with explicit asc",
			sortStr: "name:asc",
			expected: []domain.SortField{
				{Field: "name", Order: domain.SortOrderAsc},
			},
		},
		{
			name:    "Single field with explicit desc",
			sortStr: "created_at:desc",
			expected: []domain.SortField{
				{Field: "created_at", Order: domain.SortOrderDesc},
			},
		},
		{
			name:    "Multiple fields mixed formats",
			sortStr: "name:asc,-created_at,status:desc",
			expected: []domain.SortField{
				{Field: "name", Order: domain.SortOrderAsc},
				{Field: "created_at", Order: domain.SortOrderDesc},
				{Field: "status", Order: domain.SortOrderDesc},
			},
		},
		{
			name:    "Multiple fields with spaces",
			sortStr: " name : asc , -created_at , status : desc ",
			expected: []domain.SortField{
				{Field: "name", Order: domain.SortOrderAsc},
				{Field: "created_at", Order: domain.SortOrderDesc},
				{Field: "status", Order: domain.SortOrderDesc},
			},
		},
		{
			name:    "Invalid format (empty after split)",
			sortStr: "name,,status",
			expected: []domain.SortField{
				{Field: "name", Order: domain.SortOrderAsc},
				{Field: "status", Order: domain.SortOrderAsc},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			qp := QueryParams{Sort: tt.sortStr}

			// Act
			result := qp.parseSortFields()

			// Assert
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d sort fields, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing sort field at index %d", i)
					continue
				}

				if result[i].Field != expected.Field {
					t.Errorf("Expected field '%s', got '%s'", expected.Field, result[i].Field)
				}

				if result[i].Order != expected.Order {
					t.Errorf("Expected order '%s', got '%s'", expected.Order, result[i].Order)
				}
			}
		})
	}
}

func TestQueryParams_ToSimpleQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		params   QueryParams
		validate func(*testing.T, domain.IQueryParams[domain.IBaseModel])
	}{
		{
			name: "Basic conversion with defaults",
			params: QueryParams{
				Limit:  25,
				Offset: 2,
			},
			validate: func(t *testing.T, domainParams domain.IQueryParams[domain.IBaseModel]) {
				if domainParams.Limit() != 25 {
					t.Errorf("Expected limit 25, got %d", domainParams.Limit())
				}
				if domainParams.Offset() != 2 {
					t.Errorf("Expected offset 2, got %d", domainParams.Offset())
				}
			},
		},
		{
			name: "With soft-delete flags",
			params: QueryParams{
				Limit:          10,
				Offset:         1,
				IncludeDeleted: true,
				OnlyDeleted:    false,
			},
			validate: func(t *testing.T, domainParams domain.IQueryParams[domain.IBaseModel]) {
				if !domainParams.GetIncludeDeleted() {
					t.Error("Expected IncludeDeleted to be true")
				}
				if domainParams.GetOnlyDeleted() {
					t.Error("Expected OnlyDeleted to be false")
				}
			},
		},
		{
			name: "With preloads",
			params: QueryParams{
				Limit:    10,
				Offset:   1,
				Preloads: "user,profile,comments",
			},
			validate: func(t *testing.T, domainParams domain.IQueryParams[domain.IBaseModel]) {
				preloads := domainParams.Preloads()
				expectedPreloads := []string{"user", "profile", "comments"}

				if len(preloads) != len(expectedPreloads) {
					t.Errorf("Expected %d preloads, got %d", len(expectedPreloads), len(preloads))
					return
				}

				for i, expected := range expectedPreloads {
					if preloads[i] != expected {
						t.Errorf("Expected preload '%s', got '%s'", expected, preloads[i])
					}
				}
			},
		},
		{
			name: "With sort fields",
			params: QueryParams{
				Limit:  10,
				Offset: 1,
				Sort:   "name:asc,-created_at",
			},
			validate: func(t *testing.T, domainParams domain.IQueryParams[domain.IBaseModel]) {
				if !domainParams.HasSort() {
					t.Error("Expected HasSort() to be true")
				}

				sortFields := domainParams.ToSortFields()
				if len(sortFields) != 2 {
					t.Errorf("Expected 2 sort fields, got %d", len(sortFields))
					return
				}

				if sortFields[0].Field != "name" || sortFields[0].Order != domain.SortOrderAsc {
					t.Errorf("Expected first sort field: name:ASC, got %s:%s", sortFields[0].Field, sortFields[0].Order)
				}

				if sortFields[1].Field != "created_at" || sortFields[1].Order != domain.SortOrderDesc {
					t.Errorf("Expected second sort field: created_at:DESC, got %s:%s", sortFields[1].Field, sortFields[1].Order)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			err := tt.params.PrepareDefaults()
			if err != nil {
				t.Fatalf("Failed to prepare defaults: %v", err)
			}

			// Act
			domainParams := tt.params.ToSimpleQueryParams()

			// Assert
			if domainParams == nil {
				t.Fatal("ToSimpleQueryParams() returned nil")
			}

			tt.validate(t, domainParams)
		})
	}
}

func TestQueryParams_GetPaginationOffset(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		expectedOffset int
	}{
		{
			name:           "Offset 1 should return 0 for database",
			offset:         1,
			expectedOffset: 0,
		},
		{
			name:           "Offset 2 should return 1 for database",
			offset:         2,
			expectedOffset: 1,
		},
		{
			name:           "Offset 10 should return 9 for database",
			offset:         10,
			expectedOffset: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			qp := QueryParams{Offset: tt.offset}

			// Act
			result := qp.GetPaginationOffset()

			// Assert
			if result != tt.expectedOffset {
				t.Errorf("Expected pagination offset %d, got %d", tt.expectedOffset, result)
			}
		})
	}
}

func TestQueryParams_IsValidated(t *testing.T) {
	tests := []struct {
		name     string
		params   QueryParams
		expected bool
	}{
		{
			name: "Valid params after PrepareDefaults",
			params: QueryParams{
				Offset: 1,
				Limit:  50,
			},
			expected: true,
		},
		{
			name: "Invalid offset (0)",
			params: QueryParams{
				Offset: 0,
				Limit:  50,
			},
			expected: false,
		},
		{
			name: "Invalid limit (0)",
			params: QueryParams{
				Offset: 1,
				Limit:  0,
			},
			expected: false,
		},
		{
			name: "Invalid limit (exceeds 200)",
			params: QueryParams{
				Offset: 1,
				Limit:  250,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.params.IsValidated()

			// Assert
			if result != tt.expected {
				t.Errorf("Expected IsValidated() to return %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestQueryParams_String(t *testing.T) {
	// Arrange
	qp := QueryParams{
		Filter:         "active",
		Sort:           "name:asc",
		Limit:          25,
		Offset:         2,
		IncludeDeleted: true,
		OnlyDeleted:    false,
		Preloads:       "user,profile",
	}

	// Act
	result := qp.String()

	// Assert
	expected := "QueryParams{Filter:active, Sort:name:asc, Limit:25, Offset:2, IncludeDeleted:true, OnlyDeleted:false, Preloads:user,profile}"
	if result != expected {
		t.Errorf("Expected string representation:\n%s\nGot:\n%s", expected, result)
	}
}
