package domain

// SortOrder defines the direction for sorting operations
type SortOrder string

const (
	// SortOrderAsc represents ascending sort order
	SortOrderAsc SortOrder = "asc"
	// SortOrderDesc represents descending sort order
	SortOrderDesc SortOrder = "desc"
)

// SortField represents a single field to sort by with its direction
type SortField struct {
	// Field is the name of the field to sort by
	Field string `json:"field"`
	// Order is the direction to sort (asc/desc)
	Order SortOrder `json:"order"`
}

// QueryParams provides a typed, reusable structure for paginated repository access.
// It supports pagination, filtering, sorting, search, preloading, and soft-delete visibility.
// This struct is designed to be compatible with Echo query binding and JSON serialization.
type QueryParams[T IBaseModel] struct {
	// Pagination (1-based page numbering)
	Page     int `json:"page" query:"page"`           // 1-based page number
	PageSize int `json:"page_size" query:"page_size"` // Number of items per page

	// Calculated pagination values (internal use, not serialized)
	Offset int `json:"-"` // Calculated: (Page - 1) * PageSize
	Limit  int `json:"-"` // Same as PageSize

	// Sorting - supports multiple sort fields
	Sort []SortField `json:"sort"`

	// Filtering using the FilterCriteria system
	Filters []FilterCriteria `json:"filters"`

	// Search term for text-based searching across relevant fields
	Search string `json:"search" query:"search"`

	// Soft-delete visibility controls
	IncludeDeleted bool `json:"include_deleted" query:"include_deleted"` // Include soft-deleted records
	OnlyDeleted    bool `json:"only_deleted" query:"only_deleted"`       // Show only soft-deleted records

	// Preload relations - specify relation names to eagerly load
	Preloads []string `json:"preloads"`
}

// NewQueryParams creates a new QueryParams instance with sensible defaults
func NewQueryParams[T IBaseModel]() *QueryParams[T] {
	return &QueryParams[T]{
		Page:           1,
		PageSize:       50,
		Sort:           make([]SortField, 0),
		Filters:        make([]FilterCriteria, 0),
		Search:         "",
		IncludeDeleted: false,
		OnlyDeleted:    false,
		Preloads:       make([]string, 0),
	}
}

// PrepareDefaults validates and sets default values for pagination parameters.
// It ensures page and page size are within acceptable bounds and calculates offset/limit.
func (q *QueryParams[T]) PrepareDefaults(defaultLimit int, maxLimit int) {
	// Ensure page is at least 1
	if q.Page < 1 {
		q.Page = 1
	}

	// Set default page size if not provided or zero
	if q.PageSize <= 0 {
		q.PageSize = defaultLimit
	} else if q.PageSize > maxLimit {
		// Cap page size at maximum allowed
		q.PageSize = maxLimit
	}

	// Calculate offset and limit for database queries
	q.Offset = (q.Page - 1) * q.PageSize
	q.Limit = q.PageSize
}

// WithFilters applies filter criteria from an IIdentifier to the QueryParams
func (q *QueryParams[T]) WithFilters(identifier IIdentifier) *QueryParams[T] {
	if identifier != nil {
		q.Filters = identifier.ToFilterCriteria()
	} else {
		q.Filters = make([]FilterCriteria, 0)
	}
	return q
}

// AddSort adds a sort field to the query parameters
func (q *QueryParams[T]) AddSort(field string, order SortOrder) *QueryParams[T] {
	if q.Sort == nil {
		q.Sort = make([]SortField, 0)
	}
	q.Sort = append(q.Sort, SortField{
		Field: field,
		Order: order,
	})
	return q
}

// AddSortAsc adds an ascending sort field
func (q *QueryParams[T]) AddSortAsc(field string) *QueryParams[T] {
	return q.AddSort(field, SortOrderAsc)
}

// AddSortDesc adds a descending sort field
func (q *QueryParams[T]) AddSortDesc(field string) *QueryParams[T] {
	return q.AddSort(field, SortOrderDesc)
}

