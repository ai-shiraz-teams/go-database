package unit_of_work

import (
	"context"
	"testing"

	"github.com/ai-shiraz-teams/go-database/pkg/infrastructure/query"
	"github.com/ai-shiraz-teams/go-database/pkg/testutil"
)

func TestPostgresUnitOfWork_UnifiedPaginationIntegration(t *testing.T) {

	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	entities := make([]*testutil.TestEntity, 20)
	for i := 0; i < 20; i++ {
		entities[i] = &testutil.TestEntity{
			Name:   "Entity " + string(rune(i+49)),
			Status: "active",
		}
		_, err := uow.Insert(ctx, entities[i])
		if err != nil {
			t.Fatalf("Failed to insert test entity %d: %v", i, err)
		}
	}

	tests := []struct {
		name                string
		setupQueryParams    func() *query.QueryParams[*testutil.TestEntity]
		expectedResultCount int
		expectedTotalCount  int64
		expectedFirstOffset int
		description         string
	}{
		{
			name: "Offset/Limit format - page 1",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Offset = 0
				params.Limit = 5
				return params
			},
			expectedResultCount: 5,
			expectedTotalCount:  20,
			expectedFirstOffset: 0,
			description:         "Using offset=0&limit=5 should return first 5 entities",
		},
		{
			name: "Offset/Limit format - page 3",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Offset = 10
				params.Limit = 5
				return params
			},
			expectedResultCount: 5,
			expectedTotalCount:  20,
			expectedFirstOffset: 10,
			description:         "Using offset=10&limit=5 should return entities 11-15",
		},
		{
			name: "Page/PageSize format - page 2",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 2
				params.PageSize = 8
				return params
			},
			expectedResultCount: 8,
			expectedTotalCount:  20,
			expectedFirstOffset: 8,
			description:         "Using page=2&page_size=8 should return entities 9-16",
		},
		{
			name: "Page/PageSize format - last page",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 3
				params.PageSize = 8
				return params
			},
			expectedResultCount: 4,
			expectedTotalCount:  20,
			expectedFirstOffset: 16,
			description:         "Using page=3&page_size=8 should return last 4 entities",
		},
		{
			name: "Mixed params - offset/limit priority",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()

				params.Offset = 8
				params.Limit = 4
				params.Page = 10
				params.PageSize = 20
				return params
			},
			expectedResultCount: 4,
			expectedTotalCount:  20,
			expectedFirstOffset: 8,
			description:         "When both formats provided, offset/limit should take priority",
		},
		{
			name: "Default pagination",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				return query.NewQueryParams[*testutil.TestEntity]()
			},
			expectedResultCount: 20,
			expectedTotalCount:  20,
			expectedFirstOffset: 0,
			description:         "Default pagination should show first page with default size",
		},
		{
			name: "Large page size capped",
			setupQueryParams: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 1
				params.PageSize = 500
				return params
			},
			expectedResultCount: 20,
			expectedTotalCount:  20,
			expectedFirstOffset: 0,
			description:         "Excessive page size should be capped at 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			queryParams := tt.setupQueryParams()

			results, total, err := uow.FindAllWithPagination(ctx, queryParams)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if len(results) != tt.expectedResultCount {
				t.Errorf("Expected %d results, got %d. %s", tt.expectedResultCount, len(results), tt.description)
			}

			if total != tt.expectedTotalCount {
				t.Errorf("Expected total %d, got %d. %s", tt.expectedTotalCount, total, tt.description)
			}

			if queryParams.ComputedOffset != tt.expectedFirstOffset {
				t.Errorf("Expected computed offset %d, got %d. %s", tt.expectedFirstOffset, queryParams.ComputedOffset, tt.description)
			}

			if len(results) > 0 {

				firstEntityIndex := tt.expectedFirstOffset
				for i, result := range results {
					expectedIndex := firstEntityIndex + i
					if expectedIndex < len(entities) {

						if result.GetID() <= 0 {
							t.Errorf("Entity %d has invalid ID %d", i, result.GetID())
						}
					}
				}
			}

			t.Logf("✓ %s: Got %d results with total %d (offset: %d)",
				tt.description, len(results), total, queryParams.ComputedOffset)
		})
	}
}

func TestPostgresUnitOfWork_PaginationConsistency(t *testing.T) {

	db := testutil.SetupTestDB(t)
	uow := NewPostgresUnitOfWork[*testutil.TestEntity](db)
	ctx := context.Background()

	for i := 0; i < 15; i++ {
		entity := &testutil.TestEntity{
			Name:   "Consistency Test " + string(rune(i+49)),
			Status: "active",
		}
		_, err := uow.Insert(ctx, entity)
		if err != nil {
			t.Fatalf("Failed to insert test entity %d: %v", i, err)
		}
	}

	equivalentTests := []struct {
		name        string
		offsetLimit *query.QueryParams[*testutil.TestEntity]
		pageSize    *query.QueryParams[*testutil.TestEntity]
		description string
	}{
		{
			name: "First page equivalence",
			offsetLimit: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Offset = 0
				params.Limit = 5
				return params
			}(),
			pageSize: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 1
				params.PageSize = 5
				return params
			}(),
			description: "offset=0&limit=5 should equal page=1&page_size=5",
		},
		{
			name: "Second page equivalence",
			offsetLimit: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Offset = 5
				params.Limit = 5
				return params
			}(),
			pageSize: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 2
				params.PageSize = 5
				return params
			}(),
			description: "offset=5&limit=5 should equal page=2&page_size=5",
		},
		{
			name: "Third page equivalence",
			offsetLimit: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Offset = 10
				params.Limit = 5
				return params
			}(),
			pageSize: func() *query.QueryParams[*testutil.TestEntity] {
				params := query.NewQueryParams[*testutil.TestEntity]()
				params.Page = 3
				params.PageSize = 5
				return params
			}(),
			description: "offset=10&limit=5 should equal page=3&page_size=5",
		},
	}

	for _, tt := range equivalentTests {
		t.Run(tt.name, func(t *testing.T) {

			offsetResults, offsetTotal, offsetErr := uow.FindAllWithPagination(ctx, tt.offsetLimit)
			pageResults, pageTotal, pageErr := uow.FindAllWithPagination(ctx, tt.pageSize)

			if offsetErr != nil {
				t.Fatalf("Offset/limit query failed: %v", offsetErr)
			}
			if pageErr != nil {
				t.Fatalf("Page/pageSize query failed: %v", pageErr)
			}

			if len(offsetResults) != len(pageResults) {
				t.Errorf("Result count mismatch: offset/limit got %d, page/pageSize got %d. %s",
					len(offsetResults), len(pageResults), tt.description)
			}

			if offsetTotal != pageTotal {
				t.Errorf("Total count mismatch: offset/limit got %d, page/pageSize got %d. %s",
					offsetTotal, pageTotal, tt.description)
			}

			if len(offsetResults) == len(pageResults) {
				for i := 0; i < len(offsetResults); i++ {
					if offsetResults[i].GetID() != pageResults[i].GetID() {
						t.Errorf("Entity mismatch at index %d: offset/limit got ID %d, page/pageSize got ID %d. %s",
							i, offsetResults[i].GetID(), pageResults[i].GetID(), tt.description)
					}
				}
			}

			t.Logf("✓ %s: Both formats returned %d entities consistently",
				tt.description, len(offsetResults))
		})
	}
}
