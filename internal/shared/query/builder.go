package query

import (
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/identifier"
	"github.com/ai-shiraz-teams/go-database-sdk/internal/shared/types"
)

// NewQueryParams creates a new QueryParams instance with sensible defaults
func NewQueryParams[T types.IBaseModel]() *QueryParams[T] {
	return &QueryParams[T]{
		Page:           1,
		PageSize:       50,
		Search:         "",
		Sort:           make([]SortField, 0),
		Filters:        make([]identifier.FilterCriteria, 0),
		IncludeDeleted: false,
		OnlyDeleted:    false,
		Preloads:       make([]string, 0),
	}
}

// PrepareDefaults validates and sets default values for pagination parameters.
// It ensures page and page size are within acceptable bounds and calculates offset/limit.
func (qp *QueryParams[T]) PrepareDefaults() *QueryParams[T] {
	// Ensure minimum page number
	if qp.Page < 1 {
		qp.Page = 1
	}

	// Set default page size if not specified
	if qp.PageSize <= 0 {
		qp.PageSize = 50
	}

	// Cap maximum page size
	if qp.PageSize > 200 {
		qp.PageSize = 200
	}

	// Calculate offset and limit for database queries
	qp.Offset = (qp.Page - 1) * qp.PageSize
	qp.Limit = qp.PageSize

	// Initialize slices if nil
	if qp.Sort == nil {
		qp.Sort = make([]SortField, 0)
	}
	if qp.Filters == nil {
		qp.Filters = make([]identifier.FilterCriteria, 0)
	}
	if qp.Preloads == nil {
		qp.Preloads = make([]string, 0)
	}

	return qp
}

// WithFilters applies filter criteria from an IIdentifier to the QueryParams
func (qp *QueryParams[T]) WithFilters(identifier identifier.IIdentifier) *QueryParams[T] {
	if identifier != nil {
		qp.Filters = identifier.ToFilterCriteria()
	}
	return qp
}

// AddSort adds a sort field to the query parameters
func (qp *QueryParams[T]) AddSort(field string, order SortOrder) *QueryParams[T] {
	qp.Sort = append(qp.Sort, SortField{
		Field: field,
		Order: order,
	})
	return qp
}

// AddSortAsc adds an ascending sort field
func (qp *QueryParams[T]) AddSortAsc(field string) *QueryParams[T] {
	return qp.AddSort(field, SortOrderAsc)
}

// AddSortDesc adds a descending sort field
func (qp *QueryParams[T]) AddSortDesc(field string) *QueryParams[T] {
	return qp.AddSort(field, SortOrderDesc)
}

// ClearSort removes all sort fields
func (qp *QueryParams[T]) ClearSort() *QueryParams[T] {
	qp.Sort = make([]SortField, 0)
	return qp
}

// WithSearch sets the search term
func (qp *QueryParams[T]) WithSearch(searchTerm string) *QueryParams[T] {
	qp.Search = searchTerm
	return qp
}

// WithPreloads sets the preload relations
func (qp *QueryParams[T]) WithPreloads(preloads []string) *QueryParams[T] {
	qp.Preloads = preloads
	return qp
}

// AddPreload adds a preload relation to the existing list
func (qp *QueryParams[T]) AddPreload(preload string) *QueryParams[T] {
	qp.Preloads = append(qp.Preloads, preload)
	return qp
}

// WithDeletedVisibility sets the soft-delete visibility options
func (qp *QueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) *QueryParams[T] {
	qp.IncludeDeleted = includeDeleted
	qp.OnlyDeleted = onlyDeleted
	return qp
}

// IncludeDeletedRecords sets the query to include soft-deleted records
func (qp *QueryParams[T]) IncludeDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = true
	qp.OnlyDeleted = false
	return qp
}

// OnlyDeletedRecords sets the query to show only soft-deleted records
func (qp *QueryParams[T]) OnlyDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = false
	qp.OnlyDeleted = true
	return qp
}

// ExcludeDeletedRecords sets the query to exclude soft-deleted records (default behavior)
func (qp *QueryParams[T]) ExcludeDeletedRecords() *QueryParams[T] {
	qp.IncludeDeleted = false
	qp.OnlyDeleted = false
	return qp
}

// HasSearch returns true if a search term is provided
func (qp *QueryParams[T]) HasSearch() bool {
	return qp.Search != ""
}

// HasFilters returns true if any filters are applied
func (qp *QueryParams[T]) HasFilters() bool {
	return len(qp.Filters) > 0
}

// HasSort returns true if any sort fields are specified
func (qp *QueryParams[T]) HasSort() bool {
	return len(qp.Sort) > 0
}

// HasPreloads returns true if any preload relations are specified
func (qp *QueryParams[T]) HasPreloads() bool {
	return len(qp.Preloads) > 0
}

// Clone creates a deep copy of the QueryParams
func (qp *QueryParams[T]) Clone() *QueryParams[T] {
	newParams := &QueryParams[T]{
		Page:           qp.Page,
		PageSize:       qp.PageSize,
		Offset:         qp.Offset,
		Limit:          qp.Limit,
		Search:         qp.Search,
		IncludeDeleted: qp.IncludeDeleted,
		OnlyDeleted:    qp.OnlyDeleted,
	}

	// Deep copy slices
	if qp.Sort != nil {
		newParams.Sort = make([]SortField, len(qp.Sort))
		copy(newParams.Sort, qp.Sort)
	}

	if qp.Filters != nil {
		newParams.Filters = make([]identifier.FilterCriteria, len(qp.Filters))
		copy(newParams.Filters, qp.Filters)
	}

	if qp.Preloads != nil {
		newParams.Preloads = make([]string, len(qp.Preloads))
		copy(newParams.Preloads, qp.Preloads)
	}

	return newParams
}
