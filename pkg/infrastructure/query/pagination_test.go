package query

import (
	"testing"
)

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name             string
		offset           int
		limit            int
		page             int
		pageSize         int
		expectedPage     int
		expectedPageSize int
	}{
		{
			name:             "Offset/Limit takes priority - zero offset",
			offset:           0,
			limit:            10,
			page:             5,
			pageSize:         15,
			expectedPage:     1,
			expectedPageSize: 10,
		},
		{
			name:             "Offset/Limit takes priority - non-zero offset",
			offset:           20,
			limit:            10,
			page:             5,
			pageSize:         15,
			expectedPage:     3,
			expectedPageSize: 10,
		},
		{
			name:             "Offset/Limit takes priority - large offset",
			offset:           100,
			limit:            25,
			page:             2,
			pageSize:         10,
			expectedPage:     5,
			expectedPageSize: 25,
		},
		{
			name:             "Page/PageSize when no valid offset/limit",
			offset:           -1,
			limit:            0,
			page:             5,
			pageSize:         15,
			expectedPage:     5,
			expectedPageSize: 15,
		},
		{
			name:             "Page/PageSize when limit is zero",
			offset:           10,
			limit:            0,
			page:             3,
			pageSize:         20,
			expectedPage:     3,
			expectedPageSize: 20,
		},
		{
			name:             "Page/PageSize when offset is negative",
			offset:           -5,
			limit:            10,
			page:             4,
			pageSize:         12,
			expectedPage:     4,
			expectedPageSize: 12,
		},
		{
			name:             "Default values when all params are invalid",
			offset:           -1,
			limit:            0,
			page:             0,
			pageSize:         0,
			expectedPage:     1,
			expectedPageSize: 50,
		},
		{
			name:             "Default page when page is zero",
			offset:           -1,
			limit:            0,
			page:             0,
			pageSize:         25,
			expectedPage:     1,
			expectedPageSize: 25,
		},
		{
			name:             "Default page size when page size is negative",
			offset:           -1,
			limit:            0,
			page:             3,
			pageSize:         -5,
			expectedPage:     3,
			expectedPageSize: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualPage, actualPageSize := NormalizePagination(tt.offset, tt.limit, tt.page, tt.pageSize)

			if actualPage != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, actualPage)
			}

			if actualPageSize != tt.expectedPageSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedPageSize, actualPageSize)
			}
		})
	}
}

func TestCalculateOffsetLimit(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "First page",
			page:           1,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "Second page",
			page:           2,
			pageSize:       10,
			expectedOffset: 10,
			expectedLimit:  10,
		},
		{
			name:           "Fifth page with page size 25",
			page:           5,
			pageSize:       25,
			expectedOffset: 100,
			expectedLimit:  25,
		},
		{
			name:           "Large page number",
			page:           100,
			pageSize:       50,
			expectedOffset: 4950,
			expectedLimit:  50,
		},
		{
			name:           "Zero page defaults to 1",
			page:           0,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "Negative page defaults to 1",
			page:           -5,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "Zero page size defaults to 50",
			page:           3,
			pageSize:       0,
			expectedOffset: 100,
			expectedLimit:  50,
		},
		{
			name:           "Negative page size defaults to 50",
			page:           2,
			pageSize:       -10,
			expectedOffset: 50,
			expectedLimit:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualOffset, actualLimit := CalculateOffsetLimit(tt.page, tt.pageSize)

			if actualOffset != tt.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tt.expectedOffset, actualOffset)
			}

			if actualLimit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, actualLimit)
			}
		})
	}
}

