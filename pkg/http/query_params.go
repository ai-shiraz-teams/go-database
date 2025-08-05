package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ai-shiraz-teams/go-database/pkg/domain"
)

// QueryParams represents HTTP query string parameters for filtering, pagination, and sorting
type QueryParams struct {
	Filter string `form:"filter" json:"filter,omitempty"`
	Sort   string `form:"sort" json:"sort,omitempty"`
	Limit  int    `form:"limit" json:"limit,omitempty" binding:"omitempty,min=1,max=200"`
	Offset int    `form:"offset" json:"offset,omitempty" binding:"omitempty,min=1"`

	// Soft-delete visibility options
	IncludeDeleted bool `form:"include_deleted" json:"includeDeleted,omitempty"`
	OnlyDeleted    bool `form:"only_deleted" json:"onlyDeleted,omitempty"`

	// Preload relations
	Preloads string `form:"preloads" json:"preloads,omitempty"`
}

// PrepareDefaults validates and sets default values for QueryParams
func (qp *QueryParams) PrepareDefaults() error {
	// Set default offset to 1 if not specified or invalid
	if qp.Offset <= 0 {
		qp.Offset = 1
	}

	// Set default limit to 50 if not specified or invalid
	if qp.Limit <= 0 {
		qp.Limit = 50
	}

	// Cap maximum limit at 200
	if qp.Limit > 200 {
		qp.Limit = 200
	}

	return nil
}

// ValidateExplicitParams validates parameters that were explicitly provided
// This method should be called with the raw query values to distinguish
// between "not provided" and "explicitly set to invalid value"
func (qp *QueryParams) ValidateExplicitParams(rawQuery map[string][]string) error {
	// Check if offset was explicitly provided with invalid value
	if offsetValues, exists := rawQuery["offset"]; exists && len(offsetValues) > 0 {
		if offsetValue, err := strconv.Atoi(offsetValues[0]); err == nil {
			if offsetValue <= 0 {
				return fmt.Errorf("offset must be at least 1, got %d", offsetValue)
			}
		} else {
			return fmt.Errorf("offset must be a valid number")
		}
	}

	// Check if limit was explicitly provided with invalid value
	if limitValues, exists := rawQuery["limit"]; exists && len(limitValues) > 0 {
		if limitValue, err := strconv.Atoi(limitValues[0]); err == nil {
			if limitValue <= 0 {
				return fmt.Errorf("limit must be at least 1, got %d", limitValue)
			}
			if limitValue > 200 {
				return fmt.Errorf("limit cannot exceed 200, got %d", limitValue)
			}
		} else {
			return fmt.Errorf("limit must be a valid number")
		}
	}

	return nil
}

// ToSimpleQueryParams converts HTTP QueryParams to domain SimpleQueryParams
func (qp *QueryParams) ToSimpleQueryParams() domain.IQueryParams[domain.IBaseModel] {
	// Create base query params
	params := domain.NewSimpleQueryParams[domain.IBaseModel]()

	// Set limit and offset
	params = params.WithLimit(qp.Limit).WithOffset(qp.Offset)

	// Set soft-delete visibility
	params = params.WithDeletedVisibility(qp.IncludeDeleted, qp.OnlyDeleted)

	// Parse and set preloads
	if qp.Preloads != "" {
		preloadList := strings.Split(qp.Preloads, ",")
		for i, preload := range preloadList {
			preloadList[i] = strings.TrimSpace(preload)
		}
		params = params.WithPreloads(preloadList)
	}

	// Parse and set sort fields
	if qp.Sort != "" {
		sortFields := qp.parseSortFields()
		if len(sortFields) > 0 {
			params = params.WithSort(sortFields)
		}
	}

	return params
}

// parseSortFields parses sort string into SortField slice
// Expected format: "field1:asc,field2:desc" or "field1,-field2" (- prefix for desc)
func (qp *QueryParams) parseSortFields() []domain.SortField {
	if qp.Sort == "" {
		return nil
	}

	var sortFields []domain.SortField
	sortParts := strings.Split(qp.Sort, ",")

	for _, part := range sortParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var field string
		var order domain.SortOrder = domain.SortOrderAsc

		// Handle "field:order" format
		if strings.Contains(part, ":") {
			fieldOrder := strings.Split(part, ":")
			if len(fieldOrder) == 2 {
				field = strings.TrimSpace(fieldOrder[0])
				orderStr := strings.ToLower(strings.TrimSpace(fieldOrder[1]))
				if orderStr == "desc" {
					order = domain.SortOrderDesc
				}
			}
		} else if strings.HasPrefix(part, "-") {
			// Handle "-field" format for descending
			field = part[1:]
			order = domain.SortOrderDesc
		} else {
			// Default ascending
			field = part
		}

		if field != "" {
			sortFields = append(sortFields, domain.SortField{
				Field: field,
				Order: order,
			})
		}
	}

	return sortFields
}

// IsValidated returns true if the QueryParams have been validated and prepared
func (qp *QueryParams) IsValidated() bool {
	return qp.Offset >= 1 && qp.Limit >= 1 && qp.Limit <= 200
}

// GetFilterValue returns the filter as a string for basic filtering
func (qp *QueryParams) GetFilterValue() string {
	return qp.Filter
}

// GetPaginationOffset calculates the database offset from 1-based page offset
func (qp *QueryParams) GetPaginationOffset() int {
	return qp.Offset - 1
}

// String returns a string representation of the QueryParams for debugging
func (qp *QueryParams) String() string {
	return "QueryParams{" +
		"Filter:" + qp.Filter +
		", Sort:" + qp.Sort +
		", Limit:" + strconv.Itoa(qp.Limit) +
		", Offset:" + strconv.Itoa(qp.Offset) +
		", IncludeDeleted:" + strconv.FormatBool(qp.IncludeDeleted) +
		", OnlyDeleted:" + strconv.FormatBool(qp.OnlyDeleted) +
		", Preloads:" + qp.Preloads +
		"}"
}