// ClearSort removes all sort fields
func (q *QueryParams[T]) ClearSort() *QueryParams[T] {
	q.Sort = make([]SortField, 0)
	return q
}

// WithSearch sets the search term
func (q *QueryParams[T]) WithSearch(searchTerm string) *QueryParams[T] {
	q.Search = searchTerm
	return q
}

// WithPreloads sets the preload relations
func (q *QueryParams[T]) WithPreloads(preloads ...string) *QueryParams[T] {
	q.Preloads = preloads
	return q
}

// AddPreload adds a preload relation to the existing list
func (q *QueryParams[T]) AddPreload(preload string) *QueryParams[T] {
	if q.Preloads == nil {
		q.Preloads = make([]string, 0)
	}
	q.Preloads = append(q.Preloads, preload)
	return q
}

// WithDeletedVisibility sets the soft-delete visibility options
func (q *QueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) *QueryParams[T] {
	q.IncludeDeleted = includeDeleted
	q.OnlyDeleted = onlyDeleted
	return q
}

// IncludeDeletedRecords sets the query to include soft-deleted records
func (q *QueryParams[T]) IncludeDeletedRecords() *QueryParams[T] {
	q.IncludeDeleted = true
	q.OnlyDeleted = false
	return q
}

// OnlyDeletedRecords sets the query to show only soft-deleted records
func (q *QueryParams[T]) OnlyDeletedRecords() *QueryParams[T] {
	q.IncludeDeleted = false
	q.OnlyDeleted = true
	return q
}

// ExcludeDeletedRecords sets the query to exclude soft-deleted records (default behavior)
func (q *QueryParams[T]) ExcludeDeletedRecords() *QueryParams[T] {
	q.IncludeDeleted = false
	q.OnlyDeleted = false
	return q
}

// ToListOptions converts QueryParams to the legacy ListOptions format for backward compatibility
func (q *QueryParams[T]) ToListOptions() *ListOptions {
	opts := &ListOptions{
		Limit:          q.Limit,
		Offset:         q.Offset,
		Filters:        q.Filters,
		IncludeDeleted: q.IncludeDeleted,
	}

	// Convert sorting - use first sort field for legacy format
	if len(q.Sort) > 0 {
		opts.SortBy = q.Sort[0].Field
		opts.SortOrder = string(q.Sort[0].Order)
	} else {
		opts.SortBy = "id"
		opts.SortOrder = "asc"
	}

	return opts
}

// HasSearch returns true if a search term is provided
func (q *QueryParams[T]) HasSearch() bool {
	return q.Search != ""
}

// HasFilters returns true if any filters are applied
func (q *QueryParams[T]) HasFilters() bool {
	return len(q.Filters) > 0
}

// HasSort returns true if any sort fields are specified
func (q *QueryParams[T]) HasSort() bool {
	return len(q.Sort) > 0
}

// HasPreloads returns true if any preload relations are specified
func (q *QueryParams[T]) HasPreloads() bool {
	return len(q.Preloads) > 0
}

// Clone creates a deep copy of the QueryParams
func (q *QueryParams[T]) Clone() *QueryParams[T] {
	clone := &QueryParams[T]{
		Page:           q.Page,
		PageSize:       q.PageSize,
		Offset:         q.Offset,
		Limit:          q.Limit,
		Search:         q.Search,
		IncludeDeleted: q.IncludeDeleted,
		OnlyDeleted:    q.OnlyDeleted,
	}

	// Deep copy sort fields
	if q.Sort != nil {
		clone.Sort = make([]SortField, len(q.Sort))
		copy(clone.Sort, q.Sort)
	}

	// Deep copy filters
	if q.Filters != nil {
		clone.Filters = make([]FilterCriteria, len(q.Filters))
		copy(clone.Filters, q.Filters)
	}

	// Deep copy preloads
	if q.Preloads != nil {
		clone.Preloads = make([]string, len(q.Preloads))
		copy(clone.Preloads, q.Preloads)
	}

	return clone
}