func TestValidatePaginationBounds(t *testing.T) {
	tests := []struct {
		name             string
		page             int
		pageSize         int
		maxPageSize      int
		expectedPage     int
		expectedPageSize int
	}{
		{
			name:             "Valid values within bounds",
			page:             5,
			pageSize:         25,
			maxPageSize:      200,
			expectedPage:     5,
			expectedPageSize: 25,
		},
		{
			name:             "Page size exceeds maximum",
			page:             2,
			pageSize:         500,
			maxPageSize:      200,
			expectedPage:     2,
			expectedPageSize: 200,
		},
		{
			name:             "Page size at maximum boundary",
			page:             3,
			pageSize:         200,
			maxPageSize:      200,
			expectedPage:     3,
			expectedPageSize: 200,
		},
		{
			name:             "Zero page defaults to 1",
			page:             0,
			pageSize:         25,
			maxPageSize:      200,
			expectedPage:     1,
			expectedPageSize: 25,
		},
		{
			name:             "Negative page defaults to 1",
			page:             -3,
			pageSize:         25,
			maxPageSize:      200,
			expectedPage:     1,
			expectedPageSize: 25,
		},
		{
			name:             "Zero page size defaults to 50",
			page:             2,
			pageSize:         0,
			maxPageSize:      200,
			expectedPage:     2,
			expectedPageSize: 50,
		},
		{
			name:             "Negative page size defaults to 50",
			page:             2,
			pageSize:         -10,
			maxPageSize:      200,
			expectedPage:     2,
			expectedPageSize: 50,
		},
		{
			name:             "Zero max page size uses default 200",
			page:             1,
			pageSize:         300,
			maxPageSize:      0,
			expectedPage:     1,
			expectedPageSize: 200,
		},
		{
			name:             "Negative max page size uses default 200",
			page:             1,
			pageSize:         300,
			maxPageSize:      -50,
			expectedPage:     1,
			expectedPageSize: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualPage, actualPageSize := ValidatePaginationBounds(tt.page, tt.pageSize, tt.maxPageSize)

			if actualPage != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, actualPage)
			}

			if actualPageSize != tt.expectedPageSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedPageSize, actualPageSize)
			}
		})
	}
}

func TestPaginationIntegration(t *testing.T) {
	tests := []struct {
		name                   string
		inputOffset            int
		inputLimit             int
		inputPage              int
		inputPageSize          int
		expectedFinalOffset    int
		expectedFinalLimit     int
		expectedNormalizedPage int
		expectedNormalizedSize int
	}{
		{
			name:                   "Offset/limit priority integration",
			inputOffset:            30,
			inputLimit:             15,
			inputPage:              5,
			inputPageSize:          10,
			expectedFinalOffset:    30,
			expectedFinalLimit:     15,
			expectedNormalizedPage: 3,
			expectedNormalizedSize: 15,
		},
		{
			name:                   "Page/pageSize when no valid offset/limit",
			inputOffset:            -1,
			inputLimit:             0,
			inputPage:              4,
			inputPageSize:          20,
			expectedFinalOffset:    60,
			expectedFinalLimit:     20,
			expectedNormalizedPage: 4,
			expectedNormalizedSize: 20,
		},
		{
			name:                   "All defaults applied",
			inputOffset:            -1,
			inputLimit:             0,
			inputPage:              0,
			inputPageSize:          0,
			expectedFinalOffset:    0,
			expectedFinalLimit:     50,
			expectedNormalizedPage: 1,
			expectedNormalizedSize: 50,
		},
		{
			name:                   "Boundary validation integration",
			inputOffset:            -1,
			inputLimit:             0,
			inputPage:              2,
			inputPageSize:          500,
			expectedFinalOffset:    200,
			expectedFinalLimit:     200,
			expectedNormalizedPage: 2,
			expectedNormalizedSize: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			normalizedPage, normalizedPageSize := NormalizePagination(tt.inputOffset, tt.inputLimit, tt.inputPage, tt.inputPageSize)

			validatedPage, validatedPageSize := ValidatePaginationBounds(normalizedPage, normalizedPageSize, 200)

			finalOffset, finalLimit := CalculateOffsetLimit(validatedPage, validatedPageSize)

			if validatedPage != tt.expectedNormalizedPage {
				t.Errorf("Expected normalized page %d, got %d", tt.expectedNormalizedPage, validatedPage)
			}

			if validatedPageSize != tt.expectedNormalizedSize {
				t.Errorf("Expected normalized page size %d, got %d", tt.expectedNormalizedSize, validatedPageSize)
			}

			if finalOffset != tt.expectedFinalOffset {
				t.Errorf("Expected final offset %d, got %d", tt.expectedFinalOffset, finalOffset)
			}

			if finalLimit != tt.expectedFinalLimit {
				t.Errorf("Expected final limit %d, got %d", tt.expectedFinalLimit, finalLimit)
			}
		})
	}
}
