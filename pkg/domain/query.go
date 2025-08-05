package domain

// SortOrder represents sort direction
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// SortField represents a single sort specification
type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

// SortMap represents sorting configuration
type SortMap[T IBaseModel] map[string]SortOrder

type IQueryParams[T IBaseModel] interface {
	Filter() interface{}

	Sort() interface{}

	Limit() int

	Offset() int

	Preloads() []string

	WithFilter(filter interface{}) IQueryParams[T]

	WithSort(sort interface{}) IQueryParams[T]

	WithLimit(limit int) IQueryParams[T]

	WithOffset(offset int) IQueryParams[T]

	WithPreloads(preloads []string) IQueryParams[T]

	PrepareDefaults() error

	// Additional methods for compatibility with query package
	HasFilters() bool
	ToFilterCriteria() []FilterCriteria
	HasSort() bool
	HasPreloads() bool

	// Soft-delete visibility methods
	GetIncludeDeleted() bool
	GetOnlyDeleted() bool
	WithDeletedVisibility(includeDeleted, onlyDeleted bool) IQueryParams[T]

	// Sort field methods
	ToSortFields() []SortField

	// Methods for building queries
	WithFilters(identifier interface{}) IQueryParams[T]
	AddSortDesc(field string) IQueryParams[T]
	AddSortAsc(field string) IQueryParams[T]
}

// SimpleQueryParams provides a basic implementation of IQueryParams
type SimpleQueryParams[T IBaseModel] struct {
	filter         interface{}
	sort           interface{}
	limit          int
	offset         int
	preloads       []string
	includeDeleted bool
	onlyDeleted    bool
	filters        []FilterCriteria
}

// NewQueryParams creates a new SimpleQueryParams instance
func NewQueryParams[T IBaseModel]() IQueryParams[T] {
	return &SimpleQueryParams[T]{
		limit:    50,
		offset:   0,
		preloads: make([]string, 0),
		filters:  make([]FilterCriteria, 0),
	}
}

// NewSimpleQueryParams creates a new SimpleQueryParams instance (alias for consistency)
func NewSimpleQueryParams[T IBaseModel]() IQueryParams[T] {
	return NewQueryParams[T]()
}

func (q *SimpleQueryParams[T]) Filter() interface{} { return q.filter }
func (q *SimpleQueryParams[T]) Sort() interface{}   { return q.sort }
func (q *SimpleQueryParams[T]) Limit() int          { return q.limit }
func (q *SimpleQueryParams[T]) Offset() int         { return q.offset }
func (q *SimpleQueryParams[T]) Preloads() []string  { return q.preloads }
func (q *SimpleQueryParams[T]) WithFilter(filter interface{}) IQueryParams[T] {
	q.filter = filter
	return q
}
func (q *SimpleQueryParams[T]) WithSort(sort interface{}) IQueryParams[T] { q.sort = sort; return q }
func (q *SimpleQueryParams[T]) WithLimit(limit int) IQueryParams[T]       { q.limit = limit; return q }
func (q *SimpleQueryParams[T]) WithOffset(offset int) IQueryParams[T]     { q.offset = offset; return q }
func (q *SimpleQueryParams[T]) WithPreloads(preloads []string) IQueryParams[T] {
	q.preloads = preloads
	return q
}
func (q *SimpleQueryParams[T]) PrepareDefaults() error             { return nil }
func (q *SimpleQueryParams[T]) HasFilters() bool                   { return len(q.filters) > 0 }
func (q *SimpleQueryParams[T]) ToFilterCriteria() []FilterCriteria { return q.filters }
func (q *SimpleQueryParams[T]) HasSort() bool                      { return q.sort != nil }
func (q *SimpleQueryParams[T]) HasPreloads() bool                  { return len(q.preloads) > 0 }
func (q *SimpleQueryParams[T]) GetIncludeDeleted() bool            { return q.includeDeleted }
func (q *SimpleQueryParams[T]) GetOnlyDeleted() bool               { return q.onlyDeleted }
func (q *SimpleQueryParams[T]) WithDeletedVisibility(includeDeleted, onlyDeleted bool) IQueryParams[T] {
	q.includeDeleted = includeDeleted
	q.onlyDeleted = onlyDeleted
	return q
}
func (q *SimpleQueryParams[T]) ToSortFields() []SortField {
	// Convert sort interface to SortField slice
	if q.sort == nil {
		return []SortField{}
	}

	// Handle SortMap type
	if sortMap, ok := q.sort.(SortMap[T]); ok {
		var fields []SortField
		for field, order := range sortMap {
			fields = append(fields, SortField{Field: field, Order: order})
		}
		return fields
	}

	// Handle []SortField type
	if sortFields, ok := q.sort.([]SortField); ok {
		return sortFields
	}

	return []SortField{}
}

