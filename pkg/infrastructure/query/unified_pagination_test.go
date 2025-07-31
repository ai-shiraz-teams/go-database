package query

import (
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func TestQueryParams_PrepareDefaults_UnifiedPagination(t *testing.T) {
	tests := []struct {
		name                   string
		initialOffset          int
		initialLimit           int
		initialPage            int
		initialPageSize        int
		expectedPage           int
		expectedPageSize       int
		expectedComputedOffset int
		expectedComputedLimit  int
	}{
		{
			name:                   "Offset/Limit priority - zero offset",
			initialOffset:          0,
			initialLimit:           10,
			initialPage:            5,
			initialPageSize:        15,
			expectedPage:           1,
			expectedPageSize:       10,
			expectedComputedOffset: 0,
			expectedComputedLimit:  10,
		},
		{
			name:                   "Offset/Limit priority - page 3",
			initialOffset:          20,
			initialLimit:           10,
			initialPage:            5,
			initialPageSize:        15,
			expectedPage:           3,
			expectedPageSize:       10,
			expectedComputedOffset: 20,
			expectedComputedLimit:  10,
		},
		{
			name:                   "Page/PageSize when no valid offset/limit",
			initialOffset:          -1,
			initialLimit:           0,
			initialPage:            4,
			initialPageSize:        25,
			expectedPage:           4,
			expectedPageSize:       25,
			expectedComputedOffset: 75,
			expectedComputedLimit:  25,
		},
		{
			name:                   "All defaults applied",
			initialOffset:          -1,
			initialLimit:           0,
			initialPage:            0,
			initialPageSize:        0,
			expectedPage:           1,
			expectedPageSize:       50,
			expectedComputedOffset: 0,
			expectedComputedLimit:  50,
		},
		{
			name:                   "Page size exceeds maximum",
			initialOffset:          -1,
			initialLimit:           0,
			initialPage:            2,
			initialPageSize:        500,
			expectedPage:           2,
			expectedPageSize:       200,
			expectedComputedOffset: 200,
			expectedComputedLimit:  200,
		},
		{
			name:                   "Offset/Limit with large values",
			initialOffset:          150,
			initialLimit:           30,
			initialPage:            10,
			initialPageSize:        20,
			expectedPage:           6,
			expectedPageSize:       30,
			expectedComputedOffset: 150,
			expectedComputedLimit:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			params := QueryParams[*testutil.TestEntity]{
				Page:     tt.initialPage,
				PageSize: tt.initialPageSize,
				Offset:   tt.initialOffset,
				Limit:    tt.initialLimit,
			}

			result := params.PrepareDefaults()

			if result != &params {
				t.Error("PrepareDefaults should return pointer to same instance")
			}

			if params.Page != tt.expectedPage {
				t.Errorf("Expected Page %d, got %d", tt.expectedPage, params.Page)
			}

			if params.PageSize != tt.expectedPageSize {
				t.Errorf("Expected PageSize %d, got %d", tt.expectedPageSize, params.PageSize)
			}

			if params.ComputedOffset != tt.expectedComputedOffset {
				t.Errorf("Expected ComputedOffset %d, got %d", tt.expectedComputedOffset, params.ComputedOffset)
			}

			if params.ComputedLimit != tt.expectedComputedLimit {
				t.Errorf("Expected ComputedLimit %d, got %d", tt.expectedComputedLimit, params.ComputedLimit)
			}
		})
	}
}

func TestQueryParams_NewQueryParams_UnifiedFields(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	if params.Page != 1 {
		t.Errorf("Expected Page 1, got %d", params.Page)
	}

	if params.PageSize != 50 {
		t.Errorf("Expected PageSize 50, got %d", params.PageSize)
	}

	if params.Offset != 0 {
		t.Errorf("Expected Offset 0, got %d", params.Offset)
	}

	if params.Limit != 0 {
		t.Errorf("Expected Limit 0, got %d", params.Limit)
	}

	if params.ComputedOffset != 0 {
		t.Errorf("Expected ComputedOffset 0, got %d", params.ComputedOffset)
	}

	if params.ComputedLimit != 0 {
		t.Errorf("Expected ComputedLimit 0, got %d", params.ComputedLimit)
	}
}

func TestQueryParams_OffsetLimitPriority(t *testing.T) {

	params := NewQueryParams[*testutil.TestEntity]()

	params.Offset = 30
	params.Limit = 15
	params.Page = 10
	params.PageSize = 5

	params.PrepareDefaults()

	expectedPage := (30 / 15) + 1
	expectedPageSize := 15

	if params.Page != expectedPage {
		t.Errorf("Expected Page %d (calculated from offset/limit), got %d", expectedPage, params.Page)
	}

	if params.PageSize != expectedPageSize {
		t.Errorf("Expected PageSize %d (from limit), got %d", expectedPageSize, params.PageSize)
	}

	expectedComputedOffset := (expectedPage - 1) * expectedPageSize
	if params.ComputedOffset != expectedComputedOffset {
		t.Errorf("Expected ComputedOffset %d, got %d", expectedComputedOffset, params.ComputedOffset)
	}

	if params.ComputedLimit != expectedPageSize {
		t.Errorf("Expected ComputedLimit %d, got %d", expectedPageSize, params.ComputedLimit)
	}
}

func TestQueryParams_PaginationBoundaryValidation(t *testing.T) {
	tests := []struct {
		name         string
		offset       int
		limit        int
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
	}{
		{
			name:         "Negative values default correctly",
			offset:       -1,
			limit:        -1,
			page:         -1,
			pageSize:     -1,
			expectedPage: 1,
			expectedSize: 50,
		},
		{
			name:         "Zero values default correctly",
			offset:       0,
			limit:        0,
			page:         0,
			pageSize:     0,
			expectedPage: 1,
			expectedSize: 50,
		},
		{
			name:         "Excessive page size is capped",
			offset:       -1,
			limit:        0,
			page:         1,
			pageSize:     1000,
			expectedPage: 1,
			expectedSize: 200,
		},
		{
			name:         "Page size at boundary is preserved",
			offset:       -1,
			limit:        0,
			page:         1,
			pageSize:     200,
			expectedPage: 1,
			expectedSize: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			params := QueryParams[*testutil.TestEntity]{
				Page:     tt.page,
				PageSize: tt.pageSize,
				Offset:   tt.offset,
				Limit:    tt.limit,
			}

			params.PrepareDefaults()

			if params.Page != tt.expectedPage {
				t.Errorf("Expected Page %d, got %d", tt.expectedPage, params.Page)
			}

			if params.PageSize != tt.expectedSize {
				t.Errorf("Expected PageSize %d, got %d", tt.expectedSize, params.PageSize)
			}
		})
	}
}

func TestQueryParams_CloneWithUnifiedPagination(t *testing.T) {

	original := QueryParams[*testutil.TestEntity]{
		Page:           3,
		PageSize:       25,
		Offset:         40,
		Limit:          20,
		ComputedOffset: 50,
		ComputedLimit:  25,
		Search:         "test",
	}

	cloned := original.Clone()

	if cloned.Page != original.Page {
		t.Errorf("Expected cloned Page %d, got %d", original.Page, cloned.Page)
	}

	if cloned.PageSize != original.PageSize {
		t.Errorf("Expected cloned PageSize %d, got %d", original.PageSize, cloned.PageSize)
	}

	if cloned.Offset != original.Offset {
		t.Errorf("Expected cloned Offset %d, got %d", original.Offset, cloned.Offset)
	}

	if cloned.Limit != original.Limit {
		t.Errorf("Expected cloned Limit %d, got %d", original.Limit, cloned.Limit)
	}

	if cloned.ComputedOffset != original.ComputedOffset {
		t.Errorf("Expected cloned ComputedOffset %d, got %d", original.ComputedOffset, cloned.ComputedOffset)
	}

	if cloned.ComputedLimit != original.ComputedLimit {
		t.Errorf("Expected cloned ComputedLimit %d, got %d", original.ComputedLimit, cloned.ComputedLimit)
	}

	if cloned == &original {
		t.Error("Clone should return a different instance")
	}
}