// WithFilters applies filter criteria to the QueryParams
func (q *SimpleQueryParams[T]) WithFilters(identifier interface{}) IQueryParams[T] {
	// Convert identifier to filter criteria if possible
	if id, ok := identifier.(interface{ ToFilterCriteria() []FilterCriteria }); ok {
		q.filters = id.ToFilterCriteria()
	}
	return q
}

// AddSortDesc adds a descending sort field
func (q *SimpleQueryParams[T]) AddSortDesc(field string) IQueryParams[T] {
	return q.AddSort(field, SortOrderDesc)
}

// AddSortAsc adds an ascending sort field
func (q *SimpleQueryParams[T]) AddSortAsc(field string) IQueryParams[T] {
	return q.AddSort(field, SortOrderAsc)
}

// AddSort adds a sort field with the specified order
func (q *SimpleQueryParams[T]) AddSort(field string, order SortOrder) IQueryParams[T] {
	// Convert current sort to []SortField if needed
	currentSorts := q.ToSortFields()
	currentSorts = append(currentSorts, SortField{Field: field, Order: order})
	q.sort = currentSorts
	return q
}

// FilterCriteria represents a filter condition - moved from identifier package for domain independence
type FilterCriteria struct {
	Field     string           `json:"field"`
	Operator  FilterOperator   `json:"operator"`
	Value     interface{}      `json:"value,omitempty"`
	Values    []interface{}    `json:"values,omitempty"`
	Group     []FilterCriteria `json:"group,omitempty"`
	LogicalOp LogicalOperator  `json:"logical_op,omitempty"`
}

// FilterOperator represents comparison operators
type FilterOperator string

const (
	FilterOperatorEqual        FilterOperator = "="
	FilterOperatorNotEqual     FilterOperator = "!="
	FilterOperatorGreaterThan  FilterOperator = ">"
	FilterOperatorGreaterEqual FilterOperator = ">="
	FilterOperatorLessThan     FilterOperator = "<"
	FilterOperatorLessEqual    FilterOperator = "<="
	FilterOperatorLike         FilterOperator = "LIKE"
	FilterOperatorIn           FilterOperator = "IN"
	FilterOperatorNotIn        FilterOperator = "NOT IN"
	FilterOperatorIsNull       FilterOperator = "IS NULL"
	FilterOperatorIsNotNull    FilterOperator = "IS NOT NULL"
	FilterOperatorBetween      FilterOperator = "BETWEEN"
	FilterOperatorContains     FilterOperator = "CONTAINS"
	FilterOperatorHas          FilterOperator = "HAS"
)

// LogicalOperator represents logical operators for combining filters
type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "AND"
	LogicalOperatorOr  LogicalOperator = "OR"
)

// RecursiveFilter represents a hierarchical filter structure for complex queries
type RecursiveFilter struct {
	Field     string            `json:"field,omitempty"`
	Operator  FilterOperator    `json:"operator,omitempty"`
	Value     interface{}       `json:"value,omitempty"`
	Values    []interface{}     `json:"values,omitempty"`
	Children  []RecursiveFilter `json:"children,omitempty"`
	LogicalOp LogicalOperator   `json:"logical_op,omitempty"`
}

// RecursiveFilterParser provides parsing capabilities for recursive filters
type RecursiveFilterParser struct{}

// NewRecursiveFilterParser creates a new recursive filter parser
func NewRecursiveFilterParser() *RecursiveFilterParser {
	return &RecursiveFilterParser{}
}

// Parse converts a RecursiveFilter to FilterCriteria
func (rfp *RecursiveFilterParser) Parse(filter RecursiveFilter) []FilterCriteria {
	var criteria []FilterCriteria

	// Convert the recursive filter to filter criteria
	if filter.Field != "" {
		criterion := FilterCriteria{
			Field:     filter.Field,
			Operator:  filter.Operator,
			Value:     filter.Value,
			Values:    filter.Values,
			LogicalOp: filter.LogicalOp,
		}
		criteria = append(criteria, criterion)
	}

	// Process children
	for _, child := range filter.Children {
		childCriteria := rfp.Parse(child)
		criteria = append(criteria, childCriteria...)
	}

	return criteria
}