func TestQueryParams_RealWorldUsageScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		description string
		setup       func() *QueryParams[*testutil.TestEntity]
		validate    func(t *testing.T, params *QueryParams[*testutil.TestEntity])
	}{
		{
			name:        "API with offset/limit query params",
			description: "Simulates ?offset=20&limit=10 from API",
			setup: func() *QueryParams[*testutil.TestEntity] {
				params := NewQueryParams[*testutil.TestEntity]()

				params.Offset = 20
				params.Limit = 10
				return params
			},
			validate: func(t *testing.T, params *QueryParams[*testutil.TestEntity]) {
				params.PrepareDefaults()

				if params.Page != 3 || params.PageSize != 10 {
					t.Errorf("Expected page 3, size 10, got page %d, size %d", params.Page, params.PageSize)
				}
				if params.ComputedOffset != 20 || params.ComputedLimit != 10 {
					t.Errorf("Expected computed offset 20, limit 10, got %d, %d", params.ComputedOffset, params.ComputedLimit)
				}
			},
		},
		{
			name:        "API with page/page_size query params",
			description: "Simulates ?page=5&page_size=25 from API",
			setup: func() *QueryParams[*testutil.TestEntity] {
				params := NewQueryParams[*testutil.TestEntity]()

				params.Page = 5
				params.PageSize = 25
				return params
			},
			validate: func(t *testing.T, params *QueryParams[*testutil.TestEntity]) {
				params.PrepareDefaults()

				if params.Page != 5 || params.PageSize != 25 {
					t.Errorf("Expected page 5, size 25, got page %d, size %d", params.Page, params.PageSize)
				}
				if params.ComputedOffset != 100 || params.ComputedLimit != 25 {
					t.Errorf("Expected computed offset 100, limit 25, got %d, %d", params.ComputedOffset, params.ComputedLimit)
				}
			},
		},
		{
			name:        "API with no pagination params",
			description: "Simulates API call with no pagination query params",
			setup: func() *QueryParams[*testutil.TestEntity] {
				return NewQueryParams[*testutil.TestEntity]()
			},
			validate: func(t *testing.T, params *QueryParams[*testutil.TestEntity]) {
				params.PrepareDefaults()

				if params.Page != 1 || params.PageSize != 50 {
					t.Errorf("Expected page 1, size 50, got page %d, size %d", params.Page, params.PageSize)
				}
				if params.ComputedOffset != 0 || params.ComputedLimit != 50 {
					t.Errorf("Expected computed offset 0, limit 50, got %d, %d", params.ComputedOffset, params.ComputedLimit)
				}
			},
		},
		{
			name:        "API with mixed params - offset/limit priority",
			description: "Simulates ?offset=15&limit=5&page=10&page_size=20 (offset/limit wins)",
			setup: func() *QueryParams[*testutil.TestEntity] {
				params := NewQueryParams[*testutil.TestEntity]()

				params.Offset = 15
				params.Limit = 5
				params.Page = 10
				params.PageSize = 20
				return params
			},
			validate: func(t *testing.T, params *QueryParams[*testutil.TestEntity]) {
				params.PrepareDefaults()

				expectedPage := (15 / 5) + 1
				if params.Page != expectedPage || params.PageSize != 5 {
					t.Errorf("Expected page %d, size 5, got page %d, size %d", expectedPage, params.Page, params.PageSize)
				}
				if params.ComputedOffset != 15 || params.ComputedLimit != 5 {
					t.Errorf("Expected computed offset 15, limit 5, got %d, %d", params.ComputedOffset, params.ComputedLimit)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			params := scenario.setup()
			scenario.validate(t, params)
		})
	}
}
